package data

type IdxField struct {
	IsUNIQUE bool   // 唯一索引
	Fields   string // 多个用逗号分隔
}

type IMigratorIndex interface {
	// CreateIndex 创建索引
	CreateIndex() map[string]IdxField
}
