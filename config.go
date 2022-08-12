package slog

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
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

func NewDevelopmentConfig() *Config {
	return NewDevelopmentConfigWith("02|15:04:05")
}

func NewDevelopmentConfigWith(dateTimeFormat string) *Config {
	cfg := zap.NewDevelopmentConfig()
	cfg.DisableStacktrace = true
	cfg.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
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
