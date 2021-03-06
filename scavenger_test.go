package slog

import (
	"fmt"
	"testing"
)

func TestScavenger(t *testing.T) {
	var scav = NewScavenger()
	scav.Debug("1")
	scav.Info("it is a good day to die")
	scav.Warn("3", "c")
	scav.Error("4")

	if scav.Len() != 4 {
		t.Fatal("scav.Len() != 4")
	}

	dump := `DEBUG	1
INFO	it is a good day to die
WARN	3c
ERROR	4
`
	if scav.Dump() != dump {
		t.Fatal("something is wrong with Dump")
	}

	if _, _, ok := scav.FindString(""); ok {
		t.Fatal("FindString does not work as expected")
	}
	if _, _, ok := scav.FindString("5"); ok {
		t.Fatal("FindString does not work as expected")
	}
	if _, _, ok := scav.FindString("3"); !ok {
		t.Fatal("FindString does not work as expected")
	}

	if _, _, ok := scav.FindRegexp(""); ok {
		t.Fatal("FindRegexp does not work as expected")
	}
	if _, _, ok := scav.FindRegexp("5"); ok {
		t.Fatal("FindRegexp does not work as expected")
	}
	if _, _, ok := scav.FindRegexp("g.+?d"); !ok {
		t.Fatal("FindRegexp does not work as expected")
	}
	if _, _, ok := scav.FindRegexp("^.+good.+die$"); !ok {
		t.Fatal("FindRegexp does not work as expected")
	}

	scav.Debug()
	if _, _, ok := scav.FindString(""); !ok {
		t.Fatal("FindString does not work as expected")
	}
	if _, _, ok := scav.FindRegexp(""); !ok {
		t.Fatal("FindRegexp does not work as expected")
	}

	scav.Reset()
	if len(scav.entries) > 0 {
		t.Fatal("Scavenger should be empty now")
	}
}

func TestScavenger_FindRegexp_Panic(t *testing.T) {
	defer func() {
		_ = recover()
	}()
	var log Scavenger
	log.FindRegexp("[")
	t.Fatal("FindRegexp should panic")
}

func TestScavenger_FindSequence(t *testing.T) {
	var scav = NewScavenger()
	scav.Debug("hello 1")
	scav.Debug()
	scav.Info("it is a good day to die")
	scav.Warn("3", "world 2")
	scav.Error("foo bar")
	scav.Error("")

	a1 := []string{
		"hello",
		"world",
	}
	if _, ok := scav.FindStringSequence(a1); !ok {
		t.Fatal("FindStringSequence does not work as expected")
	}

	a2 := []string{
		"world",
		"hello",
	}
	if _, ok := scav.FindStringSequence(a2); ok {
		t.Fatal("FindStringSequence does not work as expected")
	}

	a3 := []string{
		"hello",
		"",
		"world",
	}
	if _, ok := scav.FindStringSequence(a3); !ok {
		t.Fatal("FindStringSequence does not work as expected")
	}

	a4 := []string{
		"hello",
		"world",
		"",
	}
	if _, ok := scav.FindStringSequence(a4); !ok {
		t.Fatal("FindStringSequence does not work as expected")
	}

	b1 := []string{
		"hello \\d+",
		"it is a good.+",
	}
	if _, ok := scav.FindRegexpSequence(b1); !ok {
		t.Fatal("FindRegexpSequence does not work as expected")
	}

	b2 := []string{
		"hello \\d+",
		"fo+ bar",
		"it is a good.+",
	}
	if _, ok := scav.FindRegexpSequence(b2); ok {
		t.Fatal("FindRegexpSequence does not work as expected")
	}

	b3 := []string{
		"hello \\d+",
		"",
		"it is a good.+",
	}
	if _, ok := scav.FindRegexpSequence(b3); !ok {
		t.Fatal("FindRegexpSequence does not work as expected")
	}

	c1 := []string{
		"rex: hello \\d+",
		"it is a good day",
	}
	if _, ok := scav.FindSequence(c1); !ok {
		t.Fatal("FindSequence does not work as expected")
	}

	c2 := []string{
		"rex: hello \\d+",
		"rex: fo+ bar",
		"it is a good day",
	}
	if _, ok := scav.FindSequence(c2); ok {
		t.Fatal("FindSequence does not work as expected")
	}

	c3 := []string{
		"rex: hello \\d+",
		"it is a good day",
		"rex: fo+ bar",
	}
	if _, ok := scav.FindSequence(c3); !ok {
		t.Fatal("FindSequence does not work as expected")
	}

	c4 := []string{
		"rex: hello \\d+",
		"",
		"it is a good day",
	}
	if _, ok := scav.FindSequence(c4); !ok {
		t.Fatal("FindSequence does not work as expected")
	}
}

func TestScavenger_Entries(t *testing.T) {
	var scav = NewScavenger()
	scav.Debug("hello 1")
	scav.Info("it is a good day to die")
	scav.Warn("3", "world 2")
	scav.Error("foo bar")

	a := scav.Entries()
	if len(a) != len(scav.entries) {
		t.Fatal("len(a) != len(scav.entries)")
	}
	for i, e := range scav.entries {
		if a[i].Level != e.Level || a[i].Message != e.Message {
			t.Fatalf("a[i].Level != e.Level || a[i].Message != e.Message. i: %d", i)
		}
	}
}

func TestScavenger_FindUnique(t *testing.T) {
	var scav = NewScavenger()
	scav.Debugf("%d", 1)
	scav.Infof("%s", "it is a good day to die")
	scav.Warnf("%d, %s", 3, "c")
	scav.Errorf("%d", 4)
	scav.Warnf("%s", "it is a good day to die")
	scav.Errorf("%d", 1)

	if scav.Len() != 6 {
		t.Fatal("scav.Len() != 6", scav.Len())
	}

	dump := `DEBUG	1
INFO	it is a good day to die
WARN	3, c
ERROR	4
WARN	it is a good day to die
ERROR	1
`
	if scav.Dump() != dump {
		t.Fatal("something is wrong with Dump")
	}

	if _, _, ok := scav.FindUniqueString("1"); ok {
		t.Fatal("FindUniqueString does not work as expected")
	}
	if _, _, ok := scav.FindUniqueString("it is a good day to die"); ok {
		t.Fatal("FindUniqueString does not work as expected")
	}
	if _, _, ok := scav.FindUniqueString("3"); !ok {
		t.Fatal("FindUniqueString does not work as expected")
	}

	if _, _, ok := scav.FindUniqueRegexp("1"); ok {
		t.Fatal("FindUniqueRegexp does not work as expected")
	}
	if _, _, ok := scav.FindUniqueRegexp("it is a good day to die"); ok {
		t.Fatal("FindUniqueRegexp does not work as expected")
	}
	if _, _, ok := scav.FindUniqueRegexp("3"); !ok {
		t.Fatal("FindUniqueRegexp does not work as expected")
	}
	if _, _, ok := scav.FindUniqueRegexp("[3,4]"); ok {
		t.Fatal("FindUniqueRegexp does not work as expected")
	}
}

func TestScavenger_Filter(t *testing.T) {
	var scav = NewScavenger()
	scav.Debugw(fmt.Sprintf("%d", 1))
	scav.Infow(fmt.Sprintf("%s", "it is a good day to die"))
	scav.Warnw(fmt.Sprintf("%d, %s", 3, "c"))
	scav.Errorw(fmt.Sprintf("%d", 4), "foo", 100)
	scav.Warnw(fmt.Sprintf("%s", "it is a good day to die"), "dumb")
	scav.Errorw(fmt.Sprintf("%d", 1), "foo", 100, "bar", "qux")

	newScav := scav.Filter(func(level, msg string) bool {
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
	var scav1 = NewScavenger()
	scav2 := scav1.NewLoggerWith("hello", "world").(*Scavenger)
	scav2.Debug("it is a good day to die")
	scav2.Infow("it is a good day to die")
	scav2.Warnw("it is a good day to die", "bar", 100)

	dump := `DEBUG	it is a good day to die	{"hello": "world"}
INFO	it is a good day to die	{"hello": "world"}
WARN	it is a good day to die	{"hello": "world", "bar": 100}
`
	if scav2.Dump() != dump {
		t.Fatal("something is wrong [scav2]")
	}
	if scav1.Dump() != dump {
		t.Fatal("something is wrong [scav1]")
	}
}
