package data

import (
	"github.com/farseer-go/fs/container"
	"github.com/farseer-go/fs/flog"
	"github.com/farseer-go/fs/trace"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"log"
	"os"
	"sync"
	"time"
)

var databaseConn map[string]*gorm.DB
var lock sync.Mutex

// 打开数据库（全局）
func open(dbConfig *dbConfig) (*gorm.DB, error) {
	db, exists := databaseConn[dbConfig.dbName]
	// 不存在，则创建
	if !exists {
		lock.Lock()
		defer lock.Unlock()
		traceManager := container.Resolve[trace.IManager]()
		traceDatabase := traceManager.TraceDatabaseOpen(dbConfig.dbName, dbConfig.ConnectionString)

		// 连接数据库参考：https://gorm.io/zh_CN/docs/connecting_to_the_database.html
		// Data Source ClientName 参考 https://github.com/go-sql-driver/mysql#dsn-data-source-name
		gormDB, err := gorm.Open(dbConfig.GetDriver(), &gorm.Config{
			SkipDefaultTransaction:                   false,
			DisableForeignKeyConstraintWhenMigrating: true, // 禁止自动创建数据库外键约束
			Logger: logger.New(
				log.New(os.Stdout, "\r\n", log.LstdFlags), // io writer
				logger.Config{
					SlowThreshold:             time.Second, // 慢 SQL 阈值
					Colorful:                  false,       // 禁用彩色打印
					IgnoreRecordNotFoundError: true,
					ParameterizedQueries:      false,
					LogLevel:                  logger.Error, // Log level
				},
			),
		})
		defer traceDatabase.End(err)
		_ = gormDB.Use(&TracePlugin{traceManager: traceManager})
		if err != nil {
			_ = flog.Error(err)
			return gormDB, err
		}

		setPool(gormDB, dbConfig)

		databaseConn[dbConfig.dbName] = gormDB
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
	// 设置了连接可复用的最大时间。
	sqlDB.SetConnMaxLifetime(time.Hour)
}
