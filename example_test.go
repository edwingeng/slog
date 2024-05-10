package slog_test

import (
	"context"
	"github.com/edwingeng/slog"
)

func Example() {
	logger, err := slog.NewDevelopmentConfig().Build()
	if err != nil {
		panic(err)
	}

	logger.Debug("str1")
	logger.Infof("str2: %d", 200)
	logger.Warnw("str3", "foo", 300)

	type HandlerContext struct {
		context.Context
		slog.Logger
	}

	ctx := &HandlerContext{
		Context: context.TODO(),
		Logger:  logger.NewLoggerWith("handler", "UpdateUserName"),
	}
	ctx.Error("invalid user name")
}
