package slog

import (
	"go.uber.org/zap"
)

var (
	_ Logger = ZapLogger{}
)

// ZapLogger is a wrapper of *zap.SugaredLogger.
type ZapLogger struct {
	x *zap.SugaredLogger
}

// NewZapLogger creates a new ZapLogger.
func NewZapLogger(zsl *zap.SugaredLogger) ZapLogger {
	return ZapLogger{x: zsl}
}

// Zap returns the internal *zap.SugaredLogger to the caller.
func (zl ZapLogger) Zap() *zap.SugaredLogger {
	return zl.x
}

func (zl ZapLogger) NewLoggerWith(keyVals ...interface{}) Logger {
	return ZapLogger{x: zl.x.With(keyVals...)}
}

func (zl ZapLogger) Debug(args ...interface{}) {
	zl.x.Debug(args...)
}

func (zl ZapLogger) Info(args ...interface{}) {
	zl.x.Info(args...)
}

func (zl ZapLogger) Warn(args ...interface{}) {
	zl.x.Warn(args...)
}

func (zl ZapLogger) Error(args ...interface{}) {
	zl.x.Error(args...)
}

func (zl ZapLogger) Debugf(format string, args ...interface{}) {
	zl.x.Debugf(format, args...)
}

func (zl ZapLogger) Infof(format string, args ...interface{}) {
	zl.x.Infof(format, args...)
}

func (zl ZapLogger) Warnf(format string, args ...interface{}) {
	zl.x.Warnf(format, args...)
}

func (zl ZapLogger) Errorf(format string, args ...interface{}) {
	zl.x.Errorf(format, args...)
}

func (zl ZapLogger) Debugw(msg string, keyVals ...interface{}) {
	zl.x.Debugw(msg, keyVals...)
}

func (zl ZapLogger) Infow(msg string, keyVals ...interface{}) {
	zl.x.Infow(msg, keyVals...)
}

func (zl ZapLogger) Warnw(msg string, keyVals ...interface{}) {
	zl.x.Warnw(msg, keyVals...)
}

func (zl ZapLogger) Errorw(msg string, keyVals ...interface{}) {
	zl.x.Errorw(msg, keyVals...)
}

func (zl ZapLogger) FlushLogger() error {
	return zl.x.Sync()
}
