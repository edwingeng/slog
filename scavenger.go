package slog

import (
	"bytes"
	"fmt"
	"log"
	"strings"
	"sync"
)

const rexPrefix = "rex:"

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

type Scavenger struct {
	logger        *ConsoleLogger
	extraPrinters []Printer

	*internalData
	buf bytes.Buffer
}

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

func (sc *Scavenger) NewLoggerWith(keyVals ...interface{}) Logger {
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

func (sc *Scavenger) AddEntry(level string, args []interface{}) LogEntry {
	sc.mu.Lock()
	defer sc.mu.Unlock()

	sc.buf.Reset()
	sc.logger.println(level, args)
	str := strings.TrimRight(sc.buf.String(), "\n")
	entry := LogEntry{Level: level, Message: str}
	sc.addEntryImpl(entry)
	return entry
}

func (sc *Scavenger) Debug(args ...interface{}) {
	sc.AddEntry(LevelDebug, args)
}

func (sc *Scavenger) Info(args ...interface{}) {
	sc.AddEntry(LevelInfo, args)
}

func (sc *Scavenger) Warn(args ...interface{}) {
	sc.AddEntry(LevelWarn, args)
}

func (sc *Scavenger) Error(args ...interface{}) {
	sc.AddEntry(LevelError, args)
}

func (sc *Scavenger) AddEntryf(level string, format string, args []interface{}) LogEntry {
	sc.mu.Lock()
	defer sc.mu.Unlock()

	sc.buf.Reset()
	sc.logger.printf(level, format, args)
	str := strings.TrimRight(sc.buf.String(), "\n")
	entry := LogEntry{Level: level, Message: str}
	sc.addEntryImpl(entry)
	return entry
}

func (sc *Scavenger) Debugf(format string, args ...interface{}) {
	sc.AddEntryf(LevelDebug, format, args)
}

func (sc *Scavenger) Infof(format string, args ...interface{}) {
	sc.AddEntryf(LevelInfo, format, args)
}

func (sc *Scavenger) Warnf(format string, args ...interface{}) {
	sc.AddEntryf(LevelWarn, format, args)
}

func (sc *Scavenger) Errorf(format string, args ...interface{}) {
	sc.AddEntryf(LevelError, format, args)
}

func (sc *Scavenger) AddEntryw(level string, msg string, keyVals ...interface{}) LogEntry {
	sc.mu.Lock()
	defer sc.mu.Unlock()

	sc.buf.Reset()
	sc.logger.printw(level, msg, keyVals)
	str := strings.TrimRight(sc.buf.String(), "\n")
	entry := LogEntry{Level: level, Message: str}
	sc.addEntryImpl(entry)
	return entry
}

func (sc *Scavenger) Debugw(msg string, keyVals ...interface{}) {
	sc.AddEntryw(LevelDebug, msg, keyVals...)
}

func (sc *Scavenger) Infow(msg string, keyVals ...interface{}) {
	sc.AddEntryw(LevelInfo, msg, keyVals...)
}

func (sc *Scavenger) Warnw(msg string, keyVals ...interface{}) {
	sc.AddEntryw(LevelWarn, msg, keyVals...)
}

func (sc *Scavenger) Errorw(msg string, keyVals ...interface{}) {
	sc.AddEntryw(LevelError, msg, keyVals...)
}

func (sc *Scavenger) Reset() {
	sc.mu.Lock()
	sc.entries = nil
	sc.mu.Unlock()
}

func (sc *Scavenger) Finder() *MessageFinder {
	return (*MessageFinder)(sc)
}

func (sc *Scavenger) Entries() []LogEntry {
	sc.mu.Lock()
	defer sc.mu.Unlock()

	clone := make([]LogEntry, len(sc.entries))
	copy(clone, sc.entries)
	return clone
}

func (sc *Scavenger) Len() int {
	sc.mu.Lock()
	n := len(sc.entries)
	sc.mu.Unlock()
	return n
}

func (sc *Scavenger) Dump() string {
	sc.mu.Lock()
	defer sc.mu.Unlock()

	var sb strings.Builder
	for _, e := range sc.entries {
		_, _ = fmt.Fprintf(&sb, "%s\t%s\n", e.Level, e.Message)
	}
	return sb.String()
}

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

func (sc *Scavenger) StringExists(str string) (yes bool) {
	_, _, yes = sc.Finder().FindString(str)
	return
}

func (sc *Scavenger) UniqueStringExists(str string) (yes bool) {
	_, _, yes = sc.Finder().FindUniqueString(str)
	return
}

func (sc *Scavenger) StringSequenceExists(a []string) (yes bool) {
	_, yes = sc.Finder().FindStringSequence(a)
	return
}

func (sc *Scavenger) RegexpExists(str string) (yes bool) {
	_, _, yes = sc.Finder().FindRegexp(str)
	return
}

func (sc *Scavenger) UniqueRegexpExists(str string) (yes bool) {
	_, _, yes = sc.Finder().FindUniqueRegexp(str)
	return
}

func (sc *Scavenger) RegexpSequenceExists(a []string) (yes bool) {
	_, yes = sc.Finder().FindRegexpSequence(a)
	return
}

func (sc *Scavenger) Exists(str string) (yes bool) {
	_, _, yes = sc.Finder().Find(str)
	return
}

func (sc *Scavenger) UniqueExists(str string) (yes bool) {
	_, _, yes = sc.Finder().FindUnique(str)
	return
}

func (sc *Scavenger) SequenceExists(a []string) (yes bool) {
	_, yes = sc.Finder().FindSequence(a)
	return
}
