package data

import "github.com/farseer-go/fs/container"

// DomainSet 比TableSet支持自动绑定领域层的聚合，实现通用的CRUD操作
type DomainSet[Table any, TDomainObject any] struct {
	TableSet[Table]
}

// Init 在反射的时候会调用此方法
func (r *DomainSet[Table, TDomainObject]) Init(dbContext *InternalDbContext, tableName string, autoCreateTable bool) {
	r.TableSet.Init(dbContext, tableName, autoCreateTable)

	// 注册通用的仓储服务
	if !container.IsRegister[IRepository[TDomainObject]]() {
		container.Register(func() IRepository[TDomainObject] {
			return &DefaultRepository[TDomainObject, Table]{primaryName: r.GetPrimaryName(), table: r.TableSet}
		})
	}
}
