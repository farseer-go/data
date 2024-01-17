package data

type IMigratorCreate interface {
	// CreateTable 创建表
	CreateTable() string
}
