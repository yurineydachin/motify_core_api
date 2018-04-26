package sessionlogger

import (
	"fmt"
	"os"
	"regexp"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/davecgh/go-spew/spew"
	"godep.lzd.co/service/logger"
	"godep.lzd.co/service/sessionlogger/libs/file"
	"godep.lzd.co/service/sessionlogger/libs/uniq_dumper"
	"html"
)

const noData = "NO_DATA"

// FileLogWriter writes logging data to file and dumps to berkeley db.
type FileLogWriter struct {
	object *writer
}

type writer struct {
	mutex         sync.RWMutex
	date          time.Time
	file          *file.SelfRescuingFile
	dumper        *uniq_dumper.Dumper
	queue         chan *Log
	queueFinished chan struct{}
	queueOnce     sync.Once
	stopped       bool
}

func NewFileLogWriter(date time.Time, fileName, dirName string) (*FileLogWriter, error) {
	logFile, err := file.OpenFile(fileName, os.O_CREATE|os.O_APPEND|os.O_RDWR, 0640)
	if err != nil {
		return nil, err
	}

	w := &writer{
		date:          date,
		file:          logFile,
		dumper:        uniq_dumper.New(dirName),
		queue:         make(chan *Log, 256),
		queueFinished: make(chan struct{}),
	}

	logWriter := &FileLogWriter{object: w}

	runtime.SetFinalizer(logWriter, func(logWriter *FileLogWriter) { logWriter.Close() })
	go func() { w.writingLoop() }()

	return logWriter, nil
}

func (logWriter *FileLogWriter) Date() time.Time {
	return logWriter.object.Date()
}
func (logWriter *FileLogWriter) Write(l *Log) {
	logWriter.object.Write(l)
}
func (logWriter *FileLogWriter) Close() error {
	return logWriter.object.Close()
}

func (w *writer) Date() time.Time {
	return w.date
}

func (w *writer) writingLoop() {
	defer close(w.queueFinished)

	for l := range w.queue {
		if err, ok := l.data.(error); ok {
			l.data = err.Error()
		}

		caption := escapeLogValue(l.caption)
		fmtTimestamp := l.timestamp.Format(time.RFC3339Nano)

		dump := ""
		if data, ok := l.data.(string); ok && data == noData {
			dump = noData
		} else {
			dump = strings.Replace(html.EscapeString(removeAddresses(strings.Trim(spew.Sdump(l.data), "\n"))), "\n", "<br/>", -1)

			var err error
			dump, err = w.dumper.Write([]byte(dump))
			if err != nil {
				logger.Error(nil, "Error on dump writing: %s", err)
			}
		}

		_, err := fmt.Fprintf(w.file, "%s\t%s\t%d\t%d\t%s\t%s\t%s\n",
			fmtTimestamp, l.traceID, l.id, l.parentID, l.eventType, caption, dump)
		if err != nil {
			logger.Error(nil, "%v", err)
			continue
		}
	}
}

var addrRE = regexp.MustCompile(`\(0x[0-9a-fA-F]+(?:\->0x[0-9a-fA-F]+)*\)`)

func removeAddresses(str string) string {
	return addrRE.ReplaceAllString(str, "")
}

const writingTimeout = time.Second

func (w *writer) Write(l *Log) {
	w.mutex.RLock()
	defer w.mutex.RUnlock()

	if w.stopped {
		return
	}

	select {
	case w.queue <- l:
	default:
		loggingError := fmt.Sprintf(
			"Log queue is overflowed. Can't write log: %#v\n",
			&Log{
				traceID:   l.traceID,
				id:        l.id,
				parentID:  l.parentID,
				caption:   l.caption,
				eventType: l.eventType,
				timestamp: l.timestamp,
			},
		)
		if l.eventType == "ERR" {
			select {
			case w.queue <- l:
			case <-time.After(writingTimeout):
				logger.Error(nil, "%v", loggingError)
			}
		} else {
			logger.Error(nil, "%v", loggingError)
		}
	}
}

// Close closes log writer.
// Keep in mind that after close all subsequent logs would be discarded.
// That's why it should be called after some timeout.
func (w *writer) Close() error {
	w.mutex.Lock()
	defer w.mutex.Unlock()

	if w.stopped {
		return nil
	}

	close(w.queue)
	<-w.queueFinished
	w.stopped = true

	fileErr := w.file.Close()

	if fileErr != nil {
		return fmt.Errorf("file.close() err: %v", fileErr)
	}
	return nil
}

func escapeLogValue(str string) string {
	str = strings.Replace(str, "\n", "\\n", -1)
	str = strings.Replace(str, "\t", "\\t", -1)
	return str
}
