package slog

const (
	LevelDebug = "DEBUG"
	LevelInfo  = "INFO"
	LevelWarn  = "WARN"
	LevelError = "ERROR"
)

// Logger is an interface which is highly compatible with zap.SugaredLogger.
type Logger interface {
	// NewLoggerWith adds a variadic number of fields to the logging context. It accepts a
	// mix of zap.Field objects and loosely-typed key-value pairs. When
	// processing pairs, the first element of the pair is used as the field key
	// and the second as the field value. The keys in key-value pairs should be strings.
	NewLoggerWith(keyVals ...any) Logger

	// LogLevelEnabled checks if the given log level is enabled.
	LogLevelEnabled(level int) bool

	// Debug uses fmt.Sprint to construct and log a message.
	Debug(args ...any)
	// Info uses fmt.Sprint to construct and log a message.
	Info(args ...any)
	// Warn uses fmt.Sprint to construct and log a message.
	Warn(args ...any)
	// Error uses fmt.Sprint to construct and log a message.
	Error(args ...any)

	// Debugf uses fmt.Sprintf to log a templated message.
	Debugf(format string, args ...any)
	// Infof uses fmt.Sprintf to log a templated message.
	Infof(format string, args ...any)
	// Warnf uses fmt.Sprintf to log a templated message.
	Warnf(format string, args ...any)
	// Errorf uses fmt.Sprintf to log a templated message.
	Errorf(format string, args ...any)

	// Debugw logs a message with some additional context. The variadic key-value
	// pairs are treated as they are in NewLoggerWith.
	Debugw(msg string, keyVals ...any)
	// Infow logs a message with some additional context. The variadic key-value
	// pairs are treated as they are in NewLoggerWith.
	Infow(msg string, keyVals ...any)
	// Warnw logs a message with some additional context. The variadic key-value
	// pairs are treated as they are in NewLoggerWith.
	Warnw(msg string, keyVals ...any)
	// Errorw logs a message with some additional context. The variadic key-value
	// pairs are treated as they are in NewLoggerWith.
	Errorw(msg string, keyVals ...any)

	// FlushLogger flushes any buffered log entries.
	FlushLogger() error
}
