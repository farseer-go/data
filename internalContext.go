package data

import (
	"database/sql"
	"github.com/farseer-go/fs/flog"
	"github.com/timandy/routine"
	"gorm.io/gorm"
)

// 实现同一个协程下的事务作用域
var routineOrmClient = routine.NewInheritableThreadLocal[*gorm.DB]()

// InternalContext 数据库上下文
type InternalContext struct {
	dbConfig       *dbConfig          // 数据库配置
	IsolationLevel sql.IsolationLevel // 事务等级
}

// Begin 开启事务
func (receiver *InternalContext) Begin(isolationLevels ...sql.IsolationLevel) {
	if routineOrmClient.Get() == nil {
		ormClient, err := open(receiver.dbConfig)
		if err != nil {
			_ = flog.Error(err)
			return
		}

		// 事务等级
		isolationLevel := sql.LevelDefault
		if len(isolationLevels) > 0 {
			isolationLevel = isolationLevels[0]
		}
		// 开启事务
		ormClient = ormClient.Session(&gorm.Session{}).Begin(&sql.TxOptions{
			Isolation: isolationLevel,
		})
		routineOrmClient.Set(ormClient)
	}
}

// Transaction 使用事务
func (receiver *InternalContext) Transaction(executeFn func()) {
	receiver.Begin()
	executeFn()
	receiver.Commit()
}

// Commit 事务提交
func (receiver *InternalContext) Commit() {
	routineOrmClient.Get().Commit()
	routineOrmClient.Remove()
}

// Rollback 事务回滚
func (receiver *InternalContext) Rollback() {
	routineOrmClient.Get().Rollback()
	routineOrmClient.Remove()
}

// Original 返回原生的对象
func (receiver *InternalContext) Original() *gorm.DB {
	var gormDB *gorm.DB
	var err error

	// 上下文没有开启事务
	if routineOrmClient.Get() == nil {
		gormDB, err = open(receiver.dbConfig)
		gormDB = gormDB.Session(&gorm.Session{})
	} else {
		gormDB = routineOrmClient.Get()
	}

	if err != nil {
		_ = flog.Error(err)
		return nil
	}
	return gormDB
}
