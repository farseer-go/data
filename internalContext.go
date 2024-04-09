package data

import (
	"database/sql"
	"github.com/farseer-go/fs/asyncLocal"
	"github.com/farseer-go/fs/configure"
	"github.com/farseer-go/fs/container"
	"github.com/farseer-go/fs/core"
	"github.com/farseer-go/fs/exception"
	"github.com/farseer-go/fs/trace"
	"gorm.io/gorm"
	"strings"
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
	dbName         string             // 库名
	nameReplacer   *strings.Replacer  // 替换dbName、tableName
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
		panic("[farseer.yaml]Database." + key + ".ConnectionString，配置不正确" + configString)
	}
	if config.DataType == "" {
		panic("[farseer.yaml]Database." + key + ".DataType，配置不正确：" + configString)
	}
	config.DataType = strings.ToLower(config.DataType)
	config.keyName = key

	// 获取数据库名称
	switch config.DataType {
	case "sqlserver":
		// DataType=sqlserver,PoolMaxSize=50,PoolMinSize=1,ConnectionString=sqlserver://user:123456@127.0.0.1:9930?database=dbname
		dbNames := strings.Split(config.ConnectionString, "?") // database=dbname
		for _, name := range dbNames {
			if strings.HasPrefix(strings.ToLower(name), "database=") {
				config.databaseName = strings.Split(name, "=")[1]
			}
		}
	case "sqlite":
		// DataType=sqlite,PoolMaxSize=50,PoolMinSize=1,ConnectionString=gorm.db
		dbNames := strings.Split(config.ConnectionString, ",") // ConnectionString=gorm.db
		for _, name := range dbNames {
			if strings.HasPrefix(strings.ToLower(name), "connectionstring=") {
				config.databaseName = strings.Split(name, "=")[1]
			}
		}
	case "postgresql", "postgres":
		// host=127.0.0.1 user=user password=123456 dbname=dbname port=9920 sslmode=disable TimeZone=Asia/Shanghai
		dbNames := strings.Split(config.ConnectionString, " ")
		for _, name := range dbNames {
			if strings.HasPrefix(strings.ToLower(name), "dbname=") {
				config.databaseName = strings.Split(name, "=")[1]
			}
		}
	case "clickhouse", "mysql":
		// clickhouse://user:123456@127.0.0.1:9942/dbname?dial_timeout=10s&read_timeout=20s
		// user:123456@tcp(127.0.0.1:3306)/dbname?charset=utf8mb4&parseTime=True&loc=Local
		dbNames := strings.Split(config.ConnectionString, "/") // dbname?charset=utf8mb4&parseTime=True&loc=Local
		config.databaseName = dbNames[len(dbNames)-1]          // dbname?charset=utf8mb4&parseTime=True&loc=Local
		config.databaseName = strings.Split(config.databaseName, "?")[0]
	}

	// 初始化共享事务
	routineOrmClient[key] = asyncLocal.New[*gorm.DB]()

	// 注册上下文
	ins := &internalContext{dbConfig: &config}
	ins.dbName = config.databaseName
	ins.nameReplacer = strings.NewReplacer("{database}", config.databaseName)
	container.RegisterInstance[core.ITransaction](ins, key)

	// 注册健康检查
	container.RegisterInstance[core.IHealthCheck](&healthCheck{name: key}, "db_"+key)
}

// Begin 开启事务
func (receiver *internalContext) Begin(isolationLevels ...sql.IsolationLevel) error {
	// 事务等级
	isolationLevel := sql.LevelDefault
	if len(isolationLevels) > 0 {
		isolationLevel = isolationLevels[0]
	}

	if routineOrmClient[receiver.dbConfig.keyName].Get() != nil {
		exception.ThrowException("不支持两个事务同时运行，请先将上一个事物提交后在运行下一个。")
	}

	gormDB, err := open(receiver.dbConfig)
	if err != nil {
		return err
	}
	// 开启事务
	gormDB = gormDB.Session(&gorm.Session{}).Begin(&sql.TxOptions{
		Isolation: isolationLevel,
	})
	routineOrmClient[receiver.dbConfig.keyName].Set(gormDB)
	return nil
}

// Transaction 使用事务
func (receiver *internalContext) Transaction(executeFn func(), isolationLevels ...sql.IsolationLevel) {
	var err error
	traceHand := container.Resolve[trace.IManager]().TraceHand("开启事务")
	defer func() { traceHand.End(err) }()

	// 开启事务
	if err = receiver.Begin(isolationLevels...); err != nil {
		return
	}

	// 执行数据库操作
	exception.Try(func() {
		executeFn()
		if err = routineOrmClient[receiver.dbConfig.keyName].Get().Error; err == nil {
			receiver.Commit()
		} else {
			receiver.Rollback()
		}
	}).CatchException(func(exp any) {
		receiver.Rollback()
		panic(exp)
	})
}

// Commit 事务提交
func (receiver *internalContext) Commit() {
	routineOrmClient[receiver.dbConfig.keyName].Get().Commit()
	routineOrmClient[receiver.dbConfig.keyName].Remove()
}

// Rollback 事务回滚
func (receiver *internalContext) Rollback() {
	routineOrmClient[receiver.dbConfig.keyName].Get().Rollback()
	routineOrmClient[receiver.dbConfig.keyName].Remove()
}

// Original 返回原生的对象
func (receiver *internalContext) Original() *gorm.DB {
	gormDB := routineOrmClient[receiver.dbConfig.keyName].Get()
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
	sql = receiver.nameReplacer.Replace(sql)
	result := receiver.Original().Exec(sql, values...)
	return result.RowsAffected, result.Error
}

// ExecuteSqlToResult 返回结果(执行自定义SQL)
func (receiver *internalContext) ExecuteSqlToResult(arrayOrEntity any, sql string, values ...any) (int64, error) {
	sql = receiver.nameReplacer.Replace(sql)
	result := receiver.Original().Raw(sql, values...)
	result.Find(arrayOrEntity)
	return result.RowsAffected, result.Error
}

// ExecuteSqlToValue 返回单个字段值(执行自定义SQL)
func (receiver *internalContext) ExecuteSqlToValue(field any, sql string, values ...any) (int64, error) {
	sql = receiver.nameReplacer.Replace(sql)
	result := receiver.Original().Raw(sql, values...).Scan(&field)
	return result.RowsAffected, result.Error
}
