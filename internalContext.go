package data

import (
	"database/sql"
	"fmt"
	"net/url"
	"strings"

	"github.com/farseer-go/fs/asyncLocal"
	"github.com/farseer-go/fs/configure"
	"github.com/farseer-go/fs/container"
	"github.com/farseer-go/fs/core"
	"github.com/farseer-go/fs/exception"
	"github.com/farseer-go/fs/flog"
	"github.com/farseer-go/fs/trace"
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
	// GetDatabaseList 获取数据库列表
	GetDatabaseList() ([]string, error)
	// GetTableList 获取所有表
	GetTableList(database string) ([]string, error)
}

type IGetInternalContext interface {
	GetInternalContext() IInternalContext
}

// internalContext 数据库上下文
type internalContext struct {
	dbConfig       *dbConfig          // 数据库配置
	IsolationLevel sql.IsolationLevel // 事务等级
	nameReplacer   *strings.Replacer  // 替换dbName、tableName
}

// RegisterInternalContext 注册内部上下文
// DataType=mysql,PoolMaxSize=5,PoolMinSize=1,ConnectionString=user:pass@tcp(127.0.0.1:3306)/dbname?charset=utf8mb4&parseTime=True&loc=Local
// DataType=sqlserver,PoolMaxSize=5,PoolMinSize=1,ConnectionString=sqlserver://user:123456@127.0.0.1:9930?database=dbname
// DataType=clickhouse,PoolMaxSize=5,PoolMinSize=1,ConnectionString=clickhouse://user:123456@127.0.0.1:9000/dbname?dial_timeout=10s&read_timeout=60s
// DataType=postgresql,PoolMaxSize=5,PoolMinSize=1,ConnectionString=host=127.0.0.1 user=user password=123456 dbname=dbname port=9920 sslmode=disable TimeZone=Asia/Shanghai
// DataType=sqlite,PoolMaxSize=5,PoolMinSize=1,ConnectionString=gorm.db
func RegisterInternalContext(key string, configString string) {
	ins := NewInternalContext(configString)
	if ins.dbConfig.ConnectionString == "" {
		panic("[farseer.yaml]Database." + key + ".ConnectionString，配置不正确" + configString)
	}
	if ins.dbConfig.DataType == "" {
		panic("[farseer.yaml]Database." + key + ".DataType，配置不正确：" + configString)
	}
	ins.dbConfig.keyName = key

	// 初始化共享事务
	routineOrmClient[key] = asyncLocal.New[*gorm.DB]()

	// 如果之前注册过，则先移除
	if container.IsRegister[core.ITransaction](key) {
		container.Remove[core.ITransaction](key)
	}
	container.RegisterInstance[core.ITransaction](ins, key)

	// 如果之前注册过，则先移除
	if container.IsRegister[core.IHealthCheck]("db_" + key) {
		container.Remove[core.IHealthCheck]("db_" + key)
	}
	// 注册健康检查
	container.RegisterInstance[core.IHealthCheck](&healthCheck{name: key, dataType: ins.dbConfig.DataType}, "db_"+key)
}

// 通过连接字符串解析数据库配置，得到internalContext
func NewInternalContext(configString string) *internalContext {
	config := configure.ParseString[dbConfig](configString)
	config.keyName = configString // 先默认为连接字符串，如果上游是RegisterInternalContext函数，则会覆盖（如果不设置默认值，则不会复用连接字符串）
	config.DataType = strings.ToLower(config.DataType)
	config.migrated = strings.Contains(configString, "Migrate=")

	// 获取数据库名称
	switch config.DataType {
	case "sqlserver", "mssql":
		// DataType=sqlserver,PoolMaxSize=5,PoolMinSize=1,ConnectionString=sqlserver://user:123456@127.0.0.1:9930?database=dbname
		dbNames := strings.Split(config.ConnectionString, "?") // database=dbname
		for _, name := range dbNames {
			if strings.HasPrefix(strings.ToLower(name), "database=") {
				config.databaseName = strings.Split(name, "=")[1]
			}
		}
	case "sqlite":
		// DataType=sqlite,PoolMaxSize=5,PoolMinSize=1,ConnectionString=gorm.db
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
	case "mysql":
		// user:pass@tcp(127.0.0.1:3306)/dbname?charset=utf8mb4&parseTime=True&loc=Local
		dbNames := strings.Split(config.ConnectionString, "/") // dbname?charset=utf8mb4&parseTime=True&loc=Local
		config.databaseName = dbNames[len(dbNames)-1]          // dbname?charset=utf8mb4&parseTime=True&loc=Local
		config.databaseName = strings.Split(config.databaseName, "?")[0]
	case "clickhouse":
		// clickhouse://user:123456@127.0.0.1:9000/dbname?dial_timeout=10s&read_timeout=60s
		if parsedURL, err := url.Parse(config.ConnectionString); err == nil {
			config.databaseName = strings.TrimPrefix(parsedURL.Path, "/")
		}
	}

	// 注册上下文
	ins := &internalContext{dbConfig: &config}
	//ins.dbName = config.databaseName
	ins.nameReplacer = strings.NewReplacer("{database}", config.databaseName)

	return ins
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
	if routineOrmClient[receiver.dbConfig.keyName].Get() != nil {
		executeFn()
		return
	}

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
	var gormDB *gorm.DB
	// 如果是动态连接，则routineOrmClient获取不到对象，因为receiver.dbConfig.keyName是空的
	if asyncLocalGormDB, exists := routineOrmClient[receiver.dbConfig.keyName]; exists {
		gormDB = asyncLocalGormDB.Get()
	}

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
	if result.Error != nil {
		flog.Errorf("执行ExecuteSqlToResult时出现异常,sql=%s,err=%s", sql, result.Error.Error())
	}
	return result.RowsAffected, result.Error
}

// ExecuteSqlToValue 返回单个字段值(执行自定义SQL)
func (receiver *internalContext) ExecuteSqlToValue(field any, sql string, values ...any) (int64, error) {
	sql = receiver.nameReplacer.Replace(sql)
	result := receiver.Original().Raw(sql, values...).Scan(&field)
	return result.RowsAffected, result.Error
}

func (receiver *internalContext) GetInternalContext() IInternalContext {
	return receiver
}

func (receiver *internalContext) GetDatabaseList() ([]string, error) {
	var arrayOrEntity []string
	var sql string
	switch receiver.dbConfig.DataType {
	case "mysql":
		sql = "SELECT SCHEMA_NAME FROM information_schema.SCHEMATA WHERE SCHEMA_NAME NOT IN ('mysql', 'information_schema', 'performance_schema', 'sys');"
	case "sqlserver", "mssql":
		sql = "SELECT name FROM sys.databases;"
	case "sqlite":
	case "postgresql", "postgres":
		sql = "SELECT datname FROM pg_database;"
	case "clickhouse":
		sql = "SELECT name FROM system.databases WHERE name NOT IN ('INFORMATION_SCHEMA', 'default', 'information_schema', 'system');"
	}
	original := receiver.Original()
	if original == nil {
		return arrayOrEntity, fmt.Errorf("数据库连接失败")
	}

	result := original.Raw(sql)
	result.Find(&arrayOrEntity)
	if result.Error != nil {
		return arrayOrEntity, fmt.Errorf("执行GetDatabaseList时出现异常,sql=%s,err=%s", sql, result.Error.Error())
	}
	return arrayOrEntity, nil
}

func (receiver *internalContext) GetTableList(database string) ([]string, error) {
	var arrayOrEntity []string
	var sql string
	switch receiver.dbConfig.DataType {
	case "mysql":
		sql = fmt.Sprintf("SHOW TABLES FROM %s;", database)
	case "sqlserver", "mssql":
		sql = fmt.Sprintf("SELECT TABLE_NAME FROM INFORMATION_SCHEMA.TABLES WHERE TABLE_CATALOG = '%s';", database)
	case "sqlite":
		sql = "SELECT name FROM sqlite_master WHERE type = 'table';"
	case "postgresql", "postgres":
		sql = "SELECT table_name FROM information_schema.tables WHERE table_schema = 'public';"
	case "clickhouse":
		sql = fmt.Sprintf("SELECT name FROM system.tables WHERE database = '%s' and engine <> 'View'", database)
	}
	original := receiver.Original()
	if original == nil {
		return arrayOrEntity, fmt.Errorf("数据库连接失败")
	}

	result := original.Raw(sql)
	result.Find(&arrayOrEntity)
	if result.Error != nil {
		return arrayOrEntity, fmt.Errorf("执行GetTableList时出现异常,sql=%s,err=%s", sql, result.Error.Error())
	}
	return arrayOrEntity, nil
}
