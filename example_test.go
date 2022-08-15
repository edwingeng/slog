package slog

import "context"

func Example() {
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
}
