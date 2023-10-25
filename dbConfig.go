package data

import (
	"gorm.io/driver/clickhouse"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/driver/sqlserver"
	"gorm.io/gorm"
	"strings"
)

// 数据库配置
type dbConfig struct {
	dbName           string
	DataType         string
	PoolMaxSize      int
	PoolMinSize      int
	ConnectionString string
}

// 获取对应驱动
func (receiver *dbConfig) getDriver() gorm.Dialector {
	// 参考：https://gorm.cn/zh_CN/docs/connecting_to_the_database.html
	switch strings.ToLower(receiver.DataType) {
	case "mysql":
		// user:123456@tcp(127.0.0.1:3306)/dbname?charset=utf8mb4&parseTime=True&loc=Local
		return mysql.Open(receiver.ConnectionString)
	case "postgresql":
		// host=127.0.0.1 user=user password=123456 dbname=dbname port=9920 sslmode=disable TimeZone=Asia/Shanghai
		return postgres.Open(receiver.ConnectionString)
	case "sqlite":
		// gorm.db
		return sqlite.Open(receiver.ConnectionString)
	case "sqlserver":
		// sqlserver://user:123456@127.0.0.1:9930?database=dbname
		return sqlserver.Open(receiver.ConnectionString)
	case "clickhouse":
		// tcp://127.0.0.1:9000?database=dbname&username=default&password=&read_timeout=10&write_timeout=20
		return clickhouse.Open(receiver.ConnectionString)
	}
	panic("无法识别数据库类型：" + receiver.DataType)
}
