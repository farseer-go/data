package data

import (
	"fmt"
	"time"

	"github.com/farseer-go/fs/parse"
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

// OptimizeFinal 手动执行合并（支持 string 或 time.Time 类型）
func (receiver *mergeTreeSet) OptimizeFinal(partition any) (int64, error) {
	var p string
	switch v := partition.(type) {
	case time.Time:
		p = v.Format("200601")
	default:
		p = parse.ToString(v)
	}

	query := fmt.Sprintf("OPTIMIZE TABLE %s PARTITION '%s' FINAL;", receiver.tableName, p)
	result := receiver.ormClient.Exec(query)
	return result.RowsAffected, result.Error
}

// OptimizeFinal 手动执行合并
func (receiver *mergeTreeSet) OptimizeFinalAll() (int64, error) {
	result := receiver.ormClient.Exec(fmt.Sprintf("OPTIMIZE TABLE %s FINAL;", receiver.tableName))
	return result.RowsAffected, result.Error
}
