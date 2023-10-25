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
	// 参考：https://gorm.io/zh_CN/docs/connecting_to_the_database.html
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
		// clickhouse://user:123456@127.0.0.1:9942/dbname?dial_timeout=10s&read_timeout=20s
		return clickhouse.New(clickhouse.Config{
			DSN:                          receiver.ConnectionString,
			DisableDatetimePrecision:     true,     // disable datetime64 precision, not supported before clickhouse 20.4
			DontSupportRenameColumn:      true,     // rename column not supported before clickhouse 20.4
			DontSupportEmptyDefaultValue: false,    // do not consider empty strings as valid default values
			SkipInitializeWithVersion:    false,    // smart configure based on used version
			DefaultGranularity:           3,        // 1 granule = 8192 rows
			DefaultCompression:           "LZ4",    // default compression algorithm. LZ4 is lossless
			DefaultIndexType:             "minmax", // index stores extremes of the expression
			DefaultTableEngineOpts:       "ENGINE=MergeTree() ORDER BY tuple()",
		})
	}
	panic("无法识别数据库类型：" + receiver.DataType)
}
