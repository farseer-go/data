package loggers

import (
	"context"
	"time"

	"github.com/farseer-go/fs/configure"
	"github.com/farseer-go/fs/flog"
	"gorm.io/gorm/logger"
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
	// 配置文件中开启后，才打印trace
	if configure.GetBool("Log.Component.data") {
		elapsed := time.Since(begin)
		sql, rows := fc()
		flog.Tracef("sql: %s, rows: %d, elapsed: %s, err: %v", sql, rows, elapsed, err)
	}
}
