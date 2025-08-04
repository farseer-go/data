package loggers

import (
	"context"
	"github.com/farseer-go/fs/flog"
	"gorm.io/gorm/logger"
	"time"
)

type FsLogger struct {
	LoggerLevel logger.LogLevel
}

func NewFsLogger() logger.Interface {
	return &FsLogger{}
}

func (f *FsLogger) LogMode(level logger.LogLevel) logger.Interface {
	f.LoggerLevel = level
	return f
}

func (f *FsLogger) Info(ctx context.Context, s string, i ...interface{}) {
	flog.Infof(s, i...)
}

func (f *FsLogger) Warn(ctx context.Context, s string, i ...interface{}) {
	flog.Warningf(s, i...)
}

func (f *FsLogger) Error(ctx context.Context, s string, i ...interface{}) {
	flog.Errorf(s, i...)
}

func (f *FsLogger) Trace(ctx context.Context, begin time.Time, fc func() (sql string, rowsAffected int64), err error) {
	elapsed := time.Since(begin)
	sql, rows := fc()
	flog.Tracef("sql: %s, rows: %d, elapsed: %s, err: %v", sql, rows, elapsed, err)
}
