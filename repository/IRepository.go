package repository

import "github.com/farseer-go/collections"

// IRepository 通用的仓储接口，实现常用的CURD
type IRepository[TDomainObject any] interface {
	// ToEntity 查询实体
	ToEntity(id any) TDomainObject
	// Add 添加实体
	Add(entity TDomainObject)
	// ToList 获取所有列表
	ToList() collections.List[TDomainObject]
	// ToPageList 分页列表
	ToPageList(pageSize, pageIndex int) collections.PageList[TDomainObject]
	// Count 数量
	Count() int64
}
