# slog
A collection of handy log utilities, including `ConsoleLogger`, `Scavenger`, `ZapLogger` and `DumbLogger`.

Each of them implements the following interface:

``` go
type Logger interface {
	NewLoggerWith(keyVals ...interface{}) Logger

	Debug(args ...interface{})
	Info(args ...interface{})
	Warn(args ...interface{})
	Error(args ...interface{})

	Debugf(format string, args ...interface{})
	Infof(format string, args ...interface{})
	Warnf(format string, args ...interface{})
	Errorf(format string, args ...interface{})

	Debugw(msg string, keyVals ...interface{})
	Infow(msg string, keyVals ...interface{})
	Warnw(msg string, keyVals ...interface{})
	Errorw(msg string, keyVals ...interface{})

	FlushLogger() error
}
```

I love `Scavenger` the most. `Scavenger` collects all log messages for later queries. It makes designing complex test cases much easier.

``` go
func (sc *Scavenger) StringExists(str string) (yes bool)
func (sc *Scavenger) StringSequenceExists(a []string) (yes bool)
func (sc *Scavenger) UniqueStringExists(str string) (yes bool)
func (sc *Scavenger) RegexpExists(pat string) (yes bool)
func (sc *Scavenger) RegexpSequenceExists(a []string) (yes bool)
func (sc *Scavenger) UniqueRegexpExists(pat string) (yes bool)
func (sc *Scavenger) Exists(str string) (yes bool)
func (sc *Scavenger) SequenceExists(a []string) (yes bool)
func (sc *Scavenger) UniqueExists(str string) (yes bool)

func (sc *Scavenger) Dump() string
func (sc *Scavenger) Entries() []LogEntry
func (sc *Scavenger) Filter(fn func(level, msg string) bool) *Scavenger
func (sc *Scavenger) Len() int
func (sc *Scavenger) Reset()

func (sc *Scavenger) Debug(args ...interface{})
func (sc *Scavenger) Debugf(format string, args ...interface{})
func (sc *Scavenger) Debugw(msg string, keyVals ...interface{})
func (sc *Scavenger) Info(args ...interface{})
func (sc *Scavenger) Infof(format string, args ...interface{})
func (sc *Scavenger) Infow(msg string, keyVals ...interface{})
func (sc *Scavenger) Warn(args ...interface{})
func (sc *Scavenger) Warnf(format string, args ...interface{})
func (sc *Scavenger) Warnw(msg string, keyVals ...interface{})
func (sc *Scavenger) Error(args ...interface{})
func (sc *Scavenger) Errorf(format string, args ...interface{})
func (sc *Scavenger) Errorw(msg string, keyVals ...interface{})

func (sc *Scavenger) NewLoggerWith(keyVals ...interface{}) Logger
func (sc *Scavenger) FlushLogger() error
```