package slog

type DumbLogger struct{}

func (DumbLogger) Debug(args ...interface{}) {}
func (DumbLogger) Info(args ...interface{})  {}
func (DumbLogger) Warn(args ...interface{})  {}
func (DumbLogger) Error(args ...interface{}) {}

func (DumbLogger) Debugf(format string, args ...interface{}) {}
func (DumbLogger) Infof(format string, args ...interface{})  {}
func (DumbLogger) Warnf(format string, args ...interface{})  {}
func (DumbLogger) Errorf(format string, args ...interface{}) {}
