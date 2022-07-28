package slog

import (
	"bytes"
	"fmt"
	"log"
	"strings"
	"sync"
)

const rexPrefix = "rex:"

var (
	_ Logger = (*Scavenger)(nil)
)

type Printer interface {
	Print(level, message string)
}

type LogEntry struct {
	Level   string
	Message string
}

type internalData struct {
	mu      sync.Mutex
	entries []LogEntry
}

// Scavenger collects all log messages for later queries.
type Scavenger struct {
	logger        *ConsoleLogger
	extraPrinters []Printer

	*internalData
	buf bytes.Buffer
}

// NewScavenger creates a new Scavenger.
func NewScavenger(printers ...Printer) *Scavenger {
	return newScavengerImpl(&internalData{}, printers)
}

func newScavengerImpl(data *internalData, printers []Printer) *Scavenger {
	sc := &Scavenger{internalData: data}
	stdLog := log.New(&sc.buf, "", 0)
	sc.logger = NewConsoleLogger(WithStdLogger(stdLog), WithBareMode())
	for _, p := range printers {
		if p != nil {
			sc.extraPrinters = append(sc.extraPrinters, p)
		}
	}
	return sc
}

func (sc *Scavenger) NewLoggerWith(keyVals ...any) Logger {
	scav := newScavengerImpl(sc.internalData, sc.extraPrinters)
	combineFields(sc.logger.fields, keyVals...).fn(scav.logger)
	return scav
}

func (sc *Scavenger) FlushLogger() error {
	sc.mu.Lock()
	defer sc.mu.Unlock()

	var firstErr error
	if err := sc.logger.FlushLogger(); err != nil {
		firstErr = err
	}
	for _, printer := range sc.extraPrinters {
		x, ok := printer.(interface {
			Sync() error
		})
		if ok {
			if err := x.Sync(); err != nil && firstErr == nil {
				firstErr = err
			}
		}
	}
	return firstErr
}

func (sc *Scavenger) addEntryImpl(e LogEntry) {
	sc.entries = append(sc.entries, e)
	for _, p := range sc.extraPrinters {
		p.Print(e.Level, e.Message)
	}
}

// AddEntry adds a new log entry with the help of fmt.Sprint.
func (sc *Scavenger) AddEntry(level string, args []any) LogEntry {
	sc.mu.Lock()
	defer sc.mu.Unlock()

	sc.buf.Reset()
	sc.logger.println(level, args)
	str := strings.TrimRight(sc.buf.String(), "\n")
	entry := LogEntry{Level: level, Message: str}
	sc.addEntryImpl(entry)
	return entry
}

func (sc *Scavenger) Debug(args ...any) {
	sc.AddEntry(LevelDebug, args)
}

func (sc *Scavenger) Info(args ...any) {
	sc.AddEntry(LevelInfo, args)
}

func (sc *Scavenger) Warn(args ...any) {
	sc.AddEntry(LevelWarn, args)
}

func (sc *Scavenger) Error(args ...any) {
	sc.AddEntry(LevelError, args)
}

// AddEntryf adds a new log entry with the help of fmt.Sprintf.
func (sc *Scavenger) AddEntryf(level string, format string, args []any) LogEntry {
	sc.mu.Lock()
	defer sc.mu.Unlock()

	sc.buf.Reset()
	sc.logger.printf(level, format, args)
	str := strings.TrimRight(sc.buf.String(), "\n")
	entry := LogEntry{Level: level, Message: str}
	sc.addEntryImpl(entry)
	return entry
}

func (sc *Scavenger) Debugf(format string, args ...any) {
	sc.AddEntryf(LevelDebug, format, args)
}

func (sc *Scavenger) Infof(format string, args ...any) {
	sc.AddEntryf(LevelInfo, format, args)
}

func (sc *Scavenger) Warnf(format string, args ...any) {
	sc.AddEntryf(LevelWarn, format, args)
}

func (sc *Scavenger) Errorf(format string, args ...any) {
	sc.AddEntryf(LevelError, format, args)
}

// AddEntryw adds a new log entry. The variadic key-value pairs are treated as they are in NewLoggerWith.
func (sc *Scavenger) AddEntryw(level string, msg string, keyVals ...any) LogEntry {
	sc.mu.Lock()
	defer sc.mu.Unlock()

	sc.buf.Reset()
	sc.logger.printw(level, msg, keyVals)
	str := strings.TrimRight(sc.buf.String(), "\n")
	entry := LogEntry{Level: level, Message: str}
	sc.addEntryImpl(entry)
	return entry
}

func (sc *Scavenger) Debugw(msg string, keyVals ...any) {
	sc.AddEntryw(LevelDebug, msg, keyVals...)
}

func (sc *Scavenger) Infow(msg string, keyVals ...any) {
	sc.AddEntryw(LevelInfo, msg, keyVals...)
}

func (sc *Scavenger) Warnw(msg string, keyVals ...any) {
	sc.AddEntryw(LevelWarn, msg, keyVals...)
}

func (sc *Scavenger) Errorw(msg string, keyVals ...any) {
	sc.AddEntryw(LevelError, msg, keyVals...)
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

	scav := NewScavenger(sc.extraPrinters...)
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
