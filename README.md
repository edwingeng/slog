# Overview
`slog` is a collection of handy log utilities, including `ZapLogger`, `Scavenger` and `Devourer`.

Each of them implements the following interface:

``` go
type Logger interface {
    NewLoggerWith(keyVals ...any) Logger
    LogLevelEnabled(level int) bool
    FlushLogger() error

    Debug(args ...any)
    Info(args ...any)
    Warn(args ...any)
    Error(args ...any)

    Debugf(format string, args ...any)
    Infof(format string, args ...any)
    Warnf(format string, args ...any)
    Errorf(format string, args ...any)

    Debugw(msg string, keyVals ...any)
    Infow(msg string, keyVals ...any)
    Warnw(msg string, keyVals ...any)
    Errorw(msg string, keyVals ...any)
}
```

# Getting Started
```
go get -u github.com/edwingeng/slog
```

# Usage

``` go
logger, err := NewDevelopmentConfig().Build()
if err != nil {
    panic(err)
}

logger.Debug("str1")
logger.Infof("str2: %d", 200)
logger.Warnw("str3", "foo", 300)

type HandlerContext struct {
    context.Context
    Logger
}

ctx := &HandlerContext{
    Context: context.TODO(),
    Logger:  logger.NewLoggerWith("handler", "UpdateUserName"),
}
ctx.Error("invalid user name")

// Output:
// 15|23:10:48     DEBUG   slog/example_test.go:11 str1
// 15|23:10:48     INFO    slog/example_test.go:12 str2: 200
// 15|23:10:48     WARN    slog/example_test.go:13 str3    {"foo": 300}
// 15|23:10:48     ERROR   slog/example_test.go:24 invalid user name       {"handler": "UpdateUserName"}
```

# Scavenger

I love `Scavenger` the most. `Scavenger` saves all log messages in memory for later use, which makes it much easier to design complex test cases.

``` go
func NewScavenger() *Scavenger

func (sc *Scavenger) Exists(str string) bool
func (sc *Scavenger) SequenceExists(seq []string) bool
func (sc *Scavenger) Finder() *MessageFinder

func (sc *Scavenger) Dump() string
func (sc *Scavenger) Entries() []LogEntry
func (sc *Scavenger) LogEntry(index int) LogEntry
func (sc *Scavenger) Filter(fn func(level, msg string) bool) *Scavenger
func (sc *Scavenger) Len() int
func (sc *Scavenger) Reset()

func (sc *Scavenger) NewLoggerWith(keyVals ...any) Logger
func (sc *Scavenger) LogLevelEnabled(level int) bool
func (sc *Scavenger) FlushLogger() error

func (sc *Scavenger) Debug(args ...any)
func (sc *Scavenger) Info(args ...any)
func (sc *Scavenger) Warn(args ...any)
func (sc *Scavenger) Error(args ...any)

func (sc *Scavenger) Debugf(format string, args ...any)
func (sc *Scavenger) Infof(format string, args ...any)
func (sc *Scavenger) Warnf(format string, args ...any)
func (sc *Scavenger) Errorf(format string, args ...any)

func (sc *Scavenger) Debugw(msg string, keyVals ...any)
func (sc *Scavenger) Infow(msg string, keyVals ...any)
func (sc *Scavenger) Warnw(msg string, keyVals ...any)
func (sc *Scavenger) Errorw(msg string, keyVals ...any)
```
