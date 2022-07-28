package slog

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/fatih/color"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"log"
	"os"
	"runtime"
	"sort"
	"strconv"
	"strings"
)

const (
	LevelDebug = "DEBUG"
	LevelInfo  = "INFO"
	LevelWarn  = "WARN"
	LevelError = "ERROR"
)

var (
	_ Logger = (*ConsoleLogger)(nil)
)

var (
	clrDebug = color.New(color.FgMagenta).Sprint(LevelDebug)
	clrInfo  = color.New(color.FgGreen).Sprint(LevelInfo)
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

// ConsoleLogger is a simple console logger that implements the Logger interface.
type ConsoleLogger struct {
	stdLog *log.Logger

	extraSkip    int32
	mode         workingMode
	disableDebug bool
	disableInfo  bool
	disableWarn  bool

	fields string
}

// NewConsoleLogger creates a new ConsoleLogger.
func NewConsoleLogger(opts ...Option) *ConsoleLogger {
	var cl ConsoleLogger
	cl.stdLog = log.New(os.Stderr, "", log.Ltime)
	for i := 0; i <= 2; i++ {
		for _, opt := range opts {
			if opt.priority == i {
				opt.fn(&cl)
			}
		}
	}
	return &cl
}

func (cl *ConsoleLogger) NewLoggerWith(keyVals ...any) Logger {
	newLogger := *cl
	if len(keyVals) == 0 {
		return &newLogger
	}

	combineFields(cl.fields, keyVals...).fn(&newLogger)
	return &newLogger
}

func combineFields(fields string, keyVals ...any) Option {
	if len(fields) == 0 {
		return WithFields(keyVals...)
	}

	decoder := json.NewDecoder(strings.NewReader(fields))
	decoder.UseNumber()
	var m map[string]any
	err := decoder.Decode(&m)
	if err != nil {
		panic(err)
	}
	return withFieldsImpl(m, keyVals...)
}

func caller(skip int32) (string, int, bool) {
	_, file, line, ok := runtime.Caller(int(skip) + 1)
	if ok && file != "" {
		if idx := strings.LastIndexAny(file, `/\`); idx >= 0 {
			file = file[idx+1:]
		}
	}
	return file, line, ok
}

func (cl *ConsoleLogger) buildHeader(level string, extraSkip int32) bytes.Buffer {
	var buf bytes.Buffer
	if cl.mode == modeBare {
		return buf
	}

	if cl.mode == modeColored {
		switch level[0] {
		case 'D':
			_, _ = buf.WriteString(clrDebug)
		case 'I':
			_, _ = buf.WriteString(clrInfo)
		case 'W':
			_, _ = buf.WriteString(clrWarn)
		case 'E':
			_, _ = buf.WriteString(clrError)
		default:
			_, _ = buf.WriteString(level)
		}
	} else {
		_, _ = buf.WriteString(level)
	}
	if len(level) == 4 {
		_ = buf.WriteByte(' ')
	}

	file, line, ok := caller(3 + cl.extraSkip + extraSkip)
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

func (cl *ConsoleLogger) println(level string, args []any) {
	if len(args) == 0 {
		return
	}

	buf := cl.buildHeader(level, 0)
	written, _ := fmt.Fprint(&buf, args...)
	cl.outputImpl(&buf, written)
}

func (cl *ConsoleLogger) outputImpl(buf *bytes.Buffer, written int) {
	if len(cl.fields) > 0 {
		if written > 0 {
			_ = buf.WriteByte('\t')
			_, _ = buf.WriteString(cl.fields)
		} else {
			_, _ = buf.WriteString(cl.fields)
		}
	}
	_ = cl.stdLog.Output(0, buf.String())
}

func (cl *ConsoleLogger) Debug(args ...any) {
	if !cl.disableDebug {
		cl.println(LevelDebug, args)
	}
}

func (cl *ConsoleLogger) Info(args ...any) {
	if !cl.disableInfo {
		cl.println(LevelInfo, args)
	}
}

func (cl *ConsoleLogger) Warn(args ...any) {
	if !cl.disableWarn {
		cl.println(LevelWarn, args)
	}
}

func (cl *ConsoleLogger) Error(args ...any) {
	cl.println(LevelError, args)
}

func (cl *ConsoleLogger) printf(level string, format string, args []any) {
	buf := cl.buildHeader(level, 0)
	str := fmt.Sprintf(format, args...)
	written, _ := buf.WriteString(str)
	cl.outputImpl(&buf, written)
}

func (cl *ConsoleLogger) Debugf(format string, args ...any) {
	if !cl.disableDebug {
		cl.printf(LevelDebug, format, args)
	}
}

func (cl *ConsoleLogger) Infof(format string, args ...any) {
	if !cl.disableInfo {
		cl.printf(LevelInfo, format, args)
	}
}

func (cl *ConsoleLogger) Warnf(format string, args ...any) {
	if !cl.disableWarn {
		cl.printf(LevelWarn, format, args)
	}
}

func (cl *ConsoleLogger) Errorf(format string, args ...any) {
	cl.printf(LevelError, format, args)
}

// Print implements the Printer interface.
func (cl *ConsoleLogger) Print(level, msg string) {
	switch level[0] {
	case 'D':
		if cl.disableDebug {
			return
		}
	case 'I':
		if cl.disableInfo {
			return
		}
	case 'W':
		if cl.disableWarn {
			return
		}
	}

	buf := cl.buildHeader(level, -1)
	written, _ := buf.WriteString(msg)
	cl.outputImpl(&buf, written)
}

func (cl *ConsoleLogger) Debugw(msg string, keyVals ...any) {
	if !cl.disableDebug {
		cl.printw(LevelDebug, msg, keyVals)
	}
}

func (cl *ConsoleLogger) Infow(msg string, keyVals ...any) {
	if !cl.disableInfo {
		cl.printw(LevelInfo, msg, keyVals)
	}
}

func (cl *ConsoleLogger) Warnw(msg string, keyVals ...any) {
	if !cl.disableWarn {
		cl.printw(LevelWarn, msg, keyVals)
	}
}

func (cl *ConsoleLogger) Errorw(msg string, keyVals ...any) {
	cl.printw(LevelError, msg, keyVals)
}

func (cl *ConsoleLogger) printw(level string, msg string, keyVals []any) {
	buf := cl.buildHeader(level, 0)
	written, _ := buf.WriteString(msg)

	fLen := len(cl.fields)
	if fLen > 0 {
		if written > 0 {
			_ = buf.WriteByte('\t')
			_, _ = buf.WriteString(cl.fields[:fLen-1])
		} else {
			_, _ = buf.WriteString(cl.fields[:fLen-1])
		}
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

	_ = cl.stdLog.Output(0, buf.String())
}

func stringlize(v any) string {
	switch x := v.(type) {
	case string:
		return strconv.Quote(x)
	case json.Number:
		return string(x)
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

	return string(bytes.TrimSuffix(buf.Bytes(), []byte("\n")))
}

func (cl *ConsoleLogger) FlushLogger() error {
	writer := cl.stdLog.Writer()
	if writer == nil {
		return errors.New("nil writer")
	}
	syncer, ok := cl.stdLog.Writer().(interface {
		Sync() error
	})
	if ok {
		return syncer.Sync()
	}
	return nil
}

// Option works for ConsoleLogger only.
type Option struct {
	priority int
	fn       func(cl *ConsoleLogger)
}

// WithExtraCallerSkip sets the extra number of frames to skip from the
// top of the stacktrace.
func WithExtraCallerSkip(extraSkip int) Option {
	return Option{
		fn: func(cl *ConsoleLogger) {
			cl.extraSkip = int32(extraSkip)
		},
	}
}

// WithLevel sets the lowest message level to output, which must be LevelDebug,
// LevelInfo, LevelWarn or LevelError.
func WithLevel(level string) Option {
	return Option{
		fn: func(cl *ConsoleLogger) {
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
		},
	}
}

// WithStdLogger sets the internal log.Logger.
func WithStdLogger(stdLog *log.Logger) Option {
	return Option{
		fn: func(cl *ConsoleLogger) {
			stdLog.SetFlags(stdLog.Flags() &^ (log.Llongfile | log.Lshortfile))
			cl.stdLog = stdLog
		},
	}
}

// WithFields adds a variadic number of fields to the logging context.
func WithFields(keyVals ...any) Option {
	return withFieldsImpl(make(map[string]any), keyVals...)
}

func withFieldsImpl(m map[string]any, keyVals ...any) Option {
	keyVals = replaceZapFields(keyVals)
	for i := 0; i < len(keyVals)-1; i += 2 {
		k := fmt.Sprint(keyVals[i])
		m[k] = keyVals[i+1]
	}
	if len(m) == 0 {
		return Option{
			fn: func(cl *ConsoleLogger) {
				// Empty
			},
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

	return Option{
		fn: func(cl *ConsoleLogger) {
			cl.fields = buf.String()
		},
	}
}

func replaceZapFields(keyVals []any) []any {
	var a []any
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

// WithBareMode removes time, level and stacktrace from log entries.
func WithBareMode() Option {
	return Option{
		priority: 2,
		fn: func(cl *ConsoleLogger) {
			cl.mode = modeBare
			cl.stdLog.SetFlags(0)
		},
	}
}

// WithoutColor disables the color of log entries.
func WithoutColor() Option {
	return Option{
		priority: 1,
		fn: func(cl *ConsoleLogger) {
			cl.mode = modeWithoutColor
		},
	}
}
