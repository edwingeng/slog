package slog

import (
	"testing"
)

func TestScavenger(t *testing.T) {
	var scav Scavenger
	scav.Debug("1")
	scav.Info("it is a good day to die")
	scav.Warn("3", "c")
	scav.Error("4")

	if scav.Len() != 4 {
		t.Fatal("scav.Len() != 4")
	}

	dump := `DEBUG	1
INFO	it is a good day to die
WARNING	3, c
ERROR	4
`
	if scav.Dump() != dump {
		t.Fatal("something is wrong with Dump")
	}

	if _, ok := scav.FindString(""); ok {
		t.Fatal("FindString does not work as expected")
	}
	if _, ok := scav.FindString("5"); ok {
		t.Fatal("FindString does not work as expected")
	}
	if _, ok := scav.FindString("3"); !ok {
		t.Fatal("FindString does not work as expected")
	}

	if _, ok := scav.FindRegexp(""); ok {
		t.Fatal("FindRegexp does not work as expected")
	}
	if _, ok := scav.FindRegexp("5"); ok {
		t.Fatal("FindRegexp does not work as expected")
	}
	if _, ok := scav.FindRegexp("g.+?d"); !ok {
		t.Fatal("FindRegexp does not work as expected")
	}
	if _, ok := scav.FindRegexp("^.+good.+die$"); !ok {
		t.Fatal("FindRegexp does not work as expected")
	}

	scav.Debug()
	if _, ok := scav.FindString(""); !ok {
		t.Fatal("FindString does not work as expected")
	}
	if _, ok := scav.FindRegexp(""); !ok {
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
	var scav Scavenger
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
	var scav Scavenger
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
