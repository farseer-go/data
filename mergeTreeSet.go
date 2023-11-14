package data

import (
	"fmt"
	"gorm.io/gorm"
)

type mergeTreeSet struct {
	ormClient *gorm.DB // 最外层的ormClient一定是nil的
	tableName string   // 表名
}

func newClickhouse[Table any](tableSet *TableSet[Table]) *mergeTreeSet {
	return &mergeTreeSet{
		ormClient: tableSet.ormClient,
		tableName: tableSet.tableName,
	}
}

// OptimizeFinal 手动执行合并
func (receiver *mergeTreeSet) OptimizeFinal() (int64, error) {
	result := receiver.ormClient.Exec(fmt.Sprintf("OPTIMIZE TABLE %s FINAL;", receiver.tableName))
	return result.RowsAffected, result.Error
}
