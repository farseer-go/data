package repository

import (
	"github.com/farseer-go/data"
	"github.com/farseer-go/fs/container"
)

// RegisterRepository 注册通用的仓储服务
func RegisterRepository[TDomainObject any, TPoType any](table data.TableSet[TPoType]) {
	container.Register(func() IRepository[TDomainObject] {
		return &DefaultRepository[TDomainObject, TPoType]{primaryName: table.GetPrimaryName(), table: table}
	})
}
