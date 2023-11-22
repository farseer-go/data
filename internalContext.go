package data

import (
	"database/sql"
	"github.com/farseer-go/fs/asyncLocal"
	"github.com/farseer-go/fs/configure"
	"github.com/farseer-go/fs/container"
	"github.com/farseer-go/fs/core"
	"github.com/farseer-go/fs/exception"
	"gorm.io/gorm"
)

// 实现同一个协程下的事务作用域
var routineOrmClient = make(map[string]asyncLocal.AsyncLocal[*gorm.DB])

type IInternalContext interface {
	core.ITransaction
	Original() *gorm.DB
	// ExecuteSql 执行自定义SQL
	ExecuteSql(sql string, values ...any) (int64, error)
	// ExecuteSqlToResult 返回结果(执行自定义SQL)
	ExecuteSqlToResult(arrayOrEntity any, sql string, values ...any) (int64, error)
	// ExecuteSqlToValue 返回单个字段值(执行自定义SQL)
	ExecuteSqlToValue(field any, sql string, values ...any) (int64, error)
}

// internalContext 数据库上下文
type internalContext struct {
	dbConfig       *dbConfig          // 数据库配置
	IsolationLevel sql.IsolationLevel // 事务等级
}

// RegisterInternalContext 注册内部上下文
// DataType=mysql,PoolMaxSize=50,PoolMinSize=1,ConnectionString=user:pass@tcp(127.0.0.1:3306)/dbname?charset=utf8mb4&parseTime=True&loc=Local
// DataType=sqlserver,PoolMaxSize=50,PoolMinSize=1,ConnectionString=sqlserver://user:123456@127.0.0.1:9930?database=dbname
// DataType=clickhouse,PoolMaxSize=50,PoolMinSize=1,ConnectionString=tcp://192.168.1.8:9000?database=dbname&username=default&password=&read_timeout=10&write_timeout=20
// DataType=postgresql,PoolMaxSize=50,PoolMinSize=1,ConnectionString=host=127.0.0.1 user=user password=123456 dbname=dbname port=9920 sslmode=disable TimeZone=Asia/Shanghai
// DataType=sqlite,PoolMaxSize=50,PoolMinSize=1,ConnectionString=gorm.db
func RegisterInternalContext(key string, configString string) {
	config := configure.ParseString[dbConfig](configString)
	if config.ConnectionString == "" {
		panic("[farseer.yaml]Database." + key + ".ConnectionString，没有正确配置")
	}
	if config.DataType == "" {
		panic("[farseer.yaml]Database." + key + ".DataType，没有正确配置")
	}
	config.dbName = key

	// 初始化共享事务
	routineOrmClient[key] = asyncLocal.New[*gorm.DB]()

	// 注册上下文
	container.RegisterInstance[core.ITransaction](&internalContext{dbConfig: &config}, key)

	// 注册健康检查
	container.RegisterInstance[core.IHealthCheck](&healthCheck{name: key}, "db_"+key)
}

// Begin 开启事务
func (receiver *internalContext) Begin(isolationLevels ...sql.IsolationLevel) {
	if routineOrmClient[receiver.dbConfig.dbName].Get() == nil {
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
		routineOrmClient[receiver.dbConfig.dbName].Set(ormClient)
	}
}

// Transaction 使用事务
func (receiver *internalContext) Transaction(executeFn func()) {
	receiver.Begin()
	exception.Try(func() {
		executeFn()
		receiver.Commit()
	}).CatchException(func(exp any) {
		receiver.Rollback()
		panic(exp)
	})
}

// Commit 事务提交
func (receiver *internalContext) Commit() {
	routineOrmClient[receiver.dbConfig.dbName].Get().Commit()
	routineOrmClient[receiver.dbConfig.dbName].Remove()
}

// Rollback 事务回滚
func (receiver *internalContext) Rollback() {
	routineOrmClient[receiver.dbConfig.dbName].Get().Rollback()
	routineOrmClient[receiver.dbConfig.dbName].Remove()
}

// Original 返回原生的对象
func (receiver *internalContext) Original() *gorm.DB {
	gormDB := routineOrmClient[receiver.dbConfig.dbName].Get()
	var err error

	// 上下文没有开启事务
	if gormDB == nil {
		gormDB, err = open(receiver.dbConfig)
		gormDB = gormDB.Session(&gorm.Session{})
	}

	if err != nil {
		return nil
	}
	return gormDB
}

// ExecuteSql 执行自定义SQL
func (receiver *internalContext) ExecuteSql(sql string, values ...any) (int64, error) {
	result := receiver.Original().Exec(sql, values...)
	return result.RowsAffected, result.Error
}

// ExecuteSqlToResult 返回结果(执行自定义SQL)
func (receiver *internalContext) ExecuteSqlToResult(arrayOrEntity any, sql string, values ...any) (int64, error) {
	result := receiver.Original().Raw(sql, values...)
	result.Find(arrayOrEntity)
	return result.RowsAffected, result.Error
}

// ExecuteSqlToValue 返回单个字段值(执行自定义SQL)
func (receiver *internalContext) ExecuteSqlToValue(field any, sql string, values ...any) (int64, error) {
	result := receiver.Original().Raw(sql, values...)
	rows, _ := result.Rows()
	if rows == nil {
		return 0, nil
	}
	defer func() {
		_ = rows.Close()
	}()
	for rows.Next() {
		_ = rows.Scan(&field)
	}
	return result.RowsAffected, result.Error
}
