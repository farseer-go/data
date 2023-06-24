package data

import (
	"github.com/farseer-go/fs/container"
	"github.com/farseer-go/fs/core"
	"github.com/farseer-go/fs/types"
	"reflect"
	"strings"
)

// IDbContext 数据库上下文
type IDbContext interface{}

// NewContext 数据库上下文初始化
// dbName：数据库配置名称，对应./farseer.yaml 中的 Database节点
// autoCreateTable：true表示自动创建表
// 同一个上下文生命周期内，共享一个orm client
func NewContext[TDbContext IDbContext](dbName string, autoCreateTable bool) *TDbContext {
	var context TDbContext
	InitContext(&context, dbName, autoCreateTable)
	return &context
}

// InitContext 数据库上下文初始化
// dbName：数据库配置名称
// autoCreateTable：true表示自动创建表
// 同一个上下文生命周期内，共享一个orm client
func InitContext[TDbContext IDbContext](repositoryContext *TDbContext, dbName string, autoCreateTable bool) {
	if dbName == "" {
		panic("dbName入参必须设置有效的值")
	}

	internalContextIns := container.Resolve[core.ITransaction](dbName).(*internalContext)
	internalContextType := reflect.ValueOf(internalContextIns)
	contextValueOf := reflect.ValueOf(repositoryContext).Elem()

	// 遍历上下文中的TableSet字段类型
	for i := 0; i < contextValueOf.NumField(); i++ {
		field := contextValueOf.Field(i)
		if field.CanSet() {
			// 找到TableSet字段类型
			_, isDataTableSet := types.IsDataTableSet(field)
			_, isDataDomainSet := types.IsDataDomainSet(field)
			if isDataTableSet || isDataDomainSet {
				data := contextValueOf.Type().Field(i).Tag.Get("data")
				var tableName string
				if strings.HasPrefix(data, "name=") {
					tableName = data[len("name="):]
				}

				// 再取tableSet的子属性，并设置值
				field.Addr().MethodByName("Init").Call([]reflect.Value{internalContextType, reflect.ValueOf(tableName), reflect.ValueOf(autoCreateTable)})
			} else if field.Type().String() == "core.ITransaction" || field.Type().String() == "data.IInternalContext" {
				field.Set(internalContextType)
			}
		}
	}
}

// RegisterContext 注册上下文（临时生命周期）
func RegisterContext[TDbContext IDbContext](dbName string, autoCreateTable bool) {
	container.RegisterTransient(func() IDbContext {
		var context TDbContext
		InitContext(&context, dbName, autoCreateTable)
		return &context
	}, dbName)
}

// GetContext 获取上下文实例（每次获取都会创建一个实例）
func GetContext[TDbContext IDbContext](dbName string) *TDbContext {
	return container.Resolve[IDbContext](dbName).(*TDbContext)
}
