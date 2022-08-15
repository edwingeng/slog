package slog

import (
	"bytes"
	"fmt"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"runtime"
	"strconv"
	"strings"
	"sync"
)

const rexPrefix = "rex:"

var (
	_ Logger = &Scavenger{}
)

var (
	lineEnding = []byte(zapcore.DefaultLineEnding)
)

func goID() int {
	var buf [64]byte
	n := runtime.Stack(buf[:], false)
	idField := strings.Fields(strings.TrimPrefix(string(buf[:n]), "goroutine "))[0]
	id, err := strconv.Atoi(idField)
	if err != nil {
		panic(fmt.Sprintf("cannot get goroutine id: %v", err))
	}
	return id
}

type LogEntry struct {
	Level   string `json:"level"`
	Message string `json:"message"`
}

type entryHolder struct {
	mu      sync.Mutex
	entries []LogEntry
}

// Scavenger collects all log messages for later queries.
type Scavenger struct {
	*entryHolder
	buf *bytes.Buffer

	x   zap.SugaredLogger
	kvs []any
}

// NewScavenger creates a new Scavenger.
func NewScavenger() *Scavenger {
	sinkName := fmt.Sprintf("scavenger-%d", goID())
	sink := &memorySink{}
	sinkRegistry.Lock()
	sinkRegistry.m[sinkName] = sink
	sinkRegistry.Unlock()
	defer func() {
		sinkRegistry.Lock()
		delete(sinkRegistry.m, sinkName)
		sinkRegistry.Unlock()
	}()

	cfg := zap.NewDevelopmentConfig()
	cfg.OutputPaths = []string{"memory://" + sinkName}
	cfg.ErrorOutputPaths = []string{"memory://" + sinkName}
	cfg.DisableStacktrace = true
	cfg.DisableCaller = true
	cfg.EncoderConfig.TimeKey = ""
	cfg.EncoderConfig.LevelKey = ""

	l, err := cfg.Build()
	if err != nil {
		panic(err)
	}

	var sc Scavenger
	sc.x = *l.Sugar()
	sc.buf = &sink.buf
	sc.entryHolder = &entryHolder{}
	return &sc
}

func (sc *Scavenger) NewLoggerWith(keyVals ...any) Logger {
	kvs := append(sc.kvs, keyVals...)
	scav := NewScavenger()
	scav.x = *scav.x.With(kvs...)
	scav.entryHolder = sc.entryHolder
	scav.kvs = kvs
	return scav
}

func (sc *Scavenger) Debug(args ...any) {
	sc.mu.Lock()
	defer sc.mu.Unlock()
	sc.buf.Reset()
	sc.x.Debug(args...)
	sc.collectEntry(LevelDebug)
}

func (sc *Scavenger) Info(args ...any) {
	sc.mu.Lock()
	defer sc.mu.Unlock()
	sc.buf.Reset()
	sc.x.Info(args...)
	sc.collectEntry(LevelInfo)
}

func (sc *Scavenger) Warn(args ...any) {
	sc.mu.Lock()
	defer sc.mu.Unlock()
	sc.buf.Reset()
	sc.x.Warn(args...)
	sc.collectEntry(LevelWarn)
}

func (sc *Scavenger) Error(args ...any) {
	sc.mu.Lock()
	defer sc.mu.Unlock()
	sc.buf.Reset()
	sc.x.Error(args...)
	sc.collectEntry(LevelError)
}

func (sc *Scavenger) Debugf(format string, args ...any) {
	sc.mu.Lock()
	defer sc.mu.Unlock()
	sc.buf.Reset()
	sc.x.Debugf(format, args...)
	sc.collectEntry(LevelDebug)
}

func (sc *Scavenger) Infof(format string, args ...any) {
	sc.mu.Lock()
	defer sc.mu.Unlock()
	sc.buf.Reset()
	sc.x.Infof(format, args...)
	sc.collectEntry(LevelInfo)
}

func (sc *Scavenger) Warnf(format string, args ...any) {
	sc.mu.Lock()
	defer sc.mu.Unlock()
	sc.buf.Reset()
	sc.x.Warnf(format, args...)
	sc.collectEntry(LevelWarn)
}

func (sc *Scavenger) Errorf(format string, args ...any) {
	sc.mu.Lock()
	defer sc.mu.Unlock()
	sc.buf.Reset()
	sc.x.Errorf(format, args...)
	sc.collectEntry(LevelError)
}

func (sc *Scavenger) Debugw(msg string, keyVals ...any) {
	sc.mu.Lock()
	defer sc.mu.Unlock()
	sc.buf.Reset()
	sc.x.Debugw(msg, keyVals...)
	sc.collectEntry(LevelDebug)
}

func (sc *Scavenger) Infow(msg string, keyVals ...any) {
	sc.mu.Lock()
	defer sc.mu.Unlock()
	sc.buf.Reset()
	sc.x.Infow(msg, keyVals...)
	sc.collectEntry(LevelInfo)
}

func (sc *Scavenger) Warnw(msg string, keyVals ...any) {
	sc.mu.Lock()
	defer sc.mu.Unlock()
	sc.buf.Reset()
	sc.x.Warnw(msg, keyVals...)
	sc.collectEntry(LevelWarn)
}

func (sc *Scavenger) Errorw(msg string, keyVals ...any) {
	sc.mu.Lock()
	defer sc.mu.Unlock()
	sc.buf.Reset()
	sc.x.Errorw(msg, keyVals...)
	sc.collectEntry(LevelError)
}

func (sc *Scavenger) collectEntry(level string) {
	d1 := sc.buf.Bytes()
	d2 := bytes.TrimSuffix(d1, lineEnding)

	if bytes.Count(d2, lineEnding) > 0 {
		a := bytes.Split(d2, lineEnding)
		for i, v := range a {
			switch i {
			case len(a) - 1:
				sc.entries = append(sc.entries, LogEntry{
					Level:   level,
					Message: string(v),
				})
			default:
				sc.entries = append(sc.entries, LogEntry{
					Level:   LevelError,
					Message: string(v),
				})
			}
		}
		return
	}

	sc.entries = append(sc.entries, LogEntry{
		Level:   level,
		Message: string(d2),
	})
}

func (sc *Scavenger) FlushLogger() error {
	return nil
}

// Reset clears all collected messages.
func (sc *Scavenger) Reset() {
	sc.mu.Lock()
	sc.entries = nil
	sc.mu.Unlock()
}

// Finder returns a MessageFinder.
func (sc *Scavenger) Finder() *MessageFinder {
	return (*MessageFinder)(sc)
}

// Entries returns a duplicate of the collected messages.
func (sc *Scavenger) Entries() []LogEntry {
	sc.mu.Lock()
	defer sc.mu.Unlock()

	clone := make([]LogEntry, len(sc.entries))
	copy(clone, sc.entries)
	return clone
}

// Len returns the number of the collected messages.
func (sc *Scavenger) Len() int {
	sc.mu.Lock()
	n := len(sc.entries)
	sc.mu.Unlock()
	return n
}

// Dump returns a string that contains all the collected messages.
func (sc *Scavenger) Dump() string {
	sc.mu.Lock()
	defer sc.mu.Unlock()

	var sb strings.Builder
	for _, e := range sc.entries {
		_, _ = fmt.Fprintf(&sb, "%s\t%s\n", e.Level, e.Message)
	}
	return sb.String()
}

// Filter creates a new Scavenger that contains only the log messages satisfying the predicate fn.
func (sc *Scavenger) Filter(fn func(level, msg string) bool) *Scavenger {
	sc.mu.Lock()
	defer sc.mu.Unlock()

	scav := NewScavenger()
	scav.entries = make([]LogEntry, 0, len(sc.entries))
	for _, e := range sc.entries {
		if fn == nil || fn(e.Level, e.Message) {
			scav.entries = append(scav.entries, e)
		}
	}
	return scav
}

// StringExists returns whether any collected message contains str.
func (sc *Scavenger) StringExists(str string) (yes bool) {
	_, _, yes = sc.Finder().FindString(str)
	return
}

// UniqueStringExists returns whether one and only one collected message contains str.
func (sc *Scavenger) UniqueStringExists(str string) (yes bool) {
	_, _, yes = sc.Finder().FindUniqueString(str)
	return
}

// FindStringSequence returns whether the collected messages contain the specified sequence.
func (sc *Scavenger) FindStringSequence(seq []string) (found int, yes bool) {
	return sc.Finder().FindStringSequence(seq)
}

// RegexpExists returns whether any collected message contains the regular expression pat.
func (sc *Scavenger) RegexpExists(pat string) (yes bool) {
	_, _, yes = sc.Finder().FindRegexp(pat)
	return
}

// UniqueRegexpExists returns whether one and only one collected message contains the regular expression pat.
func (sc *Scavenger) UniqueRegexpExists(pat string) (yes bool) {
	_, _, yes = sc.Finder().FindUniqueRegexp(pat)
	return
}

// FindRegexpSequence returns whether the collected messages contain the specified regular expression sequence.
func (sc *Scavenger) FindRegexpSequence(seq []string) (found int, yes bool) {
	return sc.Finder().FindRegexpSequence(seq)
}

// Exists returns whether any collected message contains str.
// If str starts with "rex: ", it is regarded as a regular expression.
func (sc *Scavenger) Exists(str string) (yes bool) {
	_, _, yes = sc.Finder().Find(str)
	return
}

// UniqueExists returns whether one and only one collected message contains str.
// If str starts with "rex: ", it is regarded as a regular expression.
func (sc *Scavenger) UniqueExists(str string) (yes bool) {
	_, _, yes = sc.Finder().FindUnique(str)
	return
}

// FindSequence returns whether the collected messages contain the specified sequence.
// If a string in seq starts with "rex: ", it is regarded as a regular expression.
func (sc *Scavenger) FindSequence(seq []string) (found int, yes bool) {
	return sc.Finder().FindSequence(seq)
}
