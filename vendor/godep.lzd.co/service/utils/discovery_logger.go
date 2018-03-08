package utils

import (
	"fmt"
	"godep.lzd.co/service/logger"
)

type DiscoveryLogger struct{}

func (DiscoveryLogger) Debugf(format string, args ...interface{}) {
	logger.Debug(nil, format, args...)
}

func (DiscoveryLogger) Infof(format string, args ...interface{}) {
	logger.Info(nil, format, args...)
}

func (DiscoveryLogger) Warningf(format string, args ...interface{}) {
	logger.Warning(nil, format, args...)
}

func (DiscoveryLogger) Errorf(format string, args ...interface{}) {
	logger.Error(nil, format, args...)
}

func (DiscoveryLogger) Debug(args ...interface{}) {
	logger.Debug(nil, fmt.Sprint(args...))
}

func (DiscoveryLogger) Info(args ...interface{}) {
	logger.Info(nil, fmt.Sprint(args...))
}

func (DiscoveryLogger) Warning(args ...interface{}) {
	logger.Warning(nil, fmt.Sprint(args...))
}

func (DiscoveryLogger) Error(args ...interface{}) {
	logger.Error(nil, fmt.Sprint(args...))
}
