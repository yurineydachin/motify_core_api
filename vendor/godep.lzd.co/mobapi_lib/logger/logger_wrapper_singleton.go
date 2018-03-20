package logger

import (
	"context"
	"fmt"
	"sync"
)

const (
	singletonCallStackLevelSkip = 3
)

var (
	mtx            sync.RWMutex
	loggerInstance *loggerInstanceWrapper
)

func Init(serviceName, syslogAddrType, syslogAddr, level string) error {
	mtx.Lock()
	defer mtx.Unlock()

	if loggerInstance != nil {
		return fmt.Errorf("Logger is already initialized. Could not initialize it twice")
	}
	if serviceName == "" {
		return fmt.Errorf("Could not initialize logger: empty service name")
	}

	loggerInstance = NewLogger(serviceName, syslogAddrType, syslogAddr, level)
	if loggerInstance == nil {
		return fmt.Errorf("Could not initialize logger. Used parameters: service '%s', addrType '%s', addr '%s', level '%s'", serviceName, syslogAddrType, syslogAddr, level)
	}

	return nil
}

func GetLoggerInstance() *loggerInstanceWrapper {
	return loggerInstance
}

func ParseAndSetLevel(level string) bool {
	return GetLoggerInstance().ParseAndSetLevel(level)
}

func SetLevel(level int) {
	GetLoggerInstance().SetLevel(level)
}

func GetLevel() int {
	return GetLoggerInstance().GetLevel()
}

func Debug(ctx context.Context, message string, args ...interface{}) {
	GetLoggerInstance().Logf(ctx, singletonCallStackLevelSkip, DEBUG, message, nil, args...)
}

func Info(ctx context.Context, message string, args ...interface{}) {
	GetLoggerInstance().Logf(ctx, singletonCallStackLevelSkip, INFO, message, nil, args...)
}

func Notice(ctx context.Context, message string, args ...interface{}) {
	GetLoggerInstance().Logf(ctx, singletonCallStackLevelSkip, NOTICE, message, nil, args...)
}

func Warning(ctx context.Context, message string, args ...interface{}) {
	GetLoggerInstance().Logf(ctx, singletonCallStackLevelSkip, WARNING, message, nil, args...)
}

func Error(ctx context.Context, message string, args ...interface{}) {
	GetLoggerInstance().Logf(ctx, singletonCallStackLevelSkip, ERROR, message, nil, args...)
}

func Critical(ctx context.Context, message string, args ...interface{}) {
	GetLoggerInstance().Logf(ctx, singletonCallStackLevelSkip, CRITICAL, message, nil, args...)
}

func Alert(ctx context.Context, message string, args ...interface{}) {
	GetLoggerInstance().Logf(ctx, singletonCallStackLevelSkip, ALERT, message, nil, args...)
}

func Emergency(ctx context.Context, message string, args ...interface{}) {
	GetLoggerInstance().Logf(ctx, singletonCallStackLevelSkip, EMERGENCY, message, nil, args...)
}

func Flush(ctx context.Context) error {
	return GetLoggerInstance().Flush(ctx)
}
