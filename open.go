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

		gormDB, err := gorm.Open(dbConfig.GetDriver(), &gorm.Config{
			SkipDefaultTransaction:                   true,
			DisableForeignKeyConstraintWhenMigrating: true, // 禁止自动创建数据库外键约束
			Logger:                                   loggers.NewFsLogger(),
			// Logger: logger.New(
			// 	log.New(os.Stdout, "\r\n", log.LstdFlags), // io writer
			// 	logger.Config{
			// 		SlowThreshold:             time.Second, // 慢 SQL 阈值
			// 		Colorful:                  false,       // 禁用彩色打印
			// 		IgnoreRecordNotFoundError: true,
			// 		ParameterizedQueries:      false,
			// 		LogLevel:                  logger.Info, // Log level
			// 	},
			// ),
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

	// 建议：MaxOpenConns 和 MaxIdleConns 设置为相同值
	// 这样可以避免连接频繁创建和销毁带来的性能损耗
	// 建议值：根据你的并发量，设置为 10-20 比较合适
	maxSize := dbConfig.PoolMaxSize
	if maxSize < 10 {
		maxSize = 10 // 强制最小10个，避免瓶颈
	}

	sqlDB.SetMaxOpenConns(maxSize) // 最大连接数
	sqlDB.SetMaxIdleConns(maxSize) // 空闲连接数等于最大连接数（保持长连接）

	if strings.ToLower(dbConfig.DataType) == "clickhouse" {
		// 这样可以避免每隔几分钟就重建连接
		sqlDB.SetConnMaxLifetime(10 * time.Minute)
		// 空闲超时：既然 MaxIdleConns 已经等于 MaxOpenConns，这个参数主要防止极端空闲
		sqlDB.SetConnMaxIdleTime(5 * time.Minute)
	} else {
		sqlDB.SetConnMaxLifetime(time.Hour)
	}
}
