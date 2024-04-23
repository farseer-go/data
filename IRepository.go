package data

import "github.com/farseer-go/collections"

// IRepository 通用的仓储接口，实现常用的CURD
type IRepository[TDomainObject any] interface {
	// ToEntity 查询实体
	ToEntity(id any) TDomainObject
	// Add 添加实体
	Add(entity TDomainObject) error
	// AddList 批量添加
	AddList(lst collections.List[TDomainObject], batchSize int) (int64, error)
	// AddIgnoreList 批量添加
	AddIgnoreList(lst collections.List[TDomainObject], batchSize int) (int64, error)
	// ToList 获取所有列表
	ToList() collections.List[TDomainObject]
	// ToPageList 分页列表
	ToPageList(pageSize, pageIndex int) collections.PageList[TDomainObject]
	// Count 数量
	Count() int64
	// Update 保存数据
	Update(id any, do TDomainObject) (int64, error)
	// Delete 删除数据
	Delete(id any) (int64, error)
	// IsExists 记录是否存在
	IsExists(id any) bool
}
