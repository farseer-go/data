package data

import "fmt"

// 创建data组件需要的数据库连接字符串
func CreateConnectionString(dataType string, host string, port int, database string, username, password string) string {
	// 获取数据库名称
	switch dataType {
	case "sqlserver":
		// DataType=sqlserver,PoolMaxSize=50,PoolMinSize=1,ConnectionString=sqlserver://user:123456@127.0.0.1:9930?database=dbname
		return fmt.Sprintf("DataType=sqlserver,PoolMaxSize=50,PoolMinSize=1,ConnectionString=sqlserver://%s:%s@%s:%d?database=%s", username, password, host, port, database)
	case "sqlite":
		// DataType=sqlite,PoolMaxSize=50,PoolMinSize=1,ConnectionString=gorm.db
		return fmt.Sprintf("DataType=sqlite,PoolMaxSize=50,PoolMinSize=1,ConnectionString=%s", database)
	case "postgresql", "postgres":
		// host=127.0.0.1 user=user password=123456 dbname=dbname port=9920 sslmode=disable TimeZone=Asia/Shanghai
		return fmt.Sprintf("DataType=postgresql,PoolMaxSize=50,PoolMinSize=1,ConnectionString=host=%s user=%s password=%s dbname=%s port=%d sslmode=disable TimeZone=Asia/Shanghai", host, username, password, database, port)
	case "mysql":
		// user:pass@tcp(127.0.0.1:3306)/dbname?charset=utf8mb4&parseTime=True&loc=Local
		return fmt.Sprintf("DataType=mysql,PoolMaxSize=50,PoolMinSize=1,ConnectionString=%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local", username, password, host, port, database)
	case "clickhouse":
		// clickhouse://user:123456@127.0.0.1:9000/dbname?dial_timeout=10s&read_timeout=60s
		return fmt.Sprintf("DataType=clickhouse,PoolMaxSize=50,PoolMinSize=1,ConnectionString=clickhouse://%s:%s@%s:%d/%s?dial_timeout=10s&read_timeout=60s", username, password, host, port, database)
	}
	return ""
}
