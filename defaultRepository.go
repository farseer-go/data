package data

import (
	"github.com/farseer-go/collections"
	"github.com/farseer-go/mapper"
)

type DefaultRepository[TDomainObject any, TPoType any] struct {
	primaryName string
	table       TableSet[TPoType]
}

func NewDefaultRepository[TDomainObject any, TPoType any](table TableSet[TPoType]) IRepository[TDomainObject] {
	return &DefaultRepository[TDomainObject, TPoType]{primaryName: table.GetPrimaryName(), table: table}
}

func (receiver *DefaultRepository[TDomainObject, TPoType]) ToEntity(id any) TDomainObject {
	po := receiver.table.Where(receiver.primaryName, id).ToEntity()
	// po 转 do
	return mapper.Single[TDomainObject](&po)
}

func (receiver *DefaultRepository[TDomainObject, TPoType]) Add(entity TDomainObject) {
	po := mapper.Single[TPoType](&entity)
	_ = receiver.table.Insert(&po)
}

func (receiver *DefaultRepository[TDomainObject, TPoType]) ToList() collections.List[TDomainObject] {
	// 从数据库读数据
	lstProduct := receiver.table.ToList()
	// po 转 do
	return mapper.ToList[TDomainObject](lstProduct)
}

func (receiver *DefaultRepository[TDomainObject, TPoType]) ToPageList(pageSize, pageIndex int) collections.PageList[TDomainObject] {
	// 从数据库读数据
	lstOrder := receiver.table.Desc(receiver.primaryName).ToPageList(pageSize, pageIndex)

	// po 转 do
	var lst collections.PageList[TDomainObject]
	lstOrder.MapToPageList(&lst)
	return lst
}

func (receiver *DefaultRepository[TDomainObject, TPoType]) Count() int64 {
	return receiver.table.Count()
}

func (receiver *DefaultRepository[TDomainObject, TPoType]) Update(id any, do TDomainObject) int64 {
	po := mapper.Single[TPoType](&do)
	return receiver.table.Where(receiver.primaryName, id).Update(po)
}
