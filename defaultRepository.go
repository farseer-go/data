package data

import (
	"github.com/farseer-go/collections"
	"github.com/farseer-go/mapper"
)

type DefaultRepository[TPoType any, TDomainObject any] struct {
	primaryName string
	table       TableSet[TPoType]
}

func NewDefaultRepository[TPoType any, TDomainObject any](table TableSet[TPoType]) IRepository[TDomainObject] {
	return &DefaultRepository[TPoType, TDomainObject]{primaryName: table.primaryName, table: table}
}

func (receiver *DefaultRepository[TPoType, TDomainObject]) ToEntity(id any) TDomainObject {
	po := receiver.table.Where(receiver.primaryName, id).ToEntity()
	// po 转 do
	return mapper.Single[TDomainObject](&po)
}

func (receiver *DefaultRepository[TPoType, TDomainObject]) Add(entity TDomainObject) error {
	po := mapper.Single[TPoType](&entity)
	return receiver.table.Insert(&po)
}

func (receiver *DefaultRepository[TPoType, TDomainObject]) AddList(lst collections.List[TDomainObject], batchSize int) (int64, error) {
	var lstPO collections.List[TPoType]
	lst.Select(&lstPO, func(entity TDomainObject) any {
		return mapper.Single[TPoType](&entity)
	})
	return receiver.table.InsertList(lstPO, batchSize)
}

func (receiver *DefaultRepository[TPoType, TDomainObject]) ToList() collections.List[TDomainObject] {
	// 从数据库读数据
	lstProduct := receiver.table.ToList()
	// po 转 do
	return mapper.ToList[TDomainObject](lstProduct)
}

func (receiver *DefaultRepository[TPoType, TDomainObject]) ToPageList(pageSize, pageIndex int) collections.PageList[TDomainObject] {
	// 从数据库读数据
	lstOrder := receiver.table.Desc(receiver.primaryName).ToPageList(pageSize, pageIndex)

	// po 转 do
	return mapper.ToPageList[TDomainObject](lstOrder)
}

func (receiver *DefaultRepository[TPoType, TDomainObject]) Count() int64 {
	count := receiver.table.Count()
	return count
}

func (receiver *DefaultRepository[TPoType, TDomainObject]) Update(id any, do TDomainObject) (int64, error) {
	po := mapper.Single[TPoType](do)
	return receiver.table.Where(receiver.primaryName, id).Update(po)
}

func (receiver *DefaultRepository[TPoType, TDomainObject]) Delete(id any) (int64, error) {
	return receiver.table.Where(receiver.primaryName, id).Delete()
}

func (receiver *DefaultRepository[TPoType, TDomainObject]) IsExists(id any) bool {
	return receiver.table.Where(receiver.primaryName, id).IsExists()
}
