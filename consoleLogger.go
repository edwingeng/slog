package slog

import (
	"fmt"
	"log"
	"strings"
)

const (
	LevelDebug   = "DEBUG"
	LevelInfo    = "INFO"
	LevelWarning = "WARNING"
	LevelError   = "ERROR"
)

type ConsoleLogger struct{}

func NewConsoleLogger() ConsoleLogger {
	return ConsoleLogger{}
}

func (cl ConsoleLogger) println(level string, args []interface{}) {
	if len(args) == 0 {
		return
	}
	a := make([]interface{}, len(args))
	copy(a, args)
	a[0] = fmt.Sprintf("%s\t%v", level, a[0])
	log.Println(a...)
}

func (cl ConsoleLogger) Debug(args ...interface{}) {
	cl.println(LevelDebug, args)
}

func (cl ConsoleLogger) Info(args ...interface{}) {
	cl.println(LevelInfo, args)
}

func (cl ConsoleLogger) Warn(args ...interface{}) {
	cl.println(LevelWarning, args)
}

func (cl ConsoleLogger) Error(args ...interface{}) {
	cl.println(LevelError, args)
}

func (cl ConsoleLogger) printf(level string, format string, args []interface{}) {
	if strings.HasSuffix(format, "\n") {
		format = fmt.Sprintf("%s\t%s", level, format)
		log.Printf(format, args...)
		return
	}

	format = fmt.Sprintf("%s\t%s\n", level, format)
	log.Printf(format, args...)
}

func (cl ConsoleLogger) Debugf(format string, args ...interface{}) {
	cl.printf(LevelDebug, format, args)
}

func (cl ConsoleLogger) Infof(format string, args ...interface{}) {
	cl.printf(LevelInfo, format, args)
}

func (cl ConsoleLogger) Warnf(format string, args ...interface{}) {
	cl.printf(LevelWarning, format, args)
}

func (cl ConsoleLogger) Errorf(format string, args ...interface{}) {
	cl.printf(LevelError, format, args)
}

func (cl ConsoleLogger) Print(level, message string) {
	log.Printf("%s\t%v\n", level, message)
}
