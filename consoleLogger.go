package slog

import (
	"fmt"
	"log"
	"strings"
)

const (
	levelDebug = "DEBUG"
	levelInfo  = "INFO "
	levelWarn  = "WARNING"
	levelError = "ERROR"
)

type ConsoleLogger struct{}

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
	cl.println(levelDebug, args)
}

func (cl ConsoleLogger) Info(args ...interface{}) {
	cl.println(levelInfo, args)
}

func (cl ConsoleLogger) Warn(args ...interface{}) {
	cl.println(levelWarn, args)
}

func (cl ConsoleLogger) Error(args ...interface{}) {
	cl.println(levelError, args)
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
	cl.printf(levelDebug, format, args)
}

func (cl ConsoleLogger) Infof(format string, args ...interface{}) {
	cl.printf(levelInfo, format, args)
}

func (cl ConsoleLogger) Warnf(format string, args ...interface{}) {
	cl.printf(levelWarn, format, args)
}

func (cl ConsoleLogger) Errorf(format string, args ...interface{}) {
	cl.printf(levelError, format, args)
}
