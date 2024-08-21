package test

import (
	"github.com/farseer-go/data"
	"github.com/farseer-go/fs"
	"github.com/farseer-go/fs/configure"
)

func init() {
	// 设置配置默认值，模拟配置文件
	configure.SetDefault("Database.test", "DataType=MySql,PoolMaxSize=50,PoolMinSize=1,ConnectionString=root:steden@123@tcp(192.168.1.8:3306)/farseer_test?charset=utf8&parseTime=True&loc=Local")
	fs.Initialize[data.Module]("test data")
}
