package slog

import (
	"bytes"
	"encoding/json"
	"go.uber.org/zap"
	"log"
	"regexp"
	"strings"
	"testing"
)

const timePattern = `\d{2}:\d{2}:\d{2} `

func rexMatch(t *testing.T, level, pattern string, data []byte) {
	t.Helper()
	var padding string
	if len(level) == 4 {
		padding = " "
	}
	full := "^" + timePattern + level + padding + " .*consoleLogger_test.go:\\d+\t" + pattern + "\n$"
	matched, err := regexp.Match(full, data)
	if err != nil {
		t.Fatal(err)
	}
	if !matched {
		t.Fatalf("%q does not match %q", full, data)
	}
}

func resetBuffer(buf *bytes.Buffer) {
	buf.Reset()
}

func TestConsoleLogger_Debug(t *testing.T) {
	var buf bytes.Buffer
	stdLog := log.New(&buf, "", log.Ltime)
	logger := NewConsoleLogger(WithStdLogger(stdLog), WithoutColor())
	logger.Debug()
	if buf.Len() != 0 {
		t.Fatal(`buf.Len() != 0`)
	}

	logger.Debug(100)
	var pat1 = "100"
	rexMatch(t, LevelDebug, pat1, buf.Bytes())

	resetBuffer(&buf)
	logger.Debug("")
	rexMatch(t, LevelDebug, "", buf.Bytes())

	resetBuffer(&buf)
	logger.Debug(100, 200)
	var pat2 = "100 200"
	rexMatch(t, LevelDebug, pat2, buf.Bytes())
}

func TestConsoleLogger_Infof(t *testing.T) {
	var buf bytes.Buffer
	stdLog := log.New(&buf, "", log.Ltime)
	logger := NewConsoleLogger(WithStdLogger(stdLog), WithoutColor())
	logger.Infof("%s %d", "foo", 100)
	var pat1 = "foo 100"
	rexMatch(t, LevelInfo, pat1, buf.Bytes())
}

func TestConsoleLogger_Warnw(t *testing.T) {
	var buf bytes.Buffer
	stdLog := log.New(&buf, "", log.Ltime)
	logger := NewConsoleLogger(WithStdLogger(stdLog), WithoutColor())
	logger.Warnw("hello")
	var pat1 = "hello"
	rexMatch(t, LevelWarn, pat1, buf.Bytes())

	resetBuffer(&buf)
	logger.Warnw("hello <world>", "foo", 100)
	var pat2 = "hello <world>\t{\"foo\": 100}"
	rexMatch(t, LevelWarn, pat2, buf.Bytes())

	resetBuffer(&buf)
	logger.Warnw("hello", "foo", 100, "bar", "qux")
	var pat3 = "hello\t{\"foo\": 100, \"bar\": \"qux\"}"
	rexMatch(t, LevelWarn, pat3, buf.Bytes())

	resetBuffer(&buf)
	logger.Warnw("")
	var pat4 = ""
	rexMatch(t, LevelWarn, pat4, buf.Bytes())

	resetBuffer(&buf)
	logger.Warnw("", zap.Int("foo", 100))
	var pat5 = "{\"foo\": 100}"
	rexMatch(t, LevelWarn, pat5, buf.Bytes())

	resetBuffer(&buf)
	logger.Warnw("hello", "foo", 100, zap.String("name", "ariel"), "bar", "qux")
	var pat6 = "hello\t{\"foo\": 100, \"name\": \"ariel\", \"bar\": \"qux\"}"
	rexMatch(t, LevelWarn, pat6, buf.Bytes())

	var obj struct {
		Num int
		Str string
	}
	obj.Num = 100
	obj.Str = "world"
	resetBuffer(&buf)
	logger.Warnw("hello", "foo", 100, "obj", obj)
	var pat7 = "hello\t{\"foo\": 100, \"obj\": {\"Num\":100,\"Str\":\"world\"}}"
	rexMatch(t, LevelWarn, pat7, buf.Bytes())

	resetBuffer(&buf)
	logger.Warnw("hello", "foo", 100, zap.String("name", "ariel"), "bar")
	var pat8 = "hello\t{\"foo\": 100, \"name\": \"ariel\"}"
	rexMatch(t, LevelWarn, pat8, buf.Bytes())
}

func TestConsoleLogger_Print(t *testing.T) {
	var buf bytes.Buffer
	stdLog := log.New(&buf, "", log.Ltime)
	logger := NewConsoleLogger(WithStdLogger(stdLog), WithoutColor())
	logger.Print(LevelError, "100")
	var pat1 = "100"
	rexMatch(t, LevelError, pat1, buf.Bytes())
}

func TestWithFields(t *testing.T) {
	var buf bytes.Buffer
	stdLog := log.New(&buf, "", log.Ltime)
	fields := []any{zap.String("name", "ariel"), "foo", 100, "bar"}
	logger := NewConsoleLogger(WithStdLogger(stdLog), WithFields(fields...), WithoutColor())
	logger.Error("hello")
	var pat1 = "hello\t{\"foo\": 100, \"name\": \"ariel\"}"
	rexMatch(t, LevelError, pat1, buf.Bytes())
	resetBuffer(&buf)
	logger.Errorf("%s", "hello")
	rexMatch(t, LevelError, pat1, buf.Bytes())
	resetBuffer(&buf)
	logger.Errorw("hello")
	rexMatch(t, LevelError, pat1, buf.Bytes())

	resetBuffer(&buf)
	logger.Errorw("hello", "bar", "qux", "spare")
	var pat2 = "hello\t{\"foo\": 100, \"name\": \"ariel\", \"bar\": \"qux\"}"
	rexMatch(t, LevelError, pat2, buf.Bytes())

	resetBuffer(&buf)
	logger.Error("")
	var patx = "{\"foo\": 100, \"name\": \"ariel\"}"
	rexMatch(t, LevelError, patx, buf.Bytes())

	resetBuffer(&buf)
	x3 := NewConsoleLogger(WithStdLogger(stdLog), WithFields("one"), WithoutColor())
	x3.Error("hello")
	var pat3 = "hello"
	rexMatch(t, LevelError, pat3, buf.Bytes())
}

func TestWithExtraCallerSkip(t *testing.T) {
	var buf bytes.Buffer
	stdLog := log.New(&buf, "", 0)
	logger := NewConsoleLogger(WithStdLogger(stdLog), WithExtraCallerSkip(1), WithoutColor())
	logger.Infof("hello %s", "world")
	pat1 := LevelInfo + "  .*testing.go:\\d+\thello world"
	matched, err := regexp.Match(pat1, buf.Bytes())
	if err != nil {
		t.Fatal(err)
	}
	if !matched {
		t.Fatalf("%q does not match %q", pat1, buf.Bytes())
	}
}

func TestWithLevel(t *testing.T) {
	var buf bytes.Buffer
	stdLog := log.New(&buf, "", 0)
	printMessages := func(logger *ConsoleLogger, group int) {
		switch group {
		case 0:
			logger.Debug("1")
			logger.Info("2")
			logger.Warn("3")
			logger.Error("4")
		case 1:
			logger.Debugf("%d", 1)
			logger.Infof("%d", 2)
			logger.Warnf("%d", 3)
			logger.Errorf("%d", 4)
		case 2:
			logger.Debugw("1")
			logger.Infow("2")
			logger.Warnw("3")
			logger.Errorw("4")
		case 3:
			logger.Print(LevelDebug, "1")
			logger.Print(LevelInfo, "2")
			logger.Print(LevelWarn, "3")
			logger.Print(LevelError, "4")
		default:
			panic("impossible")
		}
	}

	x1 := NewConsoleLogger(WithStdLogger(stdLog), WithLevel(LevelDebug), WithoutColor())
	for i := 0; i < 4; i++ {
		resetBuffer(&buf)
		printMessages(x1, i)
		if strings.Count(buf.String(), "\n") != 4 {
			t.Fatalf("not 4, i: %d", i)
		}
	}

	x2 := NewConsoleLogger(WithStdLogger(stdLog), WithLevel(LevelInfo), WithoutColor())
	for i := 0; i < 4; i++ {
		resetBuffer(&buf)
		printMessages(x2, i)
		if strings.Count(buf.String(), "\n") != 3 {
			t.Fatalf("not 3, i: %d", i)
		}
	}

	x3 := NewConsoleLogger(WithStdLogger(stdLog), WithLevel(LevelWarn), WithoutColor())
	for i := 0; i < 4; i++ {
		resetBuffer(&buf)
		printMessages(x3, i)
		if strings.Count(buf.String(), "\n") != 2 {
			t.Fatalf("not 2, i: %d", i)
		}
	}

	x4 := NewConsoleLogger(WithStdLogger(stdLog), WithLevel(LevelError), WithoutColor())
	for i := 0; i < 4; i++ {
		resetBuffer(&buf)
		printMessages(x4, i)
		if strings.Count(buf.String(), "\n") != 1 {
			t.Fatalf("not 1, i: %d", i)
		}
	}

	func() {
		defer func() {
			if recover() == nil {
				t.Fatal(`recover() == nil`)
			}
		}()
		NewConsoleLogger(WithLevel("x"))
	}()
}

func TestConsoleLogger_NewLoggerWith(t *testing.T) {
	var buf bytes.Buffer
	stdLog := log.New(&buf, "", log.Ltime)
	x1 := NewConsoleLogger(WithStdLogger(stdLog), WithoutColor())
	x2 := x1.NewLoggerWith("foo", 100)
	x1.Info("hello")
	pat1 := "hello"
	rexMatch(t, LevelInfo, pat1, buf.Bytes())
	resetBuffer(&buf)
	x2.Info("hello")
	pat2 := "hello\t{\"foo\": 100}"
	rexMatch(t, LevelInfo, pat2, buf.Bytes())

	if x2.NewLoggerWith().(*ConsoleLogger).fields != x2.(*ConsoleLogger).fields {
		t.Fatal(`x2.NewLoggerWith().(*ConsoleLogger).fields != x2.(*ConsoleLogger).fields`)
	}

	resetBuffer(&buf)
	fields := []any{zap.String("name", "ariel"), "bar", "qux", "xxx"}
	x3 := NewConsoleLogger(WithStdLogger(stdLog), WithFields(fields...), WithoutColor())
	x4 := x3.NewLoggerWith("foo", 100, "foo2", "world")
	x3.Info("hello")
	pat3 := "hello\t{\"bar\": \"qux\", \"name\": \"ariel\"}"
	rexMatch(t, LevelInfo, pat3, buf.Bytes())
	resetBuffer(&buf)
	x4.Info("hello")
	pat4 := "hello\t{\"bar\": \"qux\", \"foo\": 100, \"foo2\": \"world\", \"name\": \"ariel\"}"
	rexMatch(t, LevelInfo, pat4, buf.Bytes())
}

func TestConsoleLogger_bare(t *testing.T) {
	var buf bytes.Buffer
	stdLog := log.New(&buf, "", log.LstdFlags)
	x1 := NewConsoleLogger(WithStdLogger(stdLog), WithBareMode())

	x1.Debug("hello")
	if buf.String() != "hello\n" {
		t.Fatal("bare: unexpected output")
	}

	buf.Reset()
	x1.Infof("hello %d", 100)
	if buf.String() != "hello 100\n" {
		t.Fatal("bare: unexpected output")
	}

	buf.Reset()
	x1.Warnw("hello")
	if buf.String() != "hello\n" {
		t.Fatal("bare: unexpected output")
	}

	buf.Reset()
	x1.Errorw("hello", "foo", 100)
	if buf.String() != "hello\t{\"foo\": 100}\n" {
		t.Fatal("bare: unexpected output")
	}
}

func TestConsoleLogger_Colored(t *testing.T) {
	var buf bytes.Buffer
	stdLog := log.New(&buf, "", log.Ltime)
	logger := NewConsoleLogger(WithStdLogger(stdLog))
	logger.Debug("100")
	logger.Info("200")
	logger.Warn("300")
	logger.Error("400")
	matched, err := regexp.Match("(?s)^.+100.+200.+300.+400.+$", buf.Bytes())
	if err != nil {
		t.Fatal(err)
	}
	if !matched {
		t.Fatal("failed to match")
	}
}

func TestStringlize(t *testing.T) {
	m := map[any]string{
		"str\ning":               `"str\ning"`,
		json.Number("100"):       "100",
		json.Number("3.1415926"): "3.1415926",
		100:                      "100",
		uint(100):                "100",
		3.1415926:                "3.1415926",
		true:                     "true",
		uintptr(100):             "0x64",
	}
	for k, v := range m {
		str := stringlize(k)
		if str != v {
			t.Fatalf("unexpected result. k: %v, v: %v, str: %s", k, v, str)
		}
	}

	type circle struct {
		Ptr *circle
	}

	var c circle
	c.Ptr = &c
	matched, err := regexp.MatchString(`^&{Ptr:.+}$`, stringlize(&c))
	if err != nil {
		t.Fatal(err)
	} else if !matched {
		t.Fatal("something is wrong")
	}
}

type fakeWriter struct {
	n int
}

func (*fakeWriter) Write(p []byte) (n int, err error) {
	return
}

func (fw *fakeWriter) Sync() error {
	fw.n++
	return nil
}

func TestConsoleLogger_FlushLogger(t *testing.T) {
	stdLog := log.New(nil, "", log.Ltime)
	logger := NewConsoleLogger(WithStdLogger(stdLog))
	if err := logger.FlushLogger(); err == nil {
		t.Fatal(`err == nil`)
	} else if err.Error() != "nil writer" {
		t.Fatal(`err.Error() != "nil writer"`)
	}

	var fw fakeWriter
	x2a := log.New(&fw, "", log.Ltime)
	x2b := NewConsoleLogger(WithStdLogger(x2a))
	if err := x2b.FlushLogger(); err != nil {
		t.Fatal(err)
	}
	if fw.n != 1 {
		t.Fatal(`fw.n != 1`)
	}

	x3a := log.New(&bytes.Buffer{}, "", log.Ltime)
	x3b := NewConsoleLogger(WithStdLogger(x3a))
	if err := x3b.FlushLogger(); err != nil {
		t.Fatal(err)
	}
}
