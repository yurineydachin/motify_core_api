package log

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"godep.lzd.co/go-log/format"
	"godep.lzd.co/go-log/internal"
)

var interceptorLogger, discardLogger, traceLogger *Logger

type interceptorWriter struct {
	buf bytes.Buffer
}

func (w *interceptorWriter) WriteFrom(in io.WriterTo) error {
	w.buf.Reset()
	in.WriteTo(&w.buf)
	return nil
}

func (w *interceptorWriter) Close() error {
	return nil
}

type discarder struct {
	writerFromCloser
}

func (c discarder) WriteFrom(in io.WriterTo) error {
	if c.writerFromCloser != nil {
		return c.writerFromCloser.WriteFrom(in)
	}
	_, err := in.WriteTo(ioutil.Discard)
	return err
}

func (discarder) Close() error {
	return nil
}

const service = "service"

func init() {
	interceptorLogger = NewLogger(service, "", "", DEBUG)
	interceptorLogger.writer = &interceptorWriter{}
	discardLogger = NewLogger(service, "", "", DEBUG)
	discardLogger.writer = &discarder{}

	traceLogger = New(
		WithServiceName(service),
		WithLevel(DEBUG),
		WithBacktraceSkips([]string{"testing"}),
	)
	traceLogger.writer = &interceptorWriter{}
}

func resetWriters() {
	writersMtx.Lock()
	writers = make(map[key]*bufferLayer)
	writersMtx.Unlock()
}

func TestCallStackSkip(t *testing.T) {
	buf := new(bytes.Buffer)
	(&format.Std{}).WriteTo(buf)
	source := internal.CurrentLineMinusOne()
	parts := strings.Split(buf.String(), delimiter)
	if len(parts) != 13 {
		t.Errorf("WriteTo: expected 13 parts of log string, got: %d\n", len(parts))
	}
	if !strings.HasSuffix(parts[8], "go-log") {
		t.Errorf("WriteTo: expected component_name suffix \"go-log\", got: %s\n", parts[8])
	}
	if parts[9] != source {
		t.Errorf("WriteTo: expected file name and line \"%s\", got: \"%s\"\n", source, parts[9])
	}

	interceptorLogger.Record(DEBUG, &format.Std{})
	source = internal.CurrentLineMinusOne()
	if w, ok := interceptorLogger.writer.(*interceptorWriter); ok {
		parts := strings.Split(string(w.buf.Bytes()), delimiter)
		if len(parts) != 13 {
			t.Errorf("Record: expected 13 parts of log string, got: %d\n", len(parts))
		}
		if !strings.HasSuffix(parts[8], "go-log") {
			t.Errorf("Record: expected component_name suffix \"go-log\", got: %s\n", parts[8])
		}
		if parts[9] != source {
			t.Errorf("Record: expected file name and line \"%s\", got: \"%s\"\n", source, parts[9])
		}
	} else {
		t.Fail()
	}

	interceptorLogger.Log(0, Span{}, DEBUG, "", nil)
	source = internal.CurrentLineMinusOne()
	if w, ok := interceptorLogger.writer.(*interceptorWriter); ok {
		parts := strings.Split(string(w.buf.Bytes()), delimiter)
		if len(parts) != 13 {
			t.Errorf("Log: expected 13 parts of log string, got: %d\n", len(parts))
		}
		if !strings.HasSuffix(parts[8], "go-log") {
			t.Errorf("Log: expected component_name suffix \"go-log\", got: %s\n", parts[8])
		}
		if parts[9] != source {
			t.Errorf("Log: expected file name and line \"%s\", got: \"%s\"\n", source, parts[9])
		}
	} else {
		t.Fail()
	}

	interceptorLogger.Logf(0, Span{}, DEBUG, "", nil)
	source = internal.CurrentLineMinusOne()
	if w, ok := interceptorLogger.writer.(*interceptorWriter); ok {
		parts := strings.Split(string(w.buf.Bytes()), delimiter)
		if len(parts) != 13 {
			t.Errorf("Logf: expected 13 parts of log string, got: %d\n", len(parts))
		}
		if !strings.HasSuffix(parts[8], "go-log") {
			t.Errorf("Logf: expected component_name suffix \"go-log\", got: %s\n", parts[8])
		}
		if parts[9] != source {
			t.Errorf("Logf: expected file name and line \"%s\", got: \"%s\"\n", source, parts[9])
		}
	} else {
		t.Fail()
	}

	interceptorLogger.Log(1<<15, Span{}, DEBUG, "", nil)
	if w, ok := interceptorLogger.writer.(*interceptorWriter); ok {
		parts := strings.Split(string(w.buf.Bytes()), delimiter)
		if len(parts) != 13 {
			t.Errorf("Logf: expected 13 parts of log string, got: %d\n", len(parts))
		}
		if parts[8] != none {
			t.Errorf("Logf: expected component_name \"%s\", got: %s\n", none, parts[8])
		}
		if parts[9] != ":0" {
			t.Errorf("Logf: expected file name and line \":0\", got: \"%s\"\n", parts[9])
		}
	} else {
		t.Fail()
	}

	buf = new(bytes.Buffer)
	(&format.Std{CallStackSkip: NO_STACK_TRACE_INFO}).WriteTo(buf)
	parts = strings.Split(buf.String(), delimiter)
	if len(parts) != 13 {
		t.Errorf("WriteTo: expected 13 parts of log string, got: %d\n", len(parts))
	}
	if parts[8] != none {
		t.Errorf("WriteTo: expected component_name \"%s\", got: %s\n", none, parts[8])
	}
	if parts[9] != none {
		t.Errorf("WriteTo: expected file name and line \"-\", got: \"%s\"\n", parts[9])
	}

	interceptorLogger.Record(DEBUG, &format.Std{CallStackSkip: NO_STACK_TRACE_INFO})
	if w, ok := interceptorLogger.writer.(*interceptorWriter); ok {
		parts := strings.Split(string(w.buf.Bytes()), delimiter)
		if len(parts) != 13 {
			t.Errorf("Record: expected 13 parts of log string, got: %d\n", len(parts))
		}
		if parts[8] != none {
			t.Errorf("Record: expected component_name \"%s\", got: %s\n", none, parts[8])
		}
		if parts[9] != none {
			t.Errorf("Record: expected file name and line \"-\", got: \"%s\"\n", parts[9])
		}
	} else {
		t.Fail()
	}

	interceptorLogger.Log(NO_STACK_TRACE_INFO, Span{}, DEBUG, "", nil)
	if w, ok := interceptorLogger.writer.(*interceptorWriter); ok {
		parts := strings.Split(string(w.buf.Bytes()), delimiter)
		if len(parts) != 13 {
			t.Errorf("Log: expected 13 parts of log string, got: %d\n", len(parts))
		}
		if parts[8] != none {
			t.Errorf("Log: expected component_name \"%s\", got: %s\n", none, parts[8])
		}
		if parts[9] != none {
			t.Errorf("Log: expected file name and line \"-\", got: \"%s\"\n", parts[9])
		}
	} else {
		t.Fail()
	}

	interceptorLogger.Logf(NO_STACK_TRACE_INFO, Span{}, DEBUG, "", nil)
	if w, ok := interceptorLogger.writer.(*interceptorWriter); ok {
		parts := strings.Split(string(w.buf.Bytes()), delimiter)
		if len(parts) != 13 {
			t.Errorf("Logf: expected 13 parts of log string, got: %d\n", len(parts))
		}
		if parts[8] != none {
			t.Errorf("Logf: expected component_name \"%s\", got: %s\n", none, parts[8])
		}
		if parts[9] != none {
			t.Errorf("Logf: expected file name and line \"-\", got: \"%s\"\n", parts[9])
		}
	} else {
		t.Fail()
	}

	traceLogger.Logf(1, Span{}, DEBUG, "", nil)
	if w, ok := traceLogger.writer.(*interceptorWriter); ok {
		parts := strings.Split(string(w.buf.Bytes()), delimiter)
		if len(parts) != 13 {
			t.Errorf("Logf: expected 13 parts of log string, got: %d\n", len(parts))
		}
		if parts[8] != "runtime" {
			t.Errorf("Logf: expected package runtime, got: %s\n", parts[8])
		}
	} else {
		t.Fail()
	}
}

func TestParseSeverity(t *testing.T) {
	data := map[Severity][]string{
		DEBUG:        {"debug", "Debug", "DEBUG"},
		INFO:         {"info", "Info", "INFO"},
		NOTICE:       {"notice", "Notice", "NOTICE"},
		WARNING:      {"warning", "Warning", "WARNING"},
		ERROR:        {"error", "Error", "ERROR"},
		CRITICAL:     {"critical", "Critical", "CRITICAL"},
		ALERT:        {"alert", "Alert", "ALERT"},
		EMERGENCY:    {"emergency", "Emergency", "EMERGENCY"},
		Severity(-1): {"Unknown level string that should not be parsed"},
	}

	for severity, sourceStrings := range data {
		for _, str := range sourceStrings {
			parsedSeverity, ok := ParseSeverity(str)
			switch parsedSeverity {
			case DEBUG, INFO, NOTICE, WARNING, ERROR, CRITICAL, ALERT, EMERGENCY:
				if severity != parsedSeverity {
					t.Errorf("Severity parsing error. Expected: %d, got: %d."+
						" Source string before parsing: %s", severity, parsedSeverity, str)
				}
				if !ok {
					t.Errorf("Severity parsing error. Expected %t, got %t", true, ok)
				}
				break
			default:
				if ok {
					t.Errorf("Severity parsing error. Expected %t, got %t", false, ok)
				}
			}
		}
	}
}

func TestLevel(t *testing.T) {
	level := interceptorLogger.Level()
	w, ok := interceptorLogger.writer.(*interceptorWriter)
	if !ok {
		t.Error("expected bufferLayer is (*interceptorWriter)")
	}
	w.buf.Reset()

	interceptorLogger.SetLevel(EMERGENCY)
	interceptorLogger.Logf(0, Span{}, DEBUG, "test", nil)
	interceptorLogger.Logf(0, Span{}, INFO, "test", nil)
	interceptorLogger.Logf(0, Span{}, NOTICE, "test", nil)
	interceptorLogger.Logf(0, Span{}, WARNING, "test", nil)
	interceptorLogger.Logf(0, Span{}, ERROR, "test", nil)
	interceptorLogger.Logf(0, Span{}, CRITICAL, "test", nil)
	interceptorLogger.Logf(0, Span{}, ALERT, "test", nil)
	if w.buf.Len() != 0 {
		t.Errorf("expected number of written bytes is: 0, got: %d\n", w.buf.Len())
	}
	interceptorLogger.Logf(0, Span{}, EMERGENCY, "test", nil)
	if w.buf.Len() == 0 {
		t.Errorf("expected number of written bytes > 0, got: %d\n", w.buf.Len())
	}
	w.buf.Reset()

	interceptorLogger.SetLevel(ALERT)
	interceptorLogger.Logf(0, Span{}, DEBUG, "test", nil)
	interceptorLogger.Logf(0, Span{}, INFO, "test", nil)
	interceptorLogger.Logf(0, Span{}, NOTICE, "test", nil)
	interceptorLogger.Logf(0, Span{}, WARNING, "test", nil)
	interceptorLogger.Logf(0, Span{}, ERROR, "test", nil)
	interceptorLogger.Logf(0, Span{}, CRITICAL, "test", nil)
	if w.buf.Len() != 0 {
		t.Errorf("expected number of written bytes is: 0, got: %d\n", w.buf.Len())
	}
	interceptorLogger.Logf(0, Span{}, ALERT, "test", nil)
	if w.buf.Len() == 0 {
		t.Errorf("expected number of written bytes > 0, got: %d\n", w.buf.Len())
	}
	w.buf.Reset()
	interceptorLogger.Logf(0, Span{}, EMERGENCY, "test", nil)
	if w.buf.Len() == 0 {
		t.Errorf("expected number of written bytes > 0, got: %d\n", w.buf.Len())
	}
	w.buf.Reset()

	interceptorLogger.SetLevel(CRITICAL)
	interceptorLogger.Logf(0, Span{}, DEBUG, "test", nil)
	interceptorLogger.Logf(0, Span{}, INFO, "test", nil)
	interceptorLogger.Logf(0, Span{}, NOTICE, "test", nil)
	interceptorLogger.Logf(0, Span{}, WARNING, "test", nil)
	interceptorLogger.Logf(0, Span{}, ERROR, "test", nil)
	if w.buf.Len() != 0 {
		t.Errorf("expected number of written bytes is: 0, got: %d\n", w.buf.Len())
	}
	interceptorLogger.Logf(0, Span{}, CRITICAL, "test", nil)
	if w.buf.Len() == 0 {
		t.Errorf("expected number of written bytes > 0, got: %d\n", w.buf.Len())
	}
	w.buf.Reset()
	interceptorLogger.Logf(0, Span{}, ALERT, "test", nil)
	if w.buf.Len() == 0 {
		t.Errorf("expected number of written bytes > 0, got: %d\n", w.buf.Len())
	}
	w.buf.Reset()
	interceptorLogger.Logf(0, Span{}, EMERGENCY, "test", nil)
	if w.buf.Len() == 0 {
		t.Errorf("expected number of written bytes > 0, got: %d\n", w.buf.Len())
	}
	w.buf.Reset()

	interceptorLogger.SetLevel(ERROR)
	interceptorLogger.Logf(0, Span{}, DEBUG, "test", nil)
	interceptorLogger.Logf(0, Span{}, INFO, "test", nil)
	interceptorLogger.Logf(0, Span{}, NOTICE, "test", nil)
	interceptorLogger.Logf(0, Span{}, WARNING, "test", nil)
	if w.buf.Len() != 0 {
		t.Errorf("expected number of written bytes is: 0, got: %d\n", w.buf.Len())
	}
	interceptorLogger.Logf(0, Span{}, ERROR, "test", nil)
	if w.buf.Len() == 0 {
		t.Errorf("expected number of written bytes > 0, got: %d\n", w.buf.Len())
	}
	w.buf.Reset()
	interceptorLogger.Logf(0, Span{}, CRITICAL, "test", nil)
	if w.buf.Len() == 0 {
		t.Errorf("expected number of written bytes > 0, got: %d\n", w.buf.Len())
	}
	w.buf.Reset()
	interceptorLogger.Logf(0, Span{}, ALERT, "test", nil)
	if w.buf.Len() == 0 {
		t.Errorf("expected number of written bytes > 0, got: %d\n", w.buf.Len())
	}
	w.buf.Reset()
	interceptorLogger.Logf(0, Span{}, EMERGENCY, "test", nil)
	if w.buf.Len() == 0 {
		t.Errorf("expected number of written bytes > 0, got: %d\n", w.buf.Len())
	}
	w.buf.Reset()

	interceptorLogger.SetLevel(WARNING)
	interceptorLogger.Logf(0, Span{}, DEBUG, "test", nil)
	interceptorLogger.Logf(0, Span{}, INFO, "test", nil)
	interceptorLogger.Logf(0, Span{}, NOTICE, "test", nil)
	if w.buf.Len() != 0 {
		t.Errorf("expected number of written bytes is: 0, got: %d\n", w.buf.Len())
	}
	interceptorLogger.Logf(0, Span{}, WARNING, "test", nil)
	if w.buf.Len() == 0 {
		t.Errorf("expected number of written bytes > 0, got: %d\n", w.buf.Len())
	}
	w.buf.Reset()
	interceptorLogger.Logf(0, Span{}, ERROR, "test", nil)
	if w.buf.Len() == 0 {
		t.Errorf("expected number of written bytes > 0, got: %d\n", w.buf.Len())
	}
	w.buf.Reset()
	interceptorLogger.Logf(0, Span{}, CRITICAL, "test", nil)
	if w.buf.Len() == 0 {
		t.Errorf("expected number of written bytes > 0, got: %d\n", w.buf.Len())
	}
	w.buf.Reset()
	interceptorLogger.Logf(0, Span{}, ALERT, "test", nil)
	if w.buf.Len() == 0 {
		t.Errorf("expected number of written bytes > 0, got: %d\n", w.buf.Len())
	}
	w.buf.Reset()
	interceptorLogger.Logf(0, Span{}, EMERGENCY, "test", nil)
	if w.buf.Len() == 0 {
		t.Errorf("expected number of written bytes > 0, got: %d\n", w.buf.Len())
	}
	w.buf.Reset()

	interceptorLogger.SetLevel(NOTICE)
	interceptorLogger.Logf(0, Span{}, DEBUG, "test", nil)
	interceptorLogger.Logf(0, Span{}, INFO, "test", nil)
	if w.buf.Len() != 0 {
		t.Errorf("expected number of written bytes is: 0, got: %d\n", w.buf.Len())
	}
	interceptorLogger.Logf(0, Span{}, NOTICE, "test", nil)
	if w.buf.Len() == 0 {
		t.Errorf("expected number of written bytes > 0, got: %d\n", w.buf.Len())
	}
	w.buf.Reset()
	interceptorLogger.Logf(0, Span{}, WARNING, "test", nil)
	if w.buf.Len() == 0 {
		t.Errorf("expected number of written bytes > 0, got: %d\n", w.buf.Len())
	}
	w.buf.Reset()
	interceptorLogger.Logf(0, Span{}, ERROR, "test", nil)
	if w.buf.Len() == 0 {
		t.Errorf("expected number of written bytes > 0, got: %d\n", w.buf.Len())
	}
	w.buf.Reset()
	interceptorLogger.Logf(0, Span{}, CRITICAL, "test", nil)
	if w.buf.Len() == 0 {
		t.Errorf("expected number of written bytes > 0, got: %d\n", w.buf.Len())
	}
	w.buf.Reset()
	interceptorLogger.Logf(0, Span{}, ALERT, "test", nil)
	if w.buf.Len() == 0 {
		t.Errorf("expected number of written bytes > 0, got: %d\n", w.buf.Len())
	}
	w.buf.Reset()
	interceptorLogger.Logf(0, Span{}, EMERGENCY, "test", nil)
	if w.buf.Len() == 0 {
		t.Errorf("expected number of written bytes > 0, got: %d\n", w.buf.Len())
	}
	w.buf.Reset()

	interceptorLogger.SetLevel(INFO)
	interceptorLogger.Logf(0, Span{}, DEBUG, "test", nil)
	if w.buf.Len() != 0 {
		t.Errorf("expected number of written bytes is: 0, got: %d\n", w.buf.Len())
	}
	interceptorLogger.Logf(0, Span{}, INFO, "test", nil)
	if w.buf.Len() == 0 {
		t.Errorf("expected number of written bytes > 0, got: %d\n", w.buf.Len())
	}
	w.buf.Reset()
	interceptorLogger.Logf(0, Span{}, NOTICE, "test", nil)
	if w.buf.Len() == 0 {
		t.Errorf("expected number of written bytes > 0, got: %d\n", w.buf.Len())
	}
	w.buf.Reset()
	interceptorLogger.Logf(0, Span{}, WARNING, "test", nil)
	if w.buf.Len() == 0 {
		t.Errorf("expected number of written bytes > 0, got: %d\n", w.buf.Len())
	}
	w.buf.Reset()
	interceptorLogger.Logf(0, Span{}, ERROR, "test", nil)
	if w.buf.Len() == 0 {
		t.Errorf("expected number of written bytes > 0, got: %d\n", w.buf.Len())
	}
	w.buf.Reset()
	interceptorLogger.Logf(0, Span{}, CRITICAL, "test", nil)
	if w.buf.Len() == 0 {
		t.Errorf("expected number of written bytes > 0, got: %d\n", w.buf.Len())
	}
	w.buf.Reset()
	interceptorLogger.Logf(0, Span{}, ALERT, "test", nil)
	if w.buf.Len() == 0 {
		t.Errorf("expected number of written bytes > 0, got: %d\n", w.buf.Len())
	}
	w.buf.Reset()
	interceptorLogger.Logf(0, Span{}, EMERGENCY, "test", nil)
	if w.buf.Len() == 0 {
		t.Errorf("expected number of written bytes > 0, got: %d\n", w.buf.Len())
	}
	w.buf.Reset()

	interceptorLogger.SetLevel(DEBUG)
	interceptorLogger.Logf(0, Span{}, DEBUG, "test", nil)
	if w.buf.Len() == 0 {
		t.Errorf("expected number of written bytes > 0, got: %d\n", w.buf.Len())
	}
	w.buf.Reset()
	interceptorLogger.Logf(0, Span{}, INFO, "test", nil)
	if w.buf.Len() == 0 {
		t.Errorf("expected number of written bytes > 0, got: %d\n", w.buf.Len())
	}
	w.buf.Reset()
	interceptorLogger.Logf(0, Span{}, NOTICE, "test", nil)
	if w.buf.Len() == 0 {
		t.Errorf("expected number of written bytes > 0, got: %d\n", w.buf.Len())
	}
	w.buf.Reset()
	interceptorLogger.Logf(0, Span{}, WARNING, "test", nil)
	if w.buf.Len() == 0 {
		t.Errorf("expected number of written bytes > 0, got: %d\n", w.buf.Len())
	}
	w.buf.Reset()
	interceptorLogger.Logf(0, Span{}, ERROR, "test", nil)
	if w.buf.Len() == 0 {
		t.Errorf("expected number of written bytes > 0, got: %d\n", w.buf.Len())
	}
	w.buf.Reset()
	interceptorLogger.Logf(0, Span{}, CRITICAL, "test", nil)
	if w.buf.Len() == 0 {
		t.Errorf("expected number of written bytes > 0, got: %d\n", w.buf.Len())
	}
	w.buf.Reset()
	interceptorLogger.Logf(0, Span{}, ALERT, "test", nil)
	if w.buf.Len() == 0 {
		t.Errorf("expected number of written bytes > 0, got: %d\n", w.buf.Len())
	}
	w.buf.Reset()
	interceptorLogger.Logf(0, Span{}, EMERGENCY, "test", nil)
	if w.buf.Len() == 0 {
		t.Errorf("expected number of written bytes > 0, got: %d\n", w.buf.Len())
	}
	w.buf.Reset()

	interceptorLogger.SetLevel(level)
}

func log(l *Logger, callStackSkip int, span Span, level Severity, message string, data map[string]interface{}) error {
	buf := new(bytes.Buffer)
	f := &format.Std{
		CallStackSkip: callStackSkip,
		TraceId:       span.TraceID,
		ParentSpanId:  span.ParentSpanID,
		SpanId:        span.SpanID,
		RolloutType:   span.RolloutType,
		Message:       message,
		Data:          data,
	}
	f.SetLevel(level)
	f.WriteTo(buf)
	if err := l.writer.WriteFrom(buf); err != nil {
		return err
	}
	return nil
}

var pool = sync.Pool{
	New: func() interface{} {
		return new(bytes.Buffer)
	},
}

func syncPoolLog(l *Logger, callStackSkip int, span Span, level Severity, message string, data map[string]interface{}) error {
	buf := pool.Get().(*bytes.Buffer)
	f := &format.Std{
		CallStackSkip: callStackSkip,
		TraceId:       span.TraceID,
		ParentSpanId:  span.ParentSpanID,
		SpanId:        span.SpanID,
		RolloutType:   span.RolloutType,
		Message:       message,
		Data:          data,
	}
	f.SetLevel(level)
	f.WriteTo(buf)
	if err := l.writer.WriteFrom(buf); err != nil {
		return err
	}
	if buf.Cap() <= 10*1024 {
		pool.Put(buf)
	}
	return nil
}

var freeList = make(chan *bytes.Buffer, 10000)

func leakyBufferLog(l *Logger, callStackSkip int, span Span, level Severity,
	message string, data map[string]interface{}) error {

	var buf *bytes.Buffer
	select {
	case buf = <-freeList:
	default:
		buf = new(bytes.Buffer)
	}
	f := &format.Std{
		CallStackSkip: callStackSkip,
		TraceId:       span.TraceID,
		ParentSpanId:  span.ParentSpanID,
		SpanId:        span.SpanID,
		RolloutType:   span.RolloutType,
		Message:       message,
		Data:          data,
	}
	f.SetLevel(level)
	f.WriteTo(buf)
	if err := l.writer.WriteFrom(buf); err != nil {
		return err
	}
	select {
	case freeList <- buf:
	default:
	}
	return nil
}

func BenchmarkLogWithoutNet(b *testing.B) {
	for i := 0; i < b.N; i++ {
		if err := log(discardLogger, 0, Span{"3ymrswshj4sg", "", "6yaoivssj1rt", "10"}, DEBUG, "test message",
			map[string]interface{}{"attribute": "value"}); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkSyncPoolLogWithoutNet(b *testing.B) {
	for i := 0; i < b.N; i++ {
		if err := syncPoolLog(discardLogger, 0, Span{"3ymrswshj4sg", "", "6yaoivssj1rt", "10"}, DEBUG, "test message",
			map[string]interface{}{"attribute": "value"}); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkLeakyBufferLogWithoutNet(b *testing.B) {
	for i := 0; i < b.N; i++ {
		if err := leakyBufferLog(discardLogger, 0, Span{"3ymrswshj4sg", "", "6yaoivssj1rt", "10"}, DEBUG, "test message",
			map[string]interface{}{"attribute": "value"}); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkLogWithoutNetParallel(b *testing.B) {
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			if err := log(discardLogger, 0, Span{"3ymrswshj4sg", "", "6yaoivssj1rt", "10"}, DEBUG,
				"test message", map[string]interface{}{"attribute": "value"}); err != nil {
				b.Fatal(err)
			}
		}
	})
}

func BenchmarkSyncPoolLogWithoutNetParallel(b *testing.B) {
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			if err := syncPoolLog(discardLogger, 0, Span{"3ymrswshj4sg", "", "6yaoivssj1rt", "10"}, DEBUG,
				"test message", map[string]interface{}{"attribute": "value"}); err != nil {
				b.Fatal(err)
			}
		}
	})
}

func BenchmarkLeakyBufferLogWithoutNetParallel(b *testing.B) {
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			if err := leakyBufferLog(discardLogger, 0, Span{"3ymrswshj4sg", "", "6yaoivssj1rt", "10"}, DEBUG,
				"test message", map[string]interface{}{"attribute": "value"}); err != nil {
				b.Fatal(err)
			}
		}
	})
}

func BenchmarkSingleInstanceLogParallel(b *testing.B) {
	l := NewLogger(service, "", "", DEBUG)
	l.writer = discarder{}
	var i int64
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			atomic.AddInt64(&i, 1)
			l.Log(0, Span{}, DEBUG, strings.Repeat("test message", int(i)%100), nil)
		}
	})
}

func BenchmarkManyInstanceLogParallel(b *testing.B) {
	var i int64
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			l := NewLogger(service, "", "", DEBUG)
			l.writer = discarder{}
			atomic.AddInt64(&i, 1)
			l.Log(0, Span{}, DEBUG, strings.Repeat("test message", int(i)%100), nil)
		}
	})
}

func BenchmarkSingleInstanceLog(b *testing.B) {
	l := NewLogger(service, "", "", DEBUG)
	l.writer = discarder{}
	for i := 0; i < b.N; i++ {
		l.Log(0, Span{}, DEBUG, strings.Repeat("test message", i%100), nil)
	}
}

func BenchmarkManyInstanceLog(b *testing.B) {
	for i := 0; i < b.N; i++ {
		l := NewLogger(service, "", "", DEBUG)
		l.writer = discarder{}
		l.Log(0, Span{}, DEBUG, strings.Repeat("test message", i%100), nil)
	}
}

func TestWriteSettingServiceLevelSyslogFlag(t *testing.T) {
	l := NewLogger("test-api-1", "udp", ":5140", DEBUG)
	l.writer = &interceptorWriter{}

	l.Record(DEBUG, &format.Std{})
	if w, ok := l.writer.(*interceptorWriter); ok {
		pri := fmt.Sprintf("<%d> ", facility*8+DEBUG.Code())
		parts := strings.Split(string(w.buf.Bytes()), delimiter)
		if len(parts) != 13 {
			t.Errorf("expected 13 parts of log string, got: %d\n", len(parts))
		}
		if parts[0][:len(pri)] != pri {
			t.Errorf("expected start of string: '%s', got: '%s'\n", pri, string(parts[0][:len(pri)]))
		}
		if parts[6] != "test-api-1" {
			t.Errorf("expected service name \"%s\", got: \"%s\"\n", service, parts[6])
		}
		if parts[7] != DEBUG.String() {
			t.Errorf("expected log level \"%s\", got: \"%s\"\n", DEBUG.String(), parts[7])
		}
	} else {
		t.Fail()
	}

	l = NewLogger("test-api-2", "", "", DEBUG)
	l.writer = &interceptorWriter{}

	l.Record(DEBUG, &format.Std{})
	if w, ok := l.writer.(*interceptorWriter); ok {
		parts := strings.Split(string(w.buf.Bytes()), delimiter)
		if len(parts) != 13 {
			t.Errorf("expected 13 parts of log string, got: %d\n", len(parts))
		}
		ts, err := time.Parse(time.RFC3339Nano, parts[1])
		if err != nil {
			t.Errorf("expected timestamp format: \"2006-01-02T15:04:05.999999Z07:00\", got: \"%s\"\n",
				parts[1])
		}
		since := time.Since(ts)
		if since > time.Millisecond {
			t.Errorf("expected duration between now and timestamp: < 1 ms, got: %s\n", since)
		}
		if parts[6] != "test-api-2" {
			t.Errorf("expected service name \"%s\", got: \"%s\"\n", service, parts[6])
		}
		if parts[7] != DEBUG.String() {
			t.Errorf("expected log level \"%s\", got: \"%s\"\n", DEBUG.String(), parts[7])
		}
	} else {
		t.Fail()
	}
}

type mockWriter struct {
	sleep time.Duration
	i     int64
}

func (w *mockWriter) Write(p []byte) (n int, err error) {
	time.Sleep(w.sleep)
	atomic.AddInt64(&w.i, 1)
	return len(p), nil
}

func (w *mockWriter) Close() error {
	return nil
}

func TestFlushWithoutMessages(t *testing.T) {
	resetWriters()
	l := New(WithServiceName("test-flush"), WithDialOptions("", ""), WithLevel(DEBUG), WithWorkerCount(4))
	w := &mockWriter{}
	l.writer.(*bufferLayer).writer = w

	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()
	if err := l.Flush(ctx); err != nil {
		t.Errorf("Expected error from Flush is <nil>, got: %s", err)
	}
}

func TestFlushWithMessagesCountLessThenWorkersCount(t *testing.T) {
	l := New(WithServiceName("test-flush"), WithDialOptions("", ""), WithLevel(DEBUG), WithWorkerCount(100))
	w := &mockWriter{sleep: time.Millisecond}
	l.writer.(*bufferLayer).writer = w

	c := 10
	for i := 0; i < c; i++ {
		if err := l.Record(DEBUG, &format.Std{}); err != nil {
			t.Fatalf("Expected error from Record is <nil>, got: %s", err)
		}
	}
	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()
	if err := l.Flush(ctx); err != nil {
		t.Errorf("Expected error from Flush is <nil>, got: %s", err)
	}

	if v := atomic.LoadInt64(&w.i); v != int64(c) {
		t.Errorf("Expected recorded messages: %d, got: %d", c, v)
	}
}

func TestFlushWithMessagesCountMoreThenWorkersCount(t *testing.T) {
	resetWriters()
	l := New(WithLevel(DEBUG), WithWorkerCount(10))
	w := &mockWriter{sleep: time.Millisecond}
	l.writer.(*bufferLayer).writer = w

	c := 100
	for i := 0; i < c; i++ {
		l.Record(DEBUG, &format.Std{})
	}

	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()
	if err := l.Flush(ctx); err != nil {
		t.Errorf("Expected error from Flush is <nil>, got: %s", err)
	}
}

func TestFlushAsync(t *testing.T) {
	resetWriters()
	l := New(WithLevel(DEBUG), WithWorkerCount(10))
	w := &mockWriter{sleep: time.Millisecond}
	l.writer.(*bufferLayer).writer = w
	done := make(chan struct{})
	var c int64
	go func() {
		for {
			l.Record(DEBUG, &format.Std{})
			atomic.AddInt64(&c, 1)
			select {
			case <-done:
				return
			default:
			}
		}
	}()
	for atomic.LoadInt64(&c) < 100 {
		time.Sleep(time.Millisecond)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()
	if err := l.Flush(ctx); err != nil {
		t.Errorf("Expected error from Flush is <nil>, got: %s", err)
	}
	if v := atomic.LoadInt64(&w.i); v < 100 {
		t.Errorf("Expected recorded messages >= 100, got: %d", v)
	}
	close(done)
}

func TestFlushTimeout(t *testing.T) {
	resetWriters()
	l := New(WithLevel(DEBUG), WithWorkerCount(10))
	w := &mockWriter{sleep: 10 * time.Millisecond}
	l.writer.(*bufferLayer).writer = w
	l.Record(DEBUG, &format.Std{})

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Millisecond)
	defer cancel()
	if err := l.Flush(ctx); err != context.DeadlineExceeded {
		t.Errorf("Expected error from Flush is %s, got: %s", context.DeadlineExceeded, err)
	}
}
