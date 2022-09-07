package data

import (
	"github.com/farseer-go/fs/configure"
	"github.com/farseer-go/fs/types"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/driver/sqlserver"
	"gorm.io/gorm"
	"reflect"
	"strings"
)

// DbContext 数据库上下文
type DbContext struct {
	// 数据库配置
	dbConfig *dbConfig
}

// NewDbContext 初始化上下文
func initConfig(dbName string) *DbContext {
	configString := configure.GetString("Database." + dbName)
	if configString == "" {
		panic("[farseer.yaml]找不到相应的配置：Database." + dbName)
	}
	dbConfig := configure.ParseConfig[dbConfig](configString)
	dbContext := &DbContext{
		dbConfig: &dbConfig,
	}
	dbContext.dbConfig.dbName = dbName
	return dbContext
}

func checkConfig() {
	nodes := configure.GetSubNodes("Database")
	for key, configString := range nodes {
		if configString == "" {
			panic("[farseer.yaml]Database." + key + "，没有正确配置")
		}
		dbConfig := configure.ParseConfig[dbConfig](configString)
		if dbConfig.ConnectionString == "" {
			panic("[farseer.yaml]Database." + key + ".ConnectionString，没有正确配置")
		}
		if dbConfig.DataType == "" {
			panic("[farseer.yaml]Database." + key + ".DataType，没有正确配置")
		}
	}
}

// NewContext 数据库上下文初始化
// dbName：数据库配置名称
func NewContext[TDbContext any](dbName string) *TDbContext {
	var context TDbContext
	InitContext(&context, dbName)
	return &context
}

// InitContext 数据库上下文初始化
// dbName：数据库配置名称
func InitContext[TDbContext any](dbContext *TDbContext, dbName string) {
	if dbName == "" {
		panic("dbName入参必须设置有效的值")
	}
	dbConfig := initConfig(dbName) // 嵌入类型
	contextValueOf := reflect.ValueOf(dbContext).Elem()

	for i := 0; i < contextValueOf.NumField(); i++ {
		field := contextValueOf.Field(i)
		_, isDataTableSet := types.IsDataTableSet(field)
		if field.CanSet() && isDataTableSet {
			data := contextValueOf.Type().Field(i).Tag.Get("data")
			var tableName string
			if strings.HasPrefix(data, "name=") {
				tableName = data[len("name="):]
			}
			if tableName != "" {
				// 再取tableSet的子属性，并设置值
				field.Addr().MethodByName("Init").Call([]reflect.Value{reflect.ValueOf(dbConfig), reflect.ValueOf(tableName)})
			}
		}
	}
}

// 获取对应驱动
func (dbContext *DbContext) getDriver() gorm.Dialector {
	// 参考：https://gorm.cn/zh_CN/docs/connecting_to_the_database.html
	switch strings.ToLower(dbContext.dbConfig.DataType) {
	case "mysql":
		return mysql.Open(dbContext.dbConfig.ConnectionString)
	case "postgresql":
		return postgres.Open(dbContext.dbConfig.ConnectionString)
	case "sqlite":
		return sqlite.Open(dbContext.dbConfig.ConnectionString)
	case "sqlserver":
		return sqlserver.Open(dbContext.dbConfig.ConnectionString)
	}
	panic("无法识别数据库类型：" + dbContext.dbConfig.DataType)
}
