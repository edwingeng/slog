package slog

var (
	_ Logger = devourer{}
)

type devourer struct{}

// NewDevourer creates a new Logger that devours all log messages and outputs nothing.
func NewDevourer() Logger {
	return devourer{}
}

func (devourer) NewLoggerWith(keyVals ...any) Logger { return devourer{} }

func (devourer) LogLevelEnabled(level int) bool { return false }

func (devourer) Debug(args ...any) {}
func (devourer) Info(args ...any)  {}
func (devourer) Warn(args ...any)  {}
func (devourer) Error(args ...any) {}

func (devourer) Debugf(format string, args ...any) {}
func (devourer) Infof(format string, args ...any)  {}
func (devourer) Warnf(format string, args ...any)  {}
func (devourer) Errorf(format string, args ...any) {}

func (devourer) Debugw(msg string, keyVals ...any) {}
func (devourer) Infow(msg string, keyVals ...any)  {}
func (devourer) Warnw(msg string, keyVals ...any)  {}
func (devourer) Errorw(msg string, keyVals ...any) {}

func (devourer) FlushLogger() error { return nil }
