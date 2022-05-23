package slog

// Logger is an interface which is highly compatible with zap.SugaredLogger.
type Logger interface {
	// NewLoggerWith adds a variadic number of fields to the logging context. It accepts a
	// mix of zap.Field objects and loosely-typed key-value pairs. When
	// processing pairs, the first element of the pair is used as the field key
	// and the second as the field value. The keys in key-value pairs should be strings.
	NewLoggerWith(keyVals ...interface{}) Logger

	// Debug uses fmt.Sprint to construct and log a message.
	Debug(args ...interface{})
	// Info uses fmt.Sprint to construct and log a message.
	Info(args ...interface{})
	// Warn uses fmt.Sprint to construct and log a message.
	Warn(args ...interface{})
	// Error uses fmt.Sprint to construct and log a message.
	Error(args ...interface{})

	// Debugf uses fmt.Sprintf to log a templated message.
	Debugf(format string, args ...interface{})
	// Infof uses fmt.Sprintf to log a templated message.
	Infof(format string, args ...interface{})
	// Warnf uses fmt.Sprintf to log a templated message.
	Warnf(format string, args ...interface{})
	// Errorf uses fmt.Sprintf to log a templated message.
	Errorf(format string, args ...interface{})

	// Debugw logs a message with some additional context. The variadic key-value
	// pairs are treated as they are in NewLoggerWith.
	Debugw(msg string, keyVals ...interface{})
	// Infow logs a message with some additional context. The variadic key-value
	// pairs are treated as they are in NewLoggerWith.
	Infow(msg string, keyVals ...interface{})
	// Warnw logs a message with some additional context. The variadic key-value
	// pairs are treated as they are in NewLoggerWith.
	Warnw(msg string, keyVals ...interface{})
	// Errorw logs a message with some additional context. The variadic key-value
	// pairs are treated as they are in NewLoggerWith.
	Errorw(msg string, keyVals ...interface{})

	// FlushLogger flushes any buffered log entries.
	FlushLogger() error
}
