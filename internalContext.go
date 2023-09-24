package data

import (
	"database/sql"
	"github.com/farseer-go/fs/configure"
	"github.com/farseer-go/fs/container"
	"github.com/farseer-go/fs/core"
	"github.com/timandy/routine"
	"gorm.io/gorm"
)

// 实现同一个协程下的事务作用域
var routineOrmClient = routine.NewInheritableThreadLocal[*gorm.DB]()

type IInternalContext interface {
	core.ITransaction
	Original() *gorm.DB
}

// internalContext 数据库上下文
type internalContext struct {
	dbConfig       *dbConfig          // 数据库配置
	IsolationLevel sql.IsolationLevel // 事务等级
}

// RegisterInternalContext 注册内部上下文
func RegisterInternalContext(key string, configString string) {
	config := configure.ParseString[dbConfig](configString)
	if config.ConnectionString == "" {
		panic("[farseer.yaml]Database." + key + ".ConnectionString，没有正确配置")
	}
	if config.DataType == "" {
		panic("[farseer.yaml]Database." + key + ".DataType，没有正确配置")
	}
	config.dbName = key

	// 注册上下文
	container.RegisterInstance[core.ITransaction](&internalContext{dbConfig: &config}, key)

	// 注册健康检查
	container.RegisterInstance[core.IHealthCheck](&healthCheck{name: key}, "db_"+key)
}

// Begin 开启事务
func (receiver *internalContext) Begin(isolationLevels ...sql.IsolationLevel) {
	if routineOrmClient.Get() == nil {
		ormClient, err := open(receiver.dbConfig)
		if err != nil {
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
func (receiver *internalContext) Transaction(executeFn func()) {
	receiver.Begin()
	executeFn()
	receiver.Commit()
}

// Commit 事务提交
func (receiver *internalContext) Commit() {
	routineOrmClient.Get().Commit()
	routineOrmClient.Remove()
}

// Rollback 事务回滚
func (receiver *internalContext) Rollback() {
	routineOrmClient.Get().Rollback()
	routineOrmClient.Remove()
}

// Original 返回原生的对象
func (receiver *internalContext) Original() *gorm.DB {
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
		return nil
	}
	return gormDB
}
