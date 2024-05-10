package slog

import (
	"fmt"
	"math"
	"regexp"
	"runtime/debug"
	"strings"
	"sync"
	"testing"
)

func TestScavenger_Exists(t *testing.T) {
	var sc = NewScavenger()
	sc.Debug("1")
	sc.Info("it is a good day to die")
	sc.Warn("3", "c")
	sc.Error("4")

	if sc.Len() != 4 {
		t.Fatal("sc.Len() != 4")
	}

	dump := `DEBUG	1
INFO	it is a good day to die
WARN	3c
ERROR	4
`
	if sc.Dump() != dump {
		t.Fatal("something is wrong with Dump: " + sc.Dump())
	}

	if yes := sc.Exists(""); yes {
		t.Fatal("Exists does not work as expected")
	}
	if yes := sc.Exists("5"); yes {
		t.Fatal("Exists does not work as expected")
	}
	if yes := sc.Exists("rex: 5"); yes {
		t.Fatal("Exists does not work as expected")
	}
	if yes := sc.Exists("3"); !yes {
		t.Fatal("Exists does not work as expected")
	}

	sc.Debug()
	if yes := sc.Exists(""); !yes {
		t.Fatal("Exists does not work as expected")
	}

	sc.Reset()
	if len(sc.entries) > 0 {
		t.Fatal("Scavenger should be empty now")
	}
}

func TestScavenger_FindRegexp_Panic(t *testing.T) {
	defer func() {
		_ = recover()
	}()
	var sc Scavenger
	sc.Finder().FindRegexp("[")
	t.Fatal("RegexpExists should panic")
}

//gocyclo:ignore
func TestScavenger_Uniqueness(t *testing.T) {
	var sc = NewScavenger()
	sc.Debug("")
	sc.Debugf("%d", 1)
	sc.Infof("%s", "it is a good day to die")
	sc.Warnf("%d, %s", 3, "c")
	sc.Errorf("%d", 4)
	sc.Warnf("%s", "it is a good day to die")
	sc.Errorf("%d", 1)

	if sc.Len() != 7 {
		t.Fatal("sc.Len() != 7", sc.Len())
	}

	dump := `DEBUG	
DEBUG	1
INFO	it is a good day to die
WARN	3, c
ERROR	4
WARN	it is a good day to die
ERROR	1
`
	if sc.Dump() != dump {
		t.Fatal("something is wrong with Dump")
	}

	if ret := sc.Finder().FindString("1"); len(ret) != 2 {
		t.Fatal("FindString does not work as expected")
	} else if ret[0] != 1 || ret[1] != 6 {
		t.Fatal("FindString does not work as expected")
	}
	if ret := sc.Finder().FindString("it is a good day to die"); len(ret) != 2 {
		t.Fatal("FindString does not work as expected")
	} else if ret[0] != 2 || ret[1] != 5 {
		t.Fatal("FindString does not work as expected")
	}
	if ret := sc.Finder().FindString("3"); len(ret) != 1 {
		t.Fatal("FindString does not work as expected")
	}
	if ret := sc.Finder().FindString(""); len(ret) != 1 {
		t.Fatal("FindString does not work as expected")
	}

	if ret := sc.Finder().FindRegexp("1"); len(ret) == 1 {
		t.Fatal("FindRegexp does not work as expected")
	} else if ret[0] != 1 || ret[1] != 6 {
		t.Fatal("FindRegexp does not work as expected")
	}
	if ret := sc.Finder().FindRegexp("it is a good day to die"); len(ret) == 1 {
		t.Fatal("FindRegexp does not work as expected")
	} else if ret[0] != 2 || ret[1] != 5 {
		t.Fatal("FindRegexp does not work as expected")
	}
	if ret := sc.Finder().FindRegexp("3"); len(ret) != 1 {
		t.Fatal("FindRegexp does not work as expected")
	}
	if ret := sc.Finder().FindRegexp("[3,4]"); len(ret) == 1 {
		t.Fatal("FindRegexp does not work as expected")
	}

	sc.Debug("")
	if ret := sc.Finder().FindString(""); len(ret) == 1 {
		t.Fatal("FindString does not work as expected")
	}
	if ret := sc.Finder().FindRegexp(""); len(ret) == 1 {
		t.Fatal("FindRegexp does not work as expected")
	}

	if yes := sc.Exists("3"); !yes {
		t.Fatal("Exists does not work as expected")
	}
	if yes := sc.Exists("rex: 3"); !yes {
		t.Fatal("Exists does not work as expected")
	}
}

//gocyclo:ignore
func TestScavenger_FindSequence(t *testing.T) {
	var sc = NewScavenger()
	sc.Debug("hello 1")
	sc.Debug()
	sc.Info("it is a good day to die")
	sc.Warn("3", "world 2")
	sc.Error("foo bar")
	sc.Error("")

	a1 := []string{
		"hello",
		"world",
	}
	if ret, ok := sc.Finder().FindStringSequence(a1); !ok || len(ret) != len(a1) {
		t.Fatal("FindStringSequence does not work as expected")
	} else if ret[0] != 0 || ret[1] != 3 {
		t.Fatal("FindStringSequence does not work as expected")
	} else if sc.LogEntry(ret[0]).Message != "hello 1" || sc.LogEntry(ret[1]).Message != "3world 2" {
		t.Fatal("FindStringSequence does not work as expected")
	}

	a2 := []string{
		"world",
		"hello",
	}
	if ret, ok := sc.Finder().FindStringSequence(a2); ok || len(ret) != 1 {
		t.Fatal("FindStringSequence does not work as expected")
	}

	a3 := []string{
		"hello",
		"",
		"world",
	}
	if ret, ok := sc.Finder().FindStringSequence(a3); !ok || len(ret) != len(a3) {
		t.Fatal("FindStringSequence does not work as expected")
	}

	a4 := []string{
		"hello",
		"world",
		"",
	}
	if ret, ok := sc.Finder().FindStringSequence(a4); !ok || len(ret) != len(a4) {
		t.Fatal("FindStringSequence does not work as expected")
	}

	b1 := []string{
		"hello \\d+",
		"it is a good.+",
	}
	if ret, ok := sc.Finder().FindRegexpSequence(b1); !ok || len(ret) != len(b1) {
		t.Fatal("FindRegexpSequence does not work as expected")
	} else if ret[0] != 0 || ret[1] != 2 {
		t.Fatal("FindRegexpSequence does not work as expected")
	}

	b2 := []string{
		"hello \\d+",
		"fo+ bar",
		"it is a good.+",
	}
	if ret, ok := sc.Finder().FindRegexpSequence(b2); ok || len(ret) != 2 {
		t.Fatal("FindRegexpSequence does not work as expected")
	}

	b3 := []string{
		"hello \\d+",
		"",
		"it is a good.+",
	}
	if ret, ok := sc.Finder().FindRegexpSequence(b3); !ok || len(ret) != len(b3) {
		t.Fatal("FindRegexpSequence does not work as expected")
	}

	b4 := append(b3, "")
	if ret, ok := sc.Finder().FindRegexpSequence(b4); !ok || len(ret) != len(b4) {
		t.Fatal("FindRegexpSequence does not work as expected")
	}

	c1 := []string{
		"rex: hello \\d+",
		"it is a good day",
	}
	if yes := sc.SequenceExists(c1); !yes {
		t.Fatal("SequenceExists does not work as expected")
	}
	if ret, ok := sc.Finder().FindSequence(c1); !ok || len(ret) != len(c1) {
		t.Fatal("FindSequence does not work as expected")
	}

	c2 := []string{
		"rex: hello \\d+",
		"rex: fo+ bar",
		"it is a good day",
	}
	if yes := sc.SequenceExists(c2); yes {
		t.Fatal("SequenceExists does not work as expected")
	}
	if ret, ok := sc.Finder().FindSequence(c2); ok || len(ret) != 2 {
		t.Fatal("FindSequence does not work as expected")
	}

	c3 := []string{
		"rex:hello \\d+",
		"it is a good day",
		"rex:  fo+ bar",
	}
	if yes := sc.SequenceExists(c3); !yes {
		t.Fatal("SequenceExists does not work as expected")
	}
	if ret, ok := sc.Finder().FindSequence(c3); !ok || len(ret) != len(c3) {
		t.Fatal("FindSequence does not work as expected")
	}

	c4 := []string{
		"rex: hello \\d+",
		"",
		"it is a good day",
	}
	if yes := sc.SequenceExists(c4); !yes {
		t.Fatal("SequenceExists does not work as expected")
	}
	if ret, ok := sc.Finder().FindSequence(c4); !ok || len(ret) != len(c4) {
		t.Fatal("FindSequence does not work as expected")
	}

	c5 := append(c4, "rex: ")
	if yes := sc.SequenceExists(c5); !yes {
		t.Fatal("SequenceExists does not work as expected")
	}
	if ret, ok := sc.Finder().FindSequence(c5); !ok || len(ret) != len(c5) {
		t.Fatal("FindSequence does not work as expected")
	}

	c6 := []string{
		"hello ",
		"it is a good day",
	}
	if yes := sc.SequenceExists(c6); !yes {
		t.Fatal("SequenceExists does not work as expected")
	}
	if ret, ok := sc.Finder().FindSequence(c6); !ok || len(ret) != len(c6) {
		t.Fatal("FindSequence does not work as expected")
	}
}

func TestScavenger_Entries(t *testing.T) {
	var sc = NewScavenger()
	sc.Debug("hello 1")
	sc.Info("it is a good day to die")
	sc.Warn("3", "world 2")
	sc.Error("foo bar")

	a := sc.Entries()
	if len(a) != len(sc.entries) {
		t.Fatal("len(a) != len(sc.entries)")
	}
	for i, e := range sc.entries {
		if a[i].Level != e.Level || a[i].Message != e.Message {
			t.Fatalf("a[i].Level != e.Level || a[i].Message != e.Message. i: %d", i)
		}
	}
}

func TestScavenger_Filter(t *testing.T) {
	var sc = NewScavenger()
	sc.Debugw(fmt.Sprintf("%d", 1))
	sc.Infow(fmt.Sprintf("%s", "it is a good day to die"))
	sc.Warnw(fmt.Sprintf("%d, %s", 3, "c"))
	sc.Errorw(fmt.Sprintf("%d", 4), "foo", 100)
	sc.Warnw(fmt.Sprintf("%s", "it is a good day to die"), "%d\nsecond")
	sc.Errorw(fmt.Sprintf("%d", 1), "foo", 100, "bar", "qux")

	newScav := sc.Filter(func(level, msg string) bool {
		switch msg {
		case "1":
			return level == LevelError
		case "it is a good day to die":
			return level == LevelInfo
		default:
			return true
		}
	})

	dump := `INFO	it is a good day to die
WARN	3, c
ERROR	4	{"foo": 100}
ERROR	Ignored key without a value.	{"ignored": "%d\nsecond"}
ERROR	1	{"foo": 100, "bar": "qux"}
`

	if newScav.Dump() != dump {
		// Edwin
		fmt.Println(newScav.Dump())
		t.Fatal("something is wrong with Dump")
	}

	n := sc.Len()
	newScav.Error("plus")
	if sc.Len() != n {
		t.Fatal(`sc.Len() != n`)
	}
}

func TestScavenger_NewLoggerWith(t *testing.T) {
	sc1 := NewScavenger()
	sc2 := sc1.NewLoggerWith("hello", "world", "x1", math.MaxInt64).(*Scavenger)
	sc3 := sc2.NewLoggerWith("hello", "world", "x2", math.MaxInt64).(*Scavenger)
	sc3.Debug("it is a good day to die")
	sc3.Infow("it is a good day to die")
	sc3.Warnw("it is a good day to die", "bar", 100)

	dump := `DEBUG	it is a good day to die	{"hello": "world", "x1": 9223372036854775807, "hello": "world", "x2": 9223372036854775807}
INFO	it is a good day to die	{"hello": "world", "x1": 9223372036854775807, "hello": "world", "x2": 9223372036854775807}
WARN	it is a good day to die	{"hello": "world", "x1": 9223372036854775807, "hello": "world", "x2": 9223372036854775807, "bar": 100}
`
	if sc3.Dump() != dump {
		t.Fatal("something is wrong [sc3]")
	}
	if sc1.Dump() != dump {
		t.Fatal("something is wrong [sc1]")
	}

	if sc3.Len() != sc1.Len() {
		t.Fatal(`sc3.Len() != sc1.Len()`)
	}
}

func TestScavenger_Race(t *testing.T) {
	var sc = NewScavenger()
	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		i := i
		scav := sc
		if i%2 == 0 {
			scav = sc.NewLoggerWith("abc", "x").(*Scavenger)
		}
		go func() {
			defer wg.Done()
			for j := 0; j < 100; j++ {
				scav.Infow("hello", "foo", i*100+j)
			}
		}()
	}

	wg.Wait()
	if sc.Len() != 10000 {
		t.Fatal(`sc.Len() != 10000`)
	}

	sc.mu.Lock()
	defer sc.mu.Unlock()
	rex1 := regexp.MustCompile(`^hello` + "\t" + `\{"abc": "x", "foo": \d+\}$`)
	rex2 := regexp.MustCompile(`^hello` + "\t" + `\{"foo": \d+\}$`)
	for _, e := range sc.entries {
		if !rex1.MatchString(e.Message) && !rex2.MatchString(e.Message) {
			t.Fatal(`Scavenger is not thread-safe`)
		}
	}
}

func TestScavenger_Multiline(t *testing.T) {
	var sc = NewScavenger()
	sc.Debugf("full stack\n%s", debug.Stack())
	if sc.Len() != 1 {
		t.Fatal(`c.Len() != 1`)
	}

	sc.Debugw("it is a good day to die", "%d\nsecond")
	if sc.Len() != 3 {
		t.Fatal(`sc.Len() != 3`)
	}

	sc.Debugw("it is a good day to die", []int{123}, "fire", []int{456})
	if sc.Len() != 6 {
		t.Fatal(`sc.Len() != 6`)
	}

	var count int
	sc.Filter(func(level, msg string) bool {
		if level == LevelError {
			if strings.HasPrefix(msg, string(_oddNumberErrMsg)) {
				count++
			} else if strings.HasPrefix(msg, string(_nonStringKeyErrMsg)) {
				count++
			}
		}
		return true
	})
	if count != 3 {
		t.Fatal(`count != 3`)
	}
}
