package data

import (
	"context"
	"gorm.io/gorm/logger"
	"time"
)

type customLogger struct {
}

func (receiver *customLogger) LogMode(logger.LogLevel) logger.Interface {
	return receiver
}

func (receiver *customLogger) Info(context.Context, string, ...interface{}) {
}

func (receiver *customLogger) Warn(context.Context, string, ...interface{}) {
}

func (receiver *customLogger) Error(context.Context, string, ...interface{}) {
}

func (receiver *customLogger) Trace(ctx context.Context, begin time.Time, fc func() (string, int64), err error) {
}
