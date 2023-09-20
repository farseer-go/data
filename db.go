package data

import (
	"github.com/farseer-go/fs/flog"
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

		// Data Source ClientName，参考 https://github.com/go-sql-driver/mysql#dsn-data-source-name
		gormDB, err := gorm.Open(dbConfig.getDriver(), &gorm.Config{
			SkipDefaultTransaction:                   true,
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
		// SetMaxIdleConns 设置空闲连接池中连接的最大数量
		sqlDB.SetMaxIdleConns(dbConfig.PoolMaxSize / 3)
		// SetMaxOpenConns 设置打开数据库连接的最大数量。
		sqlDB.SetMaxOpenConns(dbConfig.PoolMaxSize)
	}
	// SetConnMaxLifetime 设置了连接可复用的最大时间。
	sqlDB.SetConnMaxLifetime(time.Hour)
}
