package slog

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
)

const (
	LevelDebug = "DEBUG"
	LevelInfo  = "INFO"
	LevelWarn  = "WARN"
	LevelError = "ERROR"
)

type ConsoleLogger struct {
	stdLog *log.Logger

	extraSkip    int
	bare         bool
	disableDebug bool
	disableInfo  bool
	disableWarn  bool

	fields string
}

func NewConsoleLogger(opts ...Option) ConsoleLogger {
	var cl ConsoleLogger
	cl.stdLog = log.New(os.Stderr, "", log.LstdFlags)
	for _, opt := range opts {
		opt(&cl)
	}
	if cl.bare {
		cl.stdLog.SetFlags(0)
	}
	return cl
}

func (cl ConsoleLogger) NewLoggerWith(args ...interface{}) Logger {
	newLogger := cl
	if len(args) == 0 {
		return newLogger
	}

	newLogger.combineFields(cl.fields, args...)
	return newLogger
}

func (this *ConsoleLogger) combineFields(fields string, args ...interface{}) {
	if len(this.fields) == 0 {
		WithFields(args...)(this)
		return
	}

	var m map[string]interface{}
	err := json.Unmarshal([]byte(fields), &m)
	if err != nil {
		panic(err)
	}
	withFieldsImpl(m, args...)(this)
}

func caller(skip int) (string, int, bool) {
	_, file, line, ok := runtime.Caller(skip + 1)
	if ok && file != "" {
		if idx := strings.LastIndex(file, string(filepath.Separator)); idx >= 0 {
			file = file[idx+1:]
		}
	}
	return file, line, ok
}

func (cl ConsoleLogger) buildHeader(level string) bytes.Buffer {
	var buf bytes.Buffer
	if cl.bare {
		return buf
	}

	file, line, ok := caller(3 + cl.extraSkip)
	buf.WriteString(level)
	if len(level) == 4 {
		_ = buf.WriteByte(' ')
	}
	if ok && file != "" {
		_, _ = fmt.Fprintf(&buf, " %s:%d\t", file, line)
	} else {
		_, _ = buf.WriteString(" ?\t")
	}
	return buf
}

func (cl ConsoleLogger) println(level string, args []interface{}) {
	if len(args) == 0 {
		return
	}

	buf := cl.buildHeader(level)
	written, _ := fmt.Fprint(&buf, args...)
	cl.outputImpl(&buf, written)
}

func (cl ConsoleLogger) outputImpl(buf *bytes.Buffer, written int) {
	if len(cl.fields) > 0 {
		if written > 0 {
			_ = buf.WriteByte('\t')
		}
		_, _ = buf.WriteString(cl.fields)
	}
	_ = cl.stdLog.Output(0, buf.String())
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
	buf := cl.buildHeader(level)
	formatted := strings.TrimSuffix(fmt.Sprintf(format, args...), "\n")
	written, _ := buf.WriteString(formatted)
	cl.outputImpl(&buf, written)
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

func (cl ConsoleLogger) Print(level, msg string) {
	buf := cl.buildHeader(level)
	written, _ := buf.WriteString(strings.TrimSuffix(msg, "\n"))
	cl.outputImpl(&buf, written)
}

func (cl ConsoleLogger) Debugw(msg string, keyVals ...interface{}) {
	if !cl.disableDebug {
		cl.printw(LevelDebug, msg, keyVals)
	}
}

func (cl ConsoleLogger) Infow(msg string, keyVals ...interface{}) {
	if !cl.disableInfo {
		cl.printw(LevelInfo, msg, keyVals)
	}
}

func (cl ConsoleLogger) Warnw(msg string, keyVals ...interface{}) {
	if !cl.disableWarn {
		cl.printw(LevelWarn, msg, keyVals)
	}
}

func (cl ConsoleLogger) Errorw(msg string, keyVals ...interface{}) {
	cl.printw(LevelError, msg, keyVals)
}

func (cl ConsoleLogger) printw(level string, msg string, keyVals []interface{}) {
	buf := cl.buildHeader(level)
	written, _ := buf.WriteString(strings.TrimSuffix(msg, "\n"))

	fLen := len(cl.fields)
	if fLen > 0 {
		if written > 0 {
			_ = buf.WriteByte('\t')
		}
		_, _ = buf.WriteString(cl.fields[:fLen-1])
	}

	n := len(keyVals)
	for i := 0; i < n-1; i += 2 {
		if i > 0 {
			_, _ = buf.WriteString(", ")
		} else if fLen > 0 {
			_, _ = buf.WriteString(", ")
		} else {
			if written > 0 {
				buf.WriteByte('\t')
			}
			buf.WriteByte('{')
		}
		d, err := json.Marshal(keyVals[i+1])
		if err != nil {
			panic(err)
		}
		_, _ = fmt.Fprintf(&buf, "%q:%s", fmt.Sprint(keyVals[i]), d)
	}

	if fLen > 0 || n > 1 {
		buf.WriteByte('}')
	}

	_ = buf.WriteByte('\n')
	_ = cl.stdLog.Output(0, buf.String())
}

func (cl ConsoleLogger) FlushLogger() error {
	if writer := cl.stdLog.Writer(); writer != nil {
		x, ok := cl.stdLog.Writer().(interface {
			Sync() error
		})
		if ok {
			return x.Sync()
		}
	}
	return nil
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

func WithStdLogger(stdLog *log.Logger) Option {
	return func(cl *ConsoleLogger) {
		stdLog.SetFlags(stdLog.Flags() &^ (log.Llongfile | log.Lshortfile))
		cl.stdLog = stdLog
	}
}

func WithFields(args ...interface{}) Option {
	return withFieldsImpl(make(map[string]interface{}), args...)
}

func withFieldsImpl(m map[string]interface{}, args ...interface{}) Option {
	for i := 0; i < len(args)-1; i += 2 {
		k := fmt.Sprint(args[i])
		m[k] = args[i+1]
	}
	if len(m) == 0 {
		return func(cl *ConsoleLogger) {
			// Empty
		}
	}

	a := make([]string, 0, len(m))
	for k := range m {
		a = append(a, k)
	}
	sort.Strings(a)

	var buf bytes.Buffer
	buf.WriteByte('{')
	for i := range a {
		if i > 0 {
			buf.WriteString(", ")
		}
		d, err := json.Marshal(m[a[i]])
		if err != nil {
			panic(err)
		}
		_, _ = fmt.Fprintf(&buf, "%q:%s", a[i], d)
	}
	buf.WriteByte('}')

	return func(cl *ConsoleLogger) {
		cl.fields = buf.String()
	}
}

func WithBareMode() Option {
	return func(cl *ConsoleLogger) {
		cl.bare = true
	}
}
