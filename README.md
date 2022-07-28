# slog
A collection of handy log utilities, including `ConsoleLogger`, `Scavenger`, `ZapLogger` and `DumbLogger`.

Each of them implements the following interface:

``` go
type Logger interface {
	NewLoggerWith(keyVals ...any) Logger

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

	FlushLogger() error
}
```

I love `Scavenger` the most. `Scavenger` collects all log messages for later queries. It makes designing complex test cases much easier.

``` go
func NewScavenger(printers ...Printer) *Scavenger

func (sc *Scavenger) StringExists(str string) (yes bool)
func (sc *Scavenger) UniqueStringExists(str string) (yes bool)
func (sc *Scavenger) FindStringSequence(seq []string) (found int, yes bool)
func (sc *Scavenger) RegexpExists(pat string) (yes bool)
func (sc *Scavenger) UniqueRegexpExists(pat string) (yes bool)
func (sc *Scavenger) FindRegexpSequence(seq []string) (found int, yes bool)
func (sc *Scavenger) Exists(str string) (yes bool)
func (sc *Scavenger) UniqueExists(str string) (yes bool)
func (sc *Scavenger) FindSequence(seq []string) (found int, yes bool)

func (sc *Scavenger) Dump() string
func (sc *Scavenger) Entries() []LogEntry
func (sc *Scavenger) Filter(fn func(level, msg string) bool) *Scavenger
func (sc *Scavenger) Len() int
func (sc *Scavenger) Reset()

func (sc *Scavenger) Debug(args ...any)
func (sc *Scavenger) Debugf(format string, args ...any)
func (sc *Scavenger) Debugw(msg string, keyVals ...any)
func (sc *Scavenger) Info(args ...any)
func (sc *Scavenger) Infof(format string, args ...any)
func (sc *Scavenger) Infow(msg string, keyVals ...any)
func (sc *Scavenger) Warn(args ...any)
func (sc *Scavenger) Warnf(format string, args ...any)
func (sc *Scavenger) Warnw(msg string, keyVals ...any)
func (sc *Scavenger) Error(args ...any)
func (sc *Scavenger) Errorf(format string, args ...any)
func (sc *Scavenger) Errorw(msg string, keyVals ...any)

func (sc *Scavenger) NewLoggerWith(keyVals ...any) Logger
func (sc *Scavenger) FlushLogger() error
```
