package data

import (
	"database/sql"
	"github.com/farseer-go/fs/container"
	"github.com/farseer-go/fs/core"
	"github.com/farseer-go/fs/flog"
	"github.com/farseer-go/fs/types"
	"github.com/timandy/routine"
	"gorm.io/gorm"
	"reflect"
	"strings"
)

// 实现同一个协程下的事务作用域
var routineOrmClient = routine.NewInheritableThreadLocal[*gorm.DB]()

// InternalDbContext 数据库上下文
type InternalDbContext struct {
	dbConfig       *dbConfig          // 数据库配置
	IsolationLevel sql.IsolationLevel // 事务等级
}

// NewContext 数据库上下文初始化
// dbName：数据库配置名称
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

	dbContext := container.Resolve[core.ITransaction](dbName).(*InternalDbContext)
	contextValueOf := reflect.ValueOf(repositoryContext).Elem()

	// 遍历上下文中的TableSet字段类型
	for i := 0; i < contextValueOf.NumField(); i++ {
		field := contextValueOf.Field(i)
		if field.CanSet() {
			// 找到TableSet字段类型
			_, isDataTableSet := types.IsDataTableSet(field)
			if isDataTableSet {
				data := contextValueOf.Type().Field(i).Tag.Get("data")
				var tableName string
				if strings.HasPrefix(data, "name=") {
					tableName = data[len("name="):]
				}

				// 再取tableSet的子属性，并设置值
				field.Addr().MethodByName("Init").Call([]reflect.Value{reflect.ValueOf(dbContext), reflect.ValueOf(tableName), reflect.ValueOf(autoCreateTable)})
			} else if field.Type().String() == "core.ITransaction" {
				field.Set(reflect.ValueOf(dbContext))
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

// Begin 开启事务
func (receiver *InternalDbContext) Begin(isolationLevels ...sql.IsolationLevel) {
	if routineOrmClient.Get() == nil {
		ormClient, err := open(receiver.dbConfig)
		if err != nil {
			_ = flog.Error(err)
			return
		}

		// 事务等级
		isolationLevel := sql.LevelDefault
		if len(isolationLevels) > 0 {
			isolationLevel = isolationLevels[0]
		}
		// 开启事务
		ormClient = ormClient.Session(&gorm.Session{}).Begin(&sql.TxOptions{
			Isolation: isolationLevel,
		})
		routineOrmClient.Set(ormClient)
	}
}

// Transaction 使用事务
func (receiver *InternalDbContext) Transaction(executeFn func()) {
	receiver.Begin()
	executeFn()
	receiver.Commit()
}

// Commit 事务提交
func (receiver *InternalDbContext) Commit() {
	routineOrmClient.Get().Commit()
	routineOrmClient.Remove()
}

// Rollback 事务回滚
func (receiver *InternalDbContext) Rollback() {
	routineOrmClient.Get().Rollback()
	routineOrmClient.Remove()
}
