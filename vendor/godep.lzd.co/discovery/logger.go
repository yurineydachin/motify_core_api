package discovery

// ILogger is an common interface for logging in discovery package.
type ILogger interface {
	Debugf(format string, args ...interface{})
	Infof(format string, args ...interface{})
	Warningf(format string, args ...interface{})
	Errorf(format string, args ...interface{})
	Debug(args ...interface{})
	Info(args ...interface{})
	Warning(args ...interface{})
	Error(args ...interface{})
}

// nilLogger is a logger which does nothing.
type nilLogger struct{}

// NewNilLogger returns new nilLogger instance
func NewNilLogger() ILogger {
	return &nilLogger{}
}

// Debugf does nothing
func (l *nilLogger) Debugf(format string, args ...interface{}) {}

// Infof does nothing
func (l *nilLogger) Infof(format string, args ...interface{}) {}

// Warningf does nothing
func (l *nilLogger) Warningf(format string, args ...interface{}) {}

// Errorf does nothing
func (l *nilLogger) Errorf(format string, args ...interface{}) {}

// Debug does nothing
func (l *nilLogger) Debug(args ...interface{}) {}

// Info does nothing
func (l *nilLogger) Info(args ...interface{}) {}

// Warning does nothing
func (l *nilLogger) Warning(args ...interface{}) {}

// Error does nothing
func (l *nilLogger) Error(args ...interface{}) {}
