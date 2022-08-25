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

func (DumbLogger) NewLoggerWith(keyVals ...any) Logger {
	return DumbLogger{}
}

func (DumbLogger) Debug(args ...any) {}
func (DumbLogger) Info(args ...any)  {}
func (DumbLogger) Warn(args ...any)  {}
func (DumbLogger) Error(args ...any) {}

func (DumbLogger) Debugf(format string, args ...any) {}
func (DumbLogger) Infof(format string, args ...any)  {}
func (DumbLogger) Warnf(format string, args ...any)  {}
func (DumbLogger) Errorf(format string, args ...any) {}

func (DumbLogger) Debugw(msg string, keyVals ...any) {}
func (DumbLogger) Infow(msg string, keyVals ...any)  {}
func (DumbLogger) Warnw(msg string, keyVals ...any)  {}
func (DumbLogger) Errorw(msg string, keyVals ...any) {}

func (DumbLogger) FlushLogger() error { return nil }
