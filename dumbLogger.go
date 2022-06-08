package slog

var (
	_ Logger = DumbLogger{}
)

// DumbLogger devours all log messages and outputs nothing.
type DumbLogger struct{}

// NewDumbLogger creates a new DumbLogger.
func NewDumbLogger() DumbLogger {
	return DumbLogger{}
}

func (DumbLogger) NewLoggerWith(keyVals ...interface{}) Logger {
	return DumbLogger{}
}

func (DumbLogger) Debug(args ...interface{}) {}
func (DumbLogger) Info(args ...interface{})  {}
func (DumbLogger) Warn(args ...interface{})  {}
func (DumbLogger) Error(args ...interface{}) {}

func (DumbLogger) Debugf(format string, args ...interface{}) {}
func (DumbLogger) Infof(format string, args ...interface{})  {}
func (DumbLogger) Warnf(format string, args ...interface{})  {}
func (DumbLogger) Errorf(format string, args ...interface{}) {}

func (DumbLogger) Debugw(msg string, keyVals ...interface{}) {}
func (DumbLogger) Infow(msg string, keyVals ...interface{})  {}
func (DumbLogger) Warnw(msg string, keyVals ...interface{})  {}
func (DumbLogger) Errorw(msg string, keyVals ...interface{}) {}

func (DumbLogger) FlushLogger() error { return nil }
