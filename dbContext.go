package data

import (
	"github.com/farseer-go/fs/configure"
	"github.com/farseer-go/fs/types"
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
	dbConfig := configure.ParseString[dbConfig](configString)
	dbContext := &DbContext{
		dbConfig: &dbConfig,
	}
	dbContext.dbConfig.dbName = dbName
	return dbContext
}

// NewContext 数据库上下文初始化
// dbName：数据库配置名称
func NewContext[TDbContext any](dbName string, autoCreateTable bool) *TDbContext {
	var context TDbContext
	InitContext(&context, dbName, autoCreateTable)
	return &context
}

// InitContext 数据库上下文初始化
// dbName：数据库配置名称
func InitContext[TDbContext any](dbContext *TDbContext, dbName string, autoCreateTable bool) {
	if dbName == "" {
		panic("dbName入参必须设置有效的值")
	}
	dbConfig := initConfig(dbName) // 嵌入类型
	contextValueOf := reflect.ValueOf(dbContext).Elem()

	for i := 0; i < contextValueOf.NumField(); i++ {
		field := contextValueOf.Field(i)
		if field.CanSet() {
			_, isDataTableSet := types.IsDataTableSet(field)
			if isDataTableSet {
				data := contextValueOf.Type().Field(i).Tag.Get("data")
				var tableName string
				if strings.HasPrefix(data, "name=") {
					tableName = data[len("name="):]
				}

				// 再取tableSet的子属性，并设置值
				field.Addr().MethodByName("Init").Call([]reflect.Value{reflect.ValueOf(dbConfig), reflect.ValueOf(tableName), reflect.ValueOf(autoCreateTable)})
			}
		}
	}
}
