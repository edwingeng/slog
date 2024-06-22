package slog

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"golang.org/x/term"
	"os"
	"time"
)

type Config struct {
	zap.Config
}

func (cfg *Config) Build(opts ...zap.Option) (*ZapLogger, error) {
	l, err := cfg.Config.Build(opts...)
	if err != nil {
		return nil, err
	}

	zsl := l.Sugar()
	return NewZapLogger(zsl), nil
}

func (cfg *Config) MustBuild(opts ...zap.Option) *ZapLogger {
	l, err := cfg.Build(opts...)
	if err != nil {
		panic(err)
	}
	return l
}

func NewDevelopmentConfig() *Config {
	return NewDevelopmentConfigWith("02|15:04:05.000")
}

func NewDevelopmentConfigWith(dateTimeFormat string) *Config {
	cfg := zap.NewDevelopmentConfig()
	cfg.DisableStacktrace = true
	if term.IsTerminal(int(os.Stdout.Fd())) {
		cfg.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	}
	cfg.EncoderConfig.EncodeTime = func(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
		enc.AppendString(t.Format(dateTimeFormat))
	}
	return &Config{cfg}
}

func NewProductionConfig() *Config {
	cfg := zap.NewProductionConfig()
	cfg.DisableStacktrace = true
	cfg.EncoderConfig.EncodeTime = func(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
		enc.AppendInt64(t.UnixMilli())
	}
	return &Config{cfg}
}
