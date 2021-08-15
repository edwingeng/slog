package slog

import (
	"bytes"
	"encoding/json"
	"fmt"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strconv"
	"strings"

	"github.com/fatih/color"
)

const (
	LevelDebug = "DEBUG"
	LevelInfo  = "INFO"
	LevelWarn  = "WARN"
	LevelError = "ERROR"
)

var (
	clrDebug = color.New(color.FgMagenta, color.Faint).Sprint(LevelDebug)
	clrInfo  = color.New(color.FgGreen, color.Faint).Sprint(LevelInfo)
	clrWarn  = color.New(color.FgYellow).Sprint(LevelWarn)
	clrError = color.New(color.FgRed).Sprint(LevelError)
	clrGray  = color.New(color.FgWhite, color.Faint)
)

type workingMode int8

const (
	modeColored workingMode = iota
	modeWithoutColor
	modeBare
)

type ConsoleLogger struct {
	stdLog *log.Logger

	extraSkip    int
	mode         workingMode
	disableDebug bool
	disableInfo  bool
	disableWarn  bool

	fields string
}

func NewConsoleLogger(opts ...Option) ConsoleLogger {
	var cl ConsoleLogger
	cl.stdLog = log.New(os.Stderr, "", log.Ltime)
	for _, opt := range opts {
		opt(&cl)
	}
	if cl.mode == modeBare {
		cl.stdLog.SetFlags(0)
	}
	return cl
}

func (cl ConsoleLogger) NewLoggerWith(keyVals ...interface{}) Logger {
	newLogger := cl
	if len(keyVals) == 0 {
		return newLogger
	}

	combineFields(cl.fields, keyVals...)(&newLogger)
	return newLogger
}

func combineFields(fields string, keyVals ...interface{}) Option {
	if len(fields) == 0 {
		return WithFields(keyVals...)
	}

	var m map[string]interface{}
	err := json.Unmarshal([]byte(fields), &m)
	if err != nil {
		panic(err)
	}
	return withFieldsImpl(m, keyVals...)
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
	if cl.mode == modeBare {
		return buf
	}

	file, line, ok := caller(3 + cl.extraSkip)
	if cl.mode == modeColored {
		switch level[0] {
		case 'D':
			buf.WriteString(clrDebug)
		case 'I':
			buf.WriteString(clrInfo)
		case 'W':
			buf.WriteString(clrWarn)
		case 'E':
			buf.WriteString(clrError)
		default:
			buf.WriteString(level)
		}
	} else {
		buf.WriteString(level)
	}
	if len(level) == 4 {
		_ = buf.WriteByte(' ')
	}
	if ok && file != "" {
		if cl.mode == modeColored {
			_, _ = clrGray.Fprintf(&buf, " %s:%d\t", file, line)
		} else {
			_, _ = fmt.Fprintf(&buf, " %s:%d\t", file, line)
		}
	} else {
		if cl.mode == modeColored {
			_, _ = clrGray.Fprint(&buf, " ?\t")
		} else {
			_, _ = buf.WriteString(" ?\t")
		}
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

	kvs := replaceZapFields(keyVals)
	n := len(kvs)
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
		d := stringlize(kvs[i+1])
		_, _ = fmt.Fprintf(&buf, "%q: %s", fmt.Sprint(kvs[i]), d)
	}

	if fLen > 0 || n > 1 {
		buf.WriteByte('}')
	}

	_ = buf.WriteByte('\n')
	_ = cl.stdLog.Output(0, buf.String())
}

func stringlize(v interface{}) string {
	switch x := v.(type) {
	case string:
		return strconv.Quote(x)
	case int, int8, int16, int32, int64:
		return fmt.Sprint(v)
	case uint, uint8, uint16, uint32, uint64:
		return fmt.Sprint(v)
	case float32, float64:
		return fmt.Sprint(v)
	case bool:
		return fmt.Sprint(v)
	case uintptr:
		return fmt.Sprintf("%#x", v)
	}

	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetEscapeHTML(false)
	err := enc.Encode(v)
	if err != nil {
		buf.Reset()
		_, _ = fmt.Fprintf(&buf, "%+v", v)
		return buf.String()
	}

	data := buf.Bytes()
	n1 := len(data)
	if n1 >= 2 {
		if data[0] == '"' {
			n2 := n1 - 1
			if data[n2] == '"' {
				return string(data[1:n2])
			}
		}
	}
	return string(data)
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

func WithFields(keyVals ...interface{}) Option {
	return withFieldsImpl(make(map[string]interface{}), keyVals...)
}

func withFieldsImpl(m map[string]interface{}, keyVals ...interface{}) Option {
	keyVals = replaceZapFields(keyVals)
	for i := 0; i < len(keyVals)-1; i += 2 {
		k := fmt.Sprint(keyVals[i])
		m[k] = keyVals[i+1]
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
		_, _ = fmt.Fprintf(&buf, "%q: %s", a[i], stringlize(m[a[i]]))
	}
	buf.WriteByte('}')

	return func(cl *ConsoleLogger) {
		cl.fields = buf.String()
	}
}

func replaceZapFields(keyVals []interface{}) []interface{} {
	var a []interface{}
	for i, n := 0, len(keyVals); i < n; i++ {
		if f, ok := keyVals[i].(zap.Field); ok {
			encoder := zapcore.NewMapObjectEncoder()
			f.AddTo(encoder)
			a = append(a, f.Key, encoder.Fields[f.Key])
		} else if i+1 < n {
			a = append(a, keyVals[i], keyVals[i+1])
			i++
		}
	}
	return a
}

func WithBareMode() Option {
	return func(cl *ConsoleLogger) {
		cl.mode = modeBare
	}
}

func WithoutColor() Option {
	return func(cl *ConsoleLogger) {
		if cl.mode == modeBare {
			return
		}
		cl.mode = modeWithoutColor
	}
}
