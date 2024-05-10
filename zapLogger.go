package slog

import (
	"go.uber.org/zap"
)

var (
	_ Logger = &ZapLogger{}
)

// ZapLogger is a wrapper of zap.SugaredLogger.
type ZapLogger struct {
	x zap.SugaredLogger
	l zap.Logger
}

// NewZapLogger creates a new ZapLogger.
func NewZapLogger(zsl *zap.SugaredLogger) *ZapLogger {
	return &ZapLogger{
		x: *zsl.WithOptions(zap.AddCallerSkip(1)),
		l: *zsl.Desugar(),
	}
}

// Zap returns the internal zap.Logger to the caller.
func (zl *ZapLogger) Zap() *zap.Logger {
	return &zl.l
}

func (zl *ZapLogger) NewLoggerWith(keyVals ...any) Logger {
	zsl := zl.x.With(keyVals...).WithOptions(zap.AddCallerSkip(-1))
	return NewZapLogger(zsl)
}

func (zl *ZapLogger) LogLevelEnabled(level int) bool {
	return level >= int(zl.x.Level())
}

func (zl *ZapLogger) Debug(args ...any) {
	zl.x.Debug(args...)
}

func (zl *ZapLogger) Info(args ...any) {
	zl.x.Info(args...)
}

func (zl *ZapLogger) Warn(args ...any) {
	zl.x.Warn(args...)
}

func (zl *ZapLogger) Error(args ...any) {
	zl.x.Error(args...)
}

func (zl *ZapLogger) Debugf(format string, args ...any) {
	zl.x.Debugf(format, args...)
}

func (zl *ZapLogger) Infof(format string, args ...any) {
	zl.x.Infof(format, args...)
}

func (zl *ZapLogger) Warnf(format string, args ...any) {
	zl.x.Warnf(format, args...)
}

func (zl *ZapLogger) Errorf(format string, args ...any) {
	zl.x.Errorf(format, args...)
}

func (zl *ZapLogger) Debugw(msg string, keyVals ...any) {
	zl.x.Debugw(msg, keyVals...)
}

func (zl *ZapLogger) Infow(msg string, keyVals ...any) {
	zl.x.Infow(msg, keyVals...)
}

func (zl *ZapLogger) Warnw(msg string, keyVals ...any) {
	zl.x.Warnw(msg, keyVals...)
}

func (zl *ZapLogger) Errorw(msg string, keyVals ...any) {
	zl.x.Errorw(msg, keyVals...)
}

func (zl *ZapLogger) FlushLogger() error {
	return zl.x.Sync()
}
