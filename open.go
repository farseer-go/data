package data

import (
	"fmt"
	"github.com/farseer-go/fs/flog"
	"github.com/farseer-go/linkTrace"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"log"
	"os"
	"strings"
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
		_ = gormDB.Callback().Raw().Register("raw", callBack)
		_ = gormDB.Callback().Create().Register("Create", callBack)
		_ = gormDB.Callback().Delete().Register("Delete", callBack)
		_ = gormDB.Callback().Update().Register("Update", callBack)
		_ = gormDB.Callback().Query().Register("Query", callBack)
		_ = gormDB.Callback().Row().Register("Row", callBack)

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

// 链路追踪记录
func callBack(db *gorm.DB) {
	sql := db.Statement.SQL.String()
	// 将参数化替换成原SQL
	for i := 0; i < len(db.Statement.Vars); i++ {
		switch db.Statement.Vars[i].(type) {
		case string, time.Time:
			sql = strings.Replace(sql, "?", fmt.Sprintf("'%v'", db.Statement.Vars[i]), 1)
		default:
			sql = strings.Replace(sql, "?", fmt.Sprintf("%v", db.Statement.Vars[i]), 1)
		}
	}
	linkTrace.TraceDatabase(db.Statement.DB.Name(), db.Statement.Table, sql)
}
