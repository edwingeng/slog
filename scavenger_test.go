package slog

import (
	"errors"
	"fmt"
	"log"
	"math"
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

	if yes := sc.StringExists(""); yes {
		t.Fatal("StringExists does not work as expected")
	}
	if yes := sc.StringExists("5"); yes {
		t.Fatal("StringExists does not work as expected")
	}
	if yes := sc.StringExists("3"); !yes {
		t.Fatal("StringExists does not work as expected")
	}

	if yes := sc.RegexpExists(""); yes {
		t.Fatal("RegexpExists does not work as expected")
	}
	if yes := sc.RegexpExists("5"); yes {
		t.Fatal("RegexpExists does not work as expected")
	}
	if yes := sc.RegexpExists("g.+?d"); !yes {
		t.Fatal("RegexpExists does not work as expected")
	}
	if yes := sc.RegexpExists("^.+good.+die$"); !yes {
		t.Fatal("RegexpExists does not work as expected")
	}

	if yes := sc.Exists("5"); yes {
		t.Fatal("Exists does not work as expected")
	}
	if yes := sc.Exists("rex: 5"); yes {
		t.Fatal("Exists does not work as expected")
	}

	sc.Debug()
	if yes := sc.StringExists(""); !yes {
		t.Fatal("StringExists does not work as expected")
	}
	if yes := sc.RegexpExists(""); !yes {
		t.Fatal("RegexpExists does not work as expected")
	}

	sc.Reset()
	if len(sc.entries) > 0 {
		t.Fatal("Scavenger should be empty now")
	}
}

func TestScavenger_RegexpExists_Panic(t *testing.T) {
	defer func() {
		_ = recover()
	}()
	var sc Scavenger
	sc.RegexpExists("[")
	t.Fatal("RegexpExists should panic")
}

func TestScavenger_UniqueExists(t *testing.T) {
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

	if yes := sc.UniqueStringExists("1"); yes {
		t.Fatal("UniqueStringExists does not work as expected")
	}
	if yes := sc.UniqueStringExists("it is a good day to die"); yes {
		t.Fatal("UniqueStringExists does not work as expected")
	}
	if yes := sc.UniqueStringExists("3"); !yes {
		t.Fatal("UniqueStringExists does not work as expected")
	}
	if yes := sc.UniqueStringExists(""); !yes {
		t.Fatal("UniqueStringExists does not work as expected")
	}

	if yes := sc.UniqueRegexpExists("1"); yes {
		t.Fatal("UniqueRegexpExists does not work as expected")
	}
	if yes := sc.UniqueRegexpExists("it is a good day to die"); yes {
		t.Fatal("UniqueRegexpExists does not work as expected")
	}
	if yes := sc.UniqueRegexpExists("3"); !yes {
		t.Fatal("UniqueRegexpExists does not work as expected")
	}
	if yes := sc.UniqueRegexpExists("[3,4]"); yes {
		t.Fatal("UniqueRegexpExists does not work as expected")
	}

	sc.Debug("")
	if yes := sc.UniqueStringExists(""); yes {
		t.Fatal("UniqueStringExists does not work as expected")
	}
	if yes := sc.UniqueRegexpExists(""); yes {
		t.Fatal("UniqueRegexpExists does not work as expected")
	}

	if yes := sc.UniqueExists("3"); !yes {
		t.Fatal("UniqueExists does not work as expected")
	}
	if yes := sc.UniqueExists("rex: 3"); !yes {
		t.Fatal("UniqueExists does not work as expected")
	}
}

func TestScavenger_SequenceExists(t *testing.T) {
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
	if yes := sc.StringSequenceExists(a1); !yes {
		t.Fatal("StringSequenceExists does not work as expected")
	}

	a2 := []string{
		"world",
		"hello",
	}
	if yes := sc.StringSequenceExists(a2); yes {
		t.Fatal("StringSequenceExists does not work as expected")
	}

	a3 := []string{
		"hello",
		"",
		"world",
	}
	if yes := sc.StringSequenceExists(a3); !yes {
		t.Fatal("StringSequenceExists does not work as expected")
	}

	a4 := []string{
		"hello",
		"world",
		"",
	}
	if yes := sc.StringSequenceExists(a4); !yes {
		t.Fatal("StringSequenceExists does not work as expected")
	}

	b1 := []string{
		"hello \\d+",
		"it is a good.+",
	}
	if yes := sc.RegexpSequenceExists(b1); !yes {
		t.Fatal("RegexpSequenceExists does not work as expected")
	}

	b2 := []string{
		"hello \\d+",
		"fo+ bar",
		"it is a good.+",
	}
	if yes := sc.RegexpSequenceExists(b2); yes {
		t.Fatal("RegexpSequenceExists does not work as expected")
	}

	b3 := []string{
		"hello \\d+",
		"",
		"it is a good.+",
	}
	if yes := sc.RegexpSequenceExists(b3); !yes {
		t.Fatal("RegexpSequenceExists does not work as expected")
	}

	b4 := append(b3, "")
	if yes := sc.RegexpSequenceExists(b4); !yes {
		t.Fatal("RegexpSequenceExists does not work as expected")
	}

	c1 := []string{
		"rex: hello \\d+",
		"it is a good day",
	}
	if yes := sc.SequenceExists(c1); !yes {
		t.Fatal("SequenceExists does not work as expected")
	}

	c2 := []string{
		"rex: hello \\d+",
		"rex: fo+ bar",
		"it is a good day",
	}
	if yes := sc.SequenceExists(c2); yes {
		t.Fatal("SequenceExists does not work as expected")
	}

	c3 := []string{
		"rex: hello \\d+",
		"it is a good day",
		"rex: fo+ bar",
	}
	if yes := sc.SequenceExists(c3); !yes {
		t.Fatal("SequenceExists does not work as expected")
	}

	c4 := []string{
		"rex: hello \\d+",
		"",
		"it is a good day",
	}
	if yes := sc.SequenceExists(c4); !yes {
		t.Fatal("SequenceExists does not work as expected")
	}

	c5 := append(c4, "rex: ")
	if yes := sc.SequenceExists(c5); !yes {
		t.Fatal("SequenceExists does not work as expected")
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
	sc.Warnw(fmt.Sprintf("%s", "it is a good day to die"), "dumb")
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
ERROR	1	{"foo": 100, "bar": "qux"}
`

	if newScav.Dump() != dump {
		t.Fatal("something is wrong with Dump")
	}
}

func TestScavenger_NewLoggerWith(t *testing.T) {
	var sc1 = NewScavenger()
	sc2 := sc1.NewLoggerWith("hello", "world", "x1", math.MaxInt64).(*Scavenger)
	sc3 := sc2.NewLoggerWith("hello", "world", "x2", math.MaxInt64).(*Scavenger)
	sc3.Debug("it is a good day to die")
	sc3.Infow("it is a good day to die")
	sc3.Warnw("it is a good day to die", "bar", 100)

	dump := `DEBUG	it is a good day to die	{"hello": "world", "x1": 9223372036854775807, "x2": 9223372036854775807}
INFO	it is a good day to die	{"hello": "world", "x1": 9223372036854775807, "x2": 9223372036854775807}
WARN	it is a good day to die	{"hello": "world", "x1": 9223372036854775807, "x2": 9223372036854775807, "bar": 100}
`
	if sc3.Dump() != dump {
		t.Fatal("something is wrong [sc3]")
	}
	if sc1.Dump() != dump {
		t.Fatal("something is wrong [sc1]")
	}
}

type fakePrinter struct {
	n   int
	err error
	msg string
}

func (fp *fakePrinter) Print(level, message string) {
	fp.msg = message
}

func (fp *fakePrinter) Sync() error {
	fp.n++
	return fp.err
}

func TestScavenger_FlushLogger(t *testing.T) {
	if err := NewScavenger().FlushLogger(); err != nil {
		t.Fatal(err)
	}

	fp1 := fakePrinter{
		err: errors.New("1"),
	}
	fp2 := fakePrinter{
		err: errors.New("2"),
	}

	sc := NewScavenger(&fp1, &fp2)
	sc.Debug("100")
	if fp1.msg != "100" {
		t.Fatal(`fp1.msg != "100"`)
	}
	if fp2.msg != "100" {
		t.Fatal(`fp2.msg != "100"`)
	}

	stdLog := sc.logger.stdLog
	sc.logger.stdLog = log.New(nil, "", 0)
	if err := sc.FlushLogger(); err == nil {
		t.Fatal(`err == nil`)
	} else if err.Error() != "nil writer" {
		t.Fatal(`err.Error() != "nil writer"`)
	}
	if fp1.n != 1 {
		t.Fatal(`fp1.n != 1`)
	}
	if fp2.n != 1 {
		t.Fatal(`fp2.n != 1`)
	}

	sc.logger.stdLog = stdLog
	if err := sc.FlushLogger(); err == nil {
		t.Fatal(`err == nil`)
	} else if err.Error() != "1" {
		t.Fatal(`err.Error() != "1"`)
	}
}
