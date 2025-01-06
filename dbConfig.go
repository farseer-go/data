package data

import (
	"fmt"

	"github.com/farseer-go/fs/container"
	"gorm.io/gorm"
)

// 数据库配置
type dbConfig struct {
	keyName          string
	DataType         string
	PoolMaxSize      int
	PoolMinSize      int
	ConnectionString string
	databaseName     string // 数据库名称
	Migrate          string // code first
	migrated         bool   // 是否包含自动创建数据库
}

// GetDriver 获取对应驱动
func (receiver *dbConfig) GetDriver() gorm.Dialector {
	if !container.IsRegister[IDataDriver](receiver.DataType) {
		panic(fmt.Sprintf("要使用%s，请加载模块：对应的驱动，通常位置在：github.com/farseer-go/data/driver/%s", receiver.DataType, receiver.DataType))
	}
	return container.Resolve[IDataDriver](receiver.DataType).GetDriver(receiver.ConnectionString)
}
