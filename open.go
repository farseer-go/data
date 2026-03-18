package data

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/farseer-go/data/loggers"

	"github.com/farseer-go/fs/container"
	"github.com/farseer-go/fs/trace"
	"gorm.io/gorm"
)

var databaseConn map[string]*gorm.DB
var lock sync.Mutex

// 打开数据库（全局）
func open(dbConfig *dbConfig) (*gorm.DB, error) {
	db, exists := databaseConn[dbConfig.keyName]
	// 不存在，则创建
	if !exists {
		lock.Lock()
		defer lock.Unlock()
		traceManager := container.Resolve[trace.IManager]()
		traceDatabase := traceManager.TraceDatabaseOpen(dbConfig.databaseName, dbConfig.ConnectionString)
		// 连接数据库参考：https://gorm.io/zh_CN/docs/connecting_to_the_database.html
		// Data Source ClientName 参考 https://github.com/go-sql-driver/mysql#dsn-data-source-name

		// 禁用默认事务
		skipDefaultTransaction := true
		// // clickhouse的BUG，必须设为false，则否会出现数据无法写入的问题
		// if strings.ToLower(dbConfig.DataType) == "clickhouse" {
		// 	skipDefaultTransaction = false
		// }

		gormDB, err := gorm.Open(dbConfig.GetDriver(), &gorm.Config{
			SkipDefaultTransaction:                   skipDefaultTransaction,
			DisableForeignKeyConstraintWhenMigrating: true, // 禁止自动创建数据库外键约束
			PrepareStmt:                              true,
			Logger:                                   loggers.NewFsLogger(),
			//Logger: logger.New(
			//	log.New(os.Stdout, "\r\n", log.LstdFlags), // io writer
			//	logger.Config{
			//		SlowThreshold:             time.Second, // 慢 SQL 阈值
			//		Colorful:                  false,       // 禁用彩色打印
			//		IgnoreRecordNotFoundError: true,
			//		ParameterizedQueries:      false,
			//		LogLevel:                  logger.Error, // Log level
			//	},
			//),
		})
		defer traceDatabase.End(err)
		if err != nil {
			return gormDB, fmt.Errorf("打开[%s]数据库[%s]失败：%s", strings.ToLower(dbConfig.DataType), dbConfig.keyName, err.Error())
		}

		_ = gormDB.Use(&TracePlugin{traceManager: traceManager})
		// 设置池大小
		setPool(gormDB, dbConfig)
		// 如果是动态连接，dbConfig.keyName是空的
		if dbConfig.keyName != "" {
			databaseConn[dbConfig.keyName] = gormDB
		}
		db = gormDB
	}
	return db, nil
}

// 设置池大小
func setPool(gormDB *gorm.DB, dbConfig *dbConfig) {
	sqlDB, _ := gormDB.DB()
	if dbConfig.PoolMaxSize > 0 {
		// 设置空闲连接池中连接的最大数量
		sqlDB.SetMaxIdleConns(dbConfig.PoolMaxSize / 3)
		// 设置打开数据库连接的最大数量。
		sqlDB.SetMaxOpenConns(dbConfig.PoolMaxSize)
	}

	// clickhouse需要更短的连接生命周期和空闲超时，避免连接状态不一致
	if strings.ToLower(dbConfig.DataType) == "clickhouse" {
		// 强制连接在使用一段时间后被关闭重建，防止连接因网络中间件超时变“脏”。 (关键！)
		// 建议设置为 5-10 分钟。
		sqlDB.SetConnMaxLifetime(5 * time.Minute)
		// 设置连接最大空闲时间
		// 连接空闲超过多久就丢弃。
		sqlDB.SetConnMaxIdleTime(1 * time.Minute)
	} else {
		// 其他数据库设置了连接可复用的最大时间。
		sqlDB.SetConnMaxLifetime(time.Hour)
	}
}
