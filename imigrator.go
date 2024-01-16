package data

type IMigratorCreate interface {
	// CreateTable 创建表
	CreateTable() string
}
type IMigratorIndex interface {
	// CreateIndex 创建索引
	CreateIndex() map[string][]string
}
