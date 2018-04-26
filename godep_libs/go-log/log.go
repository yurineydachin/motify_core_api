package log

import (
	"context"
	"fmt"
	"io"
	"strings"
	"sync"
	"sync/atomic"

	"godep.lzd.co/go-log/format"
	"godep.lzd.co/go-log/internal"
)

type Severity int

const NO_STACK_TRACE_INFO = format.NO_STACK_TRACE_INFO

// Don't modify these constants! It's Severity level from The Syslog Protocol
// https://tools.ietf.org/html/rfc5424#section-6.2.1
const (
	EMERGENCY Severity = iota
	ALERT
	CRITICAL
	ERROR
	WARNING
	NOTICE
	INFO
	DEBUG
)

func (l Severity) Code() int {
	return int(l)
}

func (l Severity) String() string {
	switch l {
	case DEBUG:
		return "DEBUG"
	case INFO:
		return "INFO"
	case NOTICE:
		return "NOTICE"
	case WARNING:
		return "WARNING"
	case ERROR:
		return "ERROR"
	case CRITICAL:
		return "CRITICAL"
	case ALERT:
		return "ALERT"
	case EMERGENCY:
		return "EMERGENCY"
	default:
		return "UNKNOWN"
	}
}

func ParseSeverity(severity string) (Severity, bool) {
	switch severity {
	case "DEBUG", "debug":
		return DEBUG, true
	case "INFO", "info":
		return INFO, true
	case "NOTICE", "notice":
		return NOTICE, true
	case "WARNING", "warning":
		return WARNING, true
	case "ERROR", "error":
		return ERROR, true
	case "CRITICAL", "critical":
		return CRITICAL, true
	case "ALERT", "alert":
		return ALERT, true
	case "EMERGENCY", "emergency":
		return EMERGENCY, true
	default:
		if s := strings.ToUpper(severity); s != severity {
			return ParseSeverity(s)
		}
		return Severity(-1), false
	}
}

const facility = 18 // local use 2 (local2) from rfc5424
const maxBufferCapacityToReuse uint64 = 16 * 1024
const delimiter = " | "
const none = "-"

/* all writers */
var writers = make(map[key]*bufferLayer)
var writersMtx sync.Mutex

type Logger struct {
	writer writerFromCloser
	key    key
	err    sync.Mutex
	syslog bool
	opts   options
}

type key struct {
	service string
	network string
	address string
}

type writerFromCloser interface {
	WriteFrom(to io.WriterTo) error
	Close() error
}

func NewLogger(service, network, address string, level Severity) *Logger {
	return New(WithServiceName(service), WithDialOptions(network, address), WithLevel(level))
}

func New(opts ...option) *Logger {
	l := &Logger{opts: defaultOptions}
	for _, opt := range opts {
		opt(&l.opts)
	}
	writersMtx.Lock()
	k := key{service: l.opts.service, network: l.opts.network, address: l.opts.address}
	w, exists := writers[k]
	if !exists {
		w = newBufferLayer(newWriter(l.opts.network, l.opts.address), l.opts.errorWriterEnabled, l.opts.bufferSize, l.opts.workerCount)
		if l.opts.metrics {
			w.metrics = &internal.GlobalLoggerMetrics
		}
		writers[k] = w
	}
	writersMtx.Unlock()
	l.writer = w
	l.key = k

	l.syslog = l.opts.network != "" && l.opts.address != ""
	l.err = sync.Mutex{}

	return l
}

func WithServiceName(name string) option {
	return func(opts *options) {
		opts.service = name
	}
}

func WithDialOptions(network, address string) option {
	return func(opts *options) {
		opts.network = network
		opts.address = address
	}
}

func WithLevel(level Severity) option {
	return func(opts *options) {
		atomic.StoreInt64(&opts.level, int64(level))
	}
}

func WithErrorWriter(enabled bool) option {
	return func(opts *options) {
		opts.errorWriterEnabled = enabled
	}
}

func WithBufferSize(b int64) option {
	return func(opts *options) {
		opts.bufferSize = b
	}
}

func WithWorkerCount(c int) option {
	return func(opts *options) {
		opts.workerCount = c
	}
}

func WithMetrics(b bool) option {
	return func(opts *options) {
		opts.metrics = b
	}
}

func WithBacktraceSkips(packages []string) option {
	return func(opts *options) {
		opts.backtraceSkips = packages
	}
}

func (l *Logger) Level() Severity {
	return Severity(atomic.LoadInt64(&l.opts.level))
}

func (l *Logger) SetLevel(lvl Severity) {
	atomic.StoreInt64(&l.opts.level, int64(lvl))
}

func (l *Logger) Close() {
	writersMtx.Lock()
	delete(writers, l.key)
	writersMtx.Unlock()
	l.writer.Close()
}

func (l *Logger) Record(level Severity, f io.WriterTo) error {
	if level > Severity(atomic.LoadInt64(&l.opts.level)) {
		return nil
	}

	if s, ok := f.(levelSetter); ok {
		s.SetLevel(level)
	}
	if s, ok := f.(serviceSetter); ok {
		s.SetService(l.key.service)
	}
	if s, ok := f.(syslogHeaderEnabler); ok {
		s.EnableSyslogHeader(l.syslog)
	}
	if s, ok := f.(callStackSkipInc); ok {
		s.IncExtCallStackSkip(2)
	}

	return l.writer.WriteFrom(f)
}

func (l *Logger) Stats() (sent, lost, totalLoggedCount, totalLoggedLength, createdBuffersCount, skippedBuf uint64) {
	if w, ok := l.writer.(*bufferLayer); ok {
		sent = atomic.LoadUint64(&w.sent)
		lost = atomic.LoadUint64(&w.lost)
		totalLoggedCount = atomic.LoadUint64(&w.totalLoggedCount)
		totalLoggedLength = atomic.LoadUint64(&w.totalLoggedLength)
		createdBuffersCount = atomic.LoadUint64(&w.createdBuffersCount)
		skippedBuf = atomic.LoadUint64(&w.skippedBuf)
	}

	return
}

type flusher interface {
	Flush(context.Context) error
}

// Flush blocks until Logger writes current buffer
func (l *Logger) Flush(ctx context.Context) error {
	if f, ok := l.writer.(flusher); ok {
		return f.Flush(ctx)
	}
	return fmt.Errorf("Writer is not flusher (Flush(context.Context) error)")
}

type options struct {
	service            string
	network            string
	address            string
	level              int64
	errorWriterEnabled bool
	bufferSize         int64
	workerCount        int
	metrics            bool
	backtraceSkips     []string
}

var defaultOptions = options{
	level:              int64(DEBUG),
	errorWriterEnabled: true,
	bufferSize:         32 * 1024 * 1024,
	workerCount:        4,
	metrics:            true,
}

type option func(*options)

type levelSetter interface {
	SetLevel(format.Severity)
}

type serviceSetter interface {
	SetService(string)
}

type syslogHeaderEnabler interface {
	EnableSyslogHeader(bool)
}

type callStackSkipInc interface {
	IncExtCallStackSkip(int)
}
