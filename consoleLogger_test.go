package slog

import (
	"bytes"
	"log"
	"regexp"
	"strings"
	"testing"
)

const timePattern = `\d{2}:\d{2}:\d{2} `

func rexMatch(t *testing.T, level, pattern string, data []byte) {
	t.Helper()
	full := "^" + timePattern + level + "\tconsoleLogger_test.go:\\d+\t" + pattern + "\n$"
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
	logger := NewConsoleLogger(WithStdLogger(stdLog))
	logger.Debug(100)
	var pat1 = "100"
	rexMatch(t, LevelDebug, pat1, buf.Bytes())

	resetBuffer(&buf)
	logger.Debug(100, 200)
	var pat2 = "100 200"
	rexMatch(t, LevelDebug, pat2, buf.Bytes())
}

func TestConsoleLogger_Infof(t *testing.T) {
	var buf bytes.Buffer
	stdLog := log.New(&buf, "", log.Ltime)
	logger := NewConsoleLogger(WithStdLogger(stdLog))
	logger.Infof("%s %d", "foo", 100)
	var pat1 = "foo 100"
	rexMatch(t, LevelInfo, pat1, buf.Bytes())
}

func TestConsoleLogger_Warnw(t *testing.T) {
	var buf bytes.Buffer
	stdLog := log.New(&buf, "", log.Ltime)
	logger := NewConsoleLogger(WithStdLogger(stdLog))
	logger.Warnw("hello")
	var pat1 = "hello"
	rexMatch(t, LevelWarn, pat1, buf.Bytes())

	resetBuffer(&buf)
	logger.Warnw("hello", "foo", 100)
	var pat2 = "hello\t{\"foo\":100}"
	rexMatch(t, LevelWarn, pat2, buf.Bytes())

	resetBuffer(&buf)
	logger.Warnw("hello", "foo", 100, "bar", "qux")
	var pat3 = "hello\t{\"foo\":100, \"bar\":\"qux\"}"
	rexMatch(t, LevelWarn, pat3, buf.Bytes())

	resetBuffer(&buf)
	logger.Warnw("")
	var pat4 = ""
	rexMatch(t, LevelWarn, pat4, buf.Bytes())
}

func TestWithFields(t *testing.T) {
	var buf bytes.Buffer
	stdLog := log.New(&buf, "", log.Ltime)
	logger := NewConsoleLogger(WithStdLogger(stdLog), WithFields("foo", 100, "bar"))
	logger.Error("hello")
	var pat1 = "hello\t{\"foo\":100}"
	rexMatch(t, LevelError, pat1, buf.Bytes())
	resetBuffer(&buf)
	logger.Errorf("%s", "hello")
	rexMatch(t, LevelError, pat1, buf.Bytes())
	resetBuffer(&buf)
	logger.Errorw("hello")
	rexMatch(t, LevelError, pat1, buf.Bytes())

	resetBuffer(&buf)
	logger.Errorw("hello", "bar", "qux", "spare")
	var pat2 = "hello\t{\"foo\":100, \"bar\":\"qux\"}"
	rexMatch(t, LevelError, pat2, buf.Bytes())
}

func TestWithExtraCallerSkip(t *testing.T) {
	var buf bytes.Buffer
	stdLog := log.New(&buf, "", 0)
	logger := NewConsoleLogger(WithStdLogger(stdLog), WithExtraCallerSkip(1))
	logger.Infof("hello %s", "world")
	pat1 := LevelInfo + "\ttesting.go:\\d+\thello world"
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
	logMessages := func(logger Logger) {
		logger.Debug("1")
		logger.Info("2")
		logger.Warn("3")
		logger.Error("4")
	}

	x1 := NewConsoleLogger(WithStdLogger(stdLog), WithLevel(LevelDebug))
	logMessages(x1)
	if strings.Count(buf.String(), "\n") != 4 {
		t.Fatal("not 4")
	}

	resetBuffer(&buf)
	x2 := NewConsoleLogger(WithStdLogger(stdLog), WithLevel(LevelInfo))
	logMessages(x2)
	if strings.Count(buf.String(), "\n") != 3 {
		t.Fatal("not 3")
	}

	resetBuffer(&buf)
	x3 := NewConsoleLogger(WithStdLogger(stdLog), WithLevel(LevelWarn))
	logMessages(x3)
	if strings.Count(buf.String(), "\n") != 2 {
		t.Fatal("not 2")
	}

	resetBuffer(&buf)
	x4 := NewConsoleLogger(WithStdLogger(stdLog), WithLevel(LevelError))
	logMessages(x4)
	if strings.Count(buf.String(), "\n") != 1 {
		t.Fatal("not 1")
	}
}

func TestConsoleLogger_NewLoggerWith(t *testing.T) {
	var buf bytes.Buffer
	stdLog := log.New(&buf, "", log.Ltime)
	x1 := NewConsoleLogger(WithStdLogger(stdLog))
	x2 := x1.NewLoggerWith("foo", 100)
	x1.Info("hello")
	pat1 := "hello"
	rexMatch(t, LevelInfo, pat1, buf.Bytes())
	resetBuffer(&buf)
	x2.Info("hello")
	pat2 := "hello\t{\"foo\":100}"
	rexMatch(t, LevelInfo, pat2, buf.Bytes())

	resetBuffer(&buf)
	x3 := NewConsoleLogger(WithStdLogger(stdLog), WithFields("bar", "qux"))
	x4 := x3.NewLoggerWith("foo", 100)
	x3.Info("hello")
	pat3 := "hello\t{\"bar\":\"qux\"}"
	rexMatch(t, LevelInfo, pat3, buf.Bytes())
	resetBuffer(&buf)
	x4.Info("hello")
	pat4 := "hello\t{\"bar\":\"qux\", \"foo\":100}"
	rexMatch(t, LevelInfo, pat4, buf.Bytes())
}
