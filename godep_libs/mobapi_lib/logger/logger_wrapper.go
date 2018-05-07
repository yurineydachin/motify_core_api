package logger

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/opentracing/opentracing-go"
	"motify_core_api/godep_libs/go-log"
	"motify_core_api/godep_libs/go-trace"
)

const (
	// for backward compatibility
	DEBUG int = iota
	INFO
	NOTICE
	WARNING
	ERROR
	CRITICAL
	ALERT
	EMERGENCY

	// used for writing log during `go test`
	failoverLogFormat = "%s | %s | %s | %s | stable | %s | %s | FAILOVER_LOGGER | - | %s | {} | .\n"

	callStackLevelSkip = 2
)

var printMtx sync.Mutex

type loggerInstanceWrapper struct {
	serviceName string
	writer      *log.Logger
}

func NewLogger(serviceName, syslogAddrType, syslogAddr string, levelStr string) *loggerInstanceWrapper {
	s, ok := log.ParseSeverity(levelStr)
	if !ok {
		return nil
	}
	return &loggerInstanceWrapper{
		serviceName: serviceName,
		writer:      log.NewLogger(serviceName, syslogAddrType, syslogAddr, s),
	}
}

func (l *loggerInstanceWrapper) Writer() *log.Logger {
	return l.writer
}

func (l *loggerInstanceWrapper) ParseAndSetLevel(level string) bool {
	s, ok := log.ParseSeverity(level)
	if !ok {
		return false
	}

	l.writer.SetLevel(s)

	return true
}

func (l *loggerInstanceWrapper) SetLevel(level int) {
	l.writer.SetLevel(convertOldLevelToNewSeverity(level))
}

func (l *loggerInstanceWrapper) GetLevel() int {
	return convertNewSeverityToOldLevel(l.writer.Level())
}

func (l *loggerInstanceWrapper) Flush(ctx context.Context) error {
	return l.writer.Flush(ctx)
}

func (l *loggerInstanceWrapper) writeLog(ctx context.Context, callStackSkip int, level log.Severity, format string, data map[string]interface{}, args ...interface{}) {
	if ctx == nil {
		ctx = context.Background()
	}
	if l == nil || l.writer == nil {
		// if it's `go test` execution
		// we need this hack for UT (when logger is called without full service initialization)
		l.writeForcedStdOutLog(ctx, convertNewSeverityToOldLevel(level), format, data, args...)
	} else {
		trackingData, _ := gotrace.SpanContext(opentracing.SpanFromContext(ctx))
		span := log.Span{
			TraceID:      trackingData.TraceID,
			SpanID:       trackingData.SpanID,
			ParentSpanID: trackingData.ParentSpanID,
		}
		l.writer.Logf(callStackSkip, span, level, format, data, args...)
	}
}

func (l *loggerInstanceWrapper) writeForcedStdOutLog(ctx context.Context, level int, format string, data map[string]interface{}, args ...interface{}) {
	var serviceName string
	trackingData, _ := gotrace.SpanContext(opentracing.SpanFromContext(ctx))
	timeFormatted := time.Now().UTC().Format(time.RFC3339Nano)
	severity := convertOldLevelToNewSeverity(level)

	//if it's not 'go test' and logger is initialized
	if l != nil {
		serviceName = l.serviceName
	}

	printMtx.Lock()
	defer printMtx.Unlock()

	s := fmt.Sprintf(format, args)

	fmt.Printf(failoverLogFormat, timeFormatted, trackingData.TraceID, trackingData.ParentSpanID, trackingData.SpanID, serviceName, severity, s)
}

func (l *loggerInstanceWrapper) Debug(ctx context.Context, message string, args ...interface{}) {
	l.writeLog(ctx, callStackLevelSkip, log.DEBUG, message, nil, args...)
}

func (l *loggerInstanceWrapper) Info(ctx context.Context, message string, args ...interface{}) {
	l.writeLog(ctx, callStackLevelSkip, log.INFO, message, nil, args...)
}

func (l *loggerInstanceWrapper) Notice(ctx context.Context, message string, args ...interface{}) {
	l.writeLog(ctx, callStackLevelSkip, log.NOTICE, message, nil, args...)
}

func (l *loggerInstanceWrapper) Warning(ctx context.Context, message string, args ...interface{}) {
	l.writeLog(ctx, callStackLevelSkip, log.WARNING, message, nil, args...)
}

func (l *loggerInstanceWrapper) Error(ctx context.Context, message string, args ...interface{}) {
	l.writeLog(ctx, callStackLevelSkip, log.ERROR, message, nil, args...)
}

func (l *loggerInstanceWrapper) Critical(ctx context.Context, message string, args ...interface{}) {
	l.writeLog(ctx, callStackLevelSkip, log.CRITICAL, message, nil, args...)
}

func (l *loggerInstanceWrapper) Alert(ctx context.Context, message string, args ...interface{}) {
	l.writeLog(ctx, callStackLevelSkip, log.ALERT, message, nil, args...)
}

func (l *loggerInstanceWrapper) Emergency(ctx context.Context, message string, args ...interface{}) {
	l.writeLog(ctx, callStackLevelSkip, log.EMERGENCY, message, nil, args...)
}

func (l *loggerInstanceWrapper) Logf(ctx context.Context, callStackSkip int, level int, message string, data map[string]interface{}, args ...interface{}) {
	l.writeLog(ctx, callStackSkip, convertOldLevelToNewSeverity(level), message, data, args...)
}

func convertOldLevelToNewSeverity(level int) log.Severity {
	switch level {
	case DEBUG:
		return log.DEBUG
	case INFO:
		return log.INFO
	case NOTICE:
		return log.NOTICE
	case WARNING:
		return log.WARNING
	case ERROR:
		return log.ERROR
	case CRITICAL:
		return log.CRITICAL
	case ALERT:
		return log.ALERT
	case EMERGENCY:
		return log.EMERGENCY
	default:
		return log.Severity(-1)
	}
}

func convertNewSeverityToOldLevel(severity log.Severity) int {
	switch severity {
	case log.DEBUG:
		return DEBUG
	case log.INFO:
		return INFO
	case log.NOTICE:
		return NOTICE
	case log.WARNING:
		return WARNING
	case log.ERROR:
		return ERROR
	case log.CRITICAL:
		return CRITICAL
	case log.ALERT:
		return ALERT
	case log.EMERGENCY:
		return EMERGENCY
	default:
		return -1
	}
}
