package slog

type Logger interface {
	NewLoggerWith(keyVals ...interface{}) Logger

	Debug(args ...interface{})
	Info(args ...interface{})
	Warn(args ...interface{})
	Error(args ...interface{})

	Debugf(template string, args ...interface{})
	Infof(template string, args ...interface{})
	Warnf(template string, args ...interface{})
	Errorf(template string, args ...interface{})

	Debugw(msg string, keyVals ...interface{})
	Infow(msg string, keyVals ...interface{})
	Warnw(msg string, keyVals ...interface{})
	Errorw(msg string, keyVals ...interface{})

	FlushLogger() error
}
