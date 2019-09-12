package slog

import (
	"fmt"
	"regexp"
	"strings"
	"sync"
	"unicode"
)

const rexPrefix = "rex:"

type Printer interface {
	Print(level, message string)
}

type LogEntry struct {
	Level   string
	Message string
}

type Scavenger struct {
	extraPrinters []Printer

	mu      sync.Mutex
	entries []LogEntry
}

func NewScavenger(printers ...Printer) (scav *Scavenger) {
	scav = &Scavenger{}
	for _, p := range printers {
		if p != nil {
			scav.extraPrinters = append(scav.extraPrinters, p)
		}
	}
	return
}

func (this *Scavenger) AddEntry(level string, args []interface{}) LogEntry {
	this.mu.Lock()
	defer this.mu.Unlock()

	if len(args) == 1 {
		msg := fmt.Sprint(args[0])
		entry := LogEntry{level, msg}
		this.entries = append(this.entries, entry)
		for _, p := range this.extraPrinters {
			p.Print(level, msg)
		}
		return entry
	}

	var sb strings.Builder
	for i, arg := range args {
		switch i {
		case 0:
			_, _ = fmt.Fprint(&sb, arg)
		default:
			_, _ = fmt.Fprintf(&sb, ", %v", arg)
		}
	}
	msg := sb.String()
	entry := LogEntry{level, msg}
	this.entries = append(this.entries, entry)
	for _, p := range this.extraPrinters {
		p.Print(level, msg)
	}
	return entry
}

func (this *Scavenger) Debug(args ...interface{}) {
	this.AddEntry(LevelDebug, args)
}

func (this *Scavenger) Info(args ...interface{}) {
	this.AddEntry(LevelInfo, args)
}

func (this *Scavenger) Warn(args ...interface{}) {
	this.AddEntry(LevelWarning, args)
}

func (this *Scavenger) Error(args ...interface{}) {
	this.AddEntry(LevelError, args)
}

func (this *Scavenger) AddEntryf(level string, format string, args []interface{}) LogEntry {
	msg := fmt.Sprintf(format, args...)
	entry := LogEntry{level, msg}
	this.entries = append(this.entries, entry)
	for _, p := range this.extraPrinters {
		p.Print(level, msg)
	}
	return entry
}

func (this *Scavenger) Debugf(format string, args ...interface{}) {
	this.AddEntryf(LevelDebug, format, args)
}

func (this *Scavenger) Infof(format string, args ...interface{}) {
	this.AddEntryf(LevelInfo, format, args)
}

func (this *Scavenger) Warnf(format string, args ...interface{}) {
	this.AddEntryf(LevelWarning, format, args)
}

func (this *Scavenger) Errorf(format string, args ...interface{}) {
	this.AddEntryf(LevelError, format, args)
}

func (this *Scavenger) Reset() {
	this.mu.Lock()
	this.entries = nil
	this.mu.Unlock()
}

func (this *Scavenger) FindString(str string) (string, bool) {
	this.mu.Lock()
	defer this.mu.Unlock()

	if str != "" {
		for _, e := range this.entries {
			if strings.Contains(e.Message, str) {
				return e.Message, true
			}
		}
	} else {
		for _, e := range this.entries {
			if e.Message == "" {
				return e.Message, true
			}
		}
	}
	return "", false
}

func (this *Scavenger) FindUniqueString(str string) (string, bool) {
	this.mu.Lock()
	defer this.mu.Unlock()

	var r string
	var n int
	if str != "" {
		for _, e := range this.entries {
			if strings.Contains(e.Message, str) {
				r = e.Message
				switch n++; n {
				case 1:
				default:
					return r, false
				}
			}
		}
	} else {
		for _, e := range this.entries {
			if e.Message == "" {
				switch n++; n {
				case 1:
				default:
					return r, false
				}
			}
		}
	}
	return r, n == 1
}

func (this *Scavenger) FindStringSequence(a []string) (int, bool) {
	this.mu.Lock()
	defer this.mu.Unlock()

	var j int
	for i := 0; i < len(this.entries); i++ {
		if a[j] != "" {
			if strings.Contains(this.entries[i].Message, a[j]) {
				if j++; j >= len(a) {
					break
				}
			}
		} else {
			if this.entries[i].Message == "" {
				if j++; j >= len(a) {
					break
				}
			}
		}
	}
	return j, j == len(a)
}

func (this *Scavenger) FindRegexp(str string) (string, bool) {
	if str == "" {
		return this.FindString(str)
	}
	rex, err := regexp.Compile(str)
	if err != nil {
		panic(err)
	}

	this.mu.Lock()
	defer this.mu.Unlock()

	for _, e := range this.entries {
		if rex.FindString(e.Message) != "" {
			return e.Message, true
		}
	}
	return "", false
}

func (this *Scavenger) FindUniqueRegexp(str string) (string, bool) {
	if str == "" {
		return this.FindString(str)
	}
	rex, err := regexp.Compile(str)
	if err != nil {
		panic(err)
	}

	this.mu.Lock()
	defer this.mu.Unlock()

	var r string
	var n int
	for _, e := range this.entries {
		if rex.FindString(e.Message) != "" {
			r = e.Message
			switch n++; n {
			case 1:
			default:
				return r, false
			}
		}
	}
	return r, n == 1
}

func (this *Scavenger) FindRegexpSequence(a []string) (int, bool) {
	this.mu.Lock()
	defer this.mu.Unlock()

	var j int
	for i := 0; i < len(this.entries); i++ {
		if a[j] != "" {
			rex, err := regexp.Compile(a[j])
			if err != nil {
				panic(err)
			}
			if rex.FindString(this.entries[i].Message) != "" {
				if j++; j >= len(a) {
					break
				}
			}
		} else {
			if this.entries[i].Message == "" {
				if j++; j >= len(a) {
					break
				}
			}
		}
	}
	return j, j == len(a)
}

func (this *Scavenger) Find(str string) (string, bool) {
	if strings.HasPrefix(str, rexPrefix) {
		str = strings.TrimLeftFunc(strings.TrimPrefix(str, rexPrefix), unicode.IsSpace)
		return this.FindRegexp(str)
	} else {
		return this.FindString(str)
	}
}

func (this *Scavenger) FindUnique(str string) (string, bool) {
	if strings.HasPrefix(str, rexPrefix) {
		str = strings.TrimLeftFunc(strings.TrimPrefix(str, rexPrefix), unicode.IsSpace)
		return this.FindUniqueRegexp(str)
	} else {
		return this.FindUniqueString(str)
	}
}

func (this *Scavenger) FindSequence(a []string) (int, bool) {
	this.mu.Lock()
	defer this.mu.Unlock()

	var j int
	for i := 0; i < len(this.entries); i++ {
		if strings.HasPrefix(a[j], rexPrefix) {
			str := strings.TrimLeftFunc(strings.TrimPrefix(a[j], rexPrefix), unicode.IsSpace)
			if str != "" {
				rex, err := regexp.Compile(str)
				if err != nil {
					panic(err)
				}
				if rex.FindString(this.entries[i].Message) != "" {
					if j++; j >= len(a) {
						break
					}
				}
				continue
			}
		} else {
			if a[j] != "" {
				if strings.Contains(this.entries[i].Message, a[j]) {
					if j++; j >= len(a) {
						break
					}
				}
				continue
			}
		}
		if this.entries[i].Message == "" {
			if j++; j >= len(a) {
				break
			}
		}
	}
	return j, j == len(a)
}

func (this *Scavenger) Entries() []LogEntry {
	this.mu.Lock()
	defer this.mu.Unlock()

	clone := make([]LogEntry, len(this.entries))
	copy(clone, this.entries)
	return clone
}

func (this *Scavenger) Len() int {
	this.mu.Lock()
	n := len(this.entries)
	this.mu.Unlock()
	return n
}

func (this *Scavenger) Dump() string {
	this.mu.Lock()
	defer this.mu.Unlock()

	var sb strings.Builder
	for _, e := range this.entries {
		_, _ = fmt.Fprintf(&sb, "%s\t%s\n", e.Level, e.Message)
	}
	return sb.String()
}

func (this *Scavenger) Filter(f func(level, msg string) bool) (scav *Scavenger) {
	this.mu.Lock()
	defer this.mu.Unlock()

	scav = NewScavenger(this.extraPrinters...)
	scav.entries = make([]LogEntry, 0, len(this.entries))
	for _, e := range this.entries {
		if f == nil || f(e.Level, e.Message) {
			scav.entries = append(scav.entries, e)
		}
	}
	return
}
