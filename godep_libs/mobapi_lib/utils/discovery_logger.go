package utils

import (
	"fmt"
	"motify_core_api/godep_libs/mobapi_lib/logger"
)

type DiscoveryLogger struct{}

func (DiscoveryLogger) Debugf(format string, args ...interface{}) {
	logger.GetLoggerInstance().Logf(nil, 5, logger.DEBUG, format, nil, args...)
}

func (DiscoveryLogger) Infof(format string, args ...interface{}) {
	logger.GetLoggerInstance().Logf(nil, 5, logger.INFO, format, nil, args...)
}

func (DiscoveryLogger) Warningf(format string, args ...interface{}) {
	logger.GetLoggerInstance().Logf(nil, 5, logger.WARNING, format, nil, args...)
}

func (DiscoveryLogger) Errorf(format string, args ...interface{}) {
	logger.GetLoggerInstance().Logf(nil, 5, logger.ERROR, format, nil, args...)
}

func (DiscoveryLogger) Debug(args ...interface{}) {
	logger.GetLoggerInstance().Logf(nil, 5, logger.DEBUG, fmt.Sprint(args...), nil)
}

func (DiscoveryLogger) Info(args ...interface{}) {
	logger.GetLoggerInstance().Logf(nil, 5, logger.INFO, fmt.Sprint(args...), nil)
}

func (DiscoveryLogger) Warning(args ...interface{}) {
	logger.GetLoggerInstance().Logf(nil, 5, logger.WARNING, fmt.Sprint(args...), nil)
}

func (DiscoveryLogger) Error(args ...interface{}) {
	logger.GetLoggerInstance().Logf(nil, 5, logger.ERROR, fmt.Sprint(args...), nil)
}
