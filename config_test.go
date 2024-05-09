package slog

import (
	"go.uber.org/zap/zapcore"
	"testing"
)

func TestLogLevelEnabled(t *testing.T) {
	a := []Logger{
		NewDevelopmentConfig().MustBuild(), NewProductionConfig().MustBuild(),
	}
	ins := []struct {
		idx      int
		level    int
		expected bool
	}{
		{idx: 0, level: int(zapcore.DebugLevel), expected: true},
		{idx: 0, level: int(zapcore.InfoLevel), expected: true},
		{idx: 0, level: int(zapcore.WarnLevel), expected: true},
		{idx: 0, level: int(zapcore.ErrorLevel), expected: true},
		{idx: 1, level: int(zapcore.DebugLevel), expected: false},
		{idx: 1, level: int(zapcore.InfoLevel), expected: true},
		{idx: 1, level: int(zapcore.WarnLevel), expected: true},
		{idx: 1, level: int(zapcore.ErrorLevel), expected: true},
	}

	for _, x := range ins {
		if a[x.idx].LogLevelEnabled(x.level) != x.expected {
			t.Fatal(`a[x.idx].LogLevelEnabled(x.level) != x.expected`)
		}
	}
}
