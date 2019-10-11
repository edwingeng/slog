package slog

import (
	"fmt"
	"log"
	"path/filepath"
	"runtime"
	"strings"
)

const (
	LevelDebug = "DEBUG"
	LevelInfo  = "INFO"
	LevelWarn  = "WARN"
	LevelError = "ERROR"
)

type ConsoleLogger struct {
	extraSkip    int
	disableDebug bool
	disableInfo  bool
	disableWarn  bool
}

func NewConsoleLogger(opts ...Option) ConsoleLogger {
	var cl ConsoleLogger
	for _, opt := range opts {
		opt(&cl)
	}
	return cl
}

func (cl ConsoleLogger) caller(skip int) (string, int, bool) {
	_, file, line, ok := runtime.Caller(skip + 1)
	if ok && file != "" {
		if idx := strings.LastIndex(file, string(filepath.Separator)); idx >= 0 {
			file = file[idx+1:]
		}
	}
	return file, line, ok
}

func (cl ConsoleLogger) println(level string, args []interface{}) {
	if len(args) == 0 {
		return
	}
	a := make([]interface{}, len(args))
	copy(a, args)
	file, line, ok := cl.caller(2 + cl.extraSkip)
	if ok && file != "" {
		a[0] = fmt.Sprintf("%s\t%s:%d\t%v", level, file, line, a[0])
	} else {
		a[0] = fmt.Sprintf("%s\t?\t%v", level, a[0])
	}

	log.Println(a...)
}

func (cl ConsoleLogger) Debug(args ...interface{}) {
	if !cl.disableDebug {
		cl.println(LevelDebug, args)
	}
}

func (cl ConsoleLogger) Info(args ...interface{}) {
	if !cl.disableInfo {
		cl.println(LevelInfo, args)
	}
}

func (cl ConsoleLogger) Warn(args ...interface{}) {
	if !cl.disableWarn {
		cl.println(LevelWarn, args)
	}
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
	file, line, ok := cl.caller(2 + cl.extraSkip)
	if ok && file != "" {
		format = fmt.Sprintf("%s\t%s:%d\t%s\n", level, file, line, format)
	} else {
		format = fmt.Sprintf("%s\t?\t%s\n", level, format)
	}
	log.Printf(format, args...)
}

func (cl ConsoleLogger) Debugf(format string, args ...interface{}) {
	if !cl.disableDebug {
		cl.printf(LevelDebug, format, args)
	}
}

func (cl ConsoleLogger) Infof(format string, args ...interface{}) {
	if !cl.disableInfo {
		cl.printf(LevelInfo, format, args)
	}
}

func (cl ConsoleLogger) Warnf(format string, args ...interface{}) {
	if !cl.disableWarn {
		cl.printf(LevelWarn, format, args)
	}
}

func (cl ConsoleLogger) Errorf(format string, args ...interface{}) {
	cl.printf(LevelError, format, args)
}

func (cl ConsoleLogger) Print(level, message string) {
	file, line, ok := cl.caller(1 + cl.extraSkip)
	if ok && file != "" {
		log.Printf("%s\t%s:%d\t%s\n", level, file, line, message)
	} else {
		log.Printf("%s\t?\t%s\n", level, message)
	}
}

type Option func(cl *ConsoleLogger)

func WithExtraCallerSkip(extraSkip int) Option {
	return func(cl *ConsoleLogger) {
		cl.extraSkip = extraSkip
	}
}

func WithLevel(level string) Option {
	return func(cl *ConsoleLogger) {
		switch level {
		case LevelDebug:
		case LevelInfo:
			cl.disableDebug = true
		case LevelWarn:
			cl.disableDebug, cl.disableInfo = true, true
		case LevelError:
			cl.disableDebug, cl.disableInfo, cl.disableWarn = true, true, true
		default:
			panic("invalid level: " + level)
		}
	}
}
