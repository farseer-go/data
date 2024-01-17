package data

import "gorm.io/gorm"

type IDataDriver interface {
	GetDriver(connectionString string) gorm.Dialector
	// CreateIndex 创建索引的SQL
	CreateIndex(tableName string, idxField IdxField) string
}
