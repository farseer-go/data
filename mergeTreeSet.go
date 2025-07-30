package data

import (
	"fmt"
	"time"

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

// OptimizeFinalByPartition 手动执行合并
func (receiver *mergeTreeSet) OptimizeFinalByPartition(partition string) (int64, error) {
	result := receiver.ormClient.Exec(fmt.Sprintf("OPTIMIZE TABLE %s PARTITION '%s' FINAL;", receiver.tableName, partition))
	return result.RowsAffected, result.Error
}

// OptimizeFinal 手动执行合并
func (receiver *mergeTreeSet) OptimizeFinal(partition time.Time) (int64, error) {
	result := receiver.ormClient.Exec(fmt.Sprintf("OPTIMIZE TABLE %s PARTITION '%s' FINAL;", receiver.tableName, partition.Format("200601")))
	return result.RowsAffected, result.Error
}

// OptimizeFinal 手动执行合并
func (receiver *mergeTreeSet) OptimizeFinalAll() (int64, error) {
	result := receiver.ormClient.Exec(fmt.Sprintf("OPTIMIZE TABLE %s FINAL;", receiver.tableName))
	return result.RowsAffected, result.Error
}
