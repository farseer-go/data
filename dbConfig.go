package data

import (
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
		return mysql.Open(receiver.ConnectionString)
	case "postgresql":
		return postgres.Open(receiver.ConnectionString)
	case "sqlite":
		return sqlite.Open(receiver.ConnectionString)
	case "sqlserver":
		return sqlserver.Open(receiver.ConnectionString)
	}
	panic("无法识别数据库类型：" + receiver.DataType)
}
