package slog

type Logger interface {
	NewLoggerWith(keyVals ...interface{}) Logger

	Debug(args ...interface{})
	Info(args ...interface{})
	Warn(args ...interface{})
	Error(args ...interface{})

	Debugf(format string, args ...interface{})
	Infof(format string, args ...interface{})
	Warnf(format string, args ...interface{})
	Errorf(format string, args ...interface{})

	Debugw(msg string, keyVals ...interface{})
	Infow(msg string, keyVals ...interface{})
	Warnw(msg string, keyVals ...interface{})
	Errorw(msg string, keyVals ...interface{})

	FlushLogger() error
}
