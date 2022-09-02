package data

import (
	"github.com/farseer-go/fs/configure"
	"github.com/stretchr/testify/assert"
	"testing"
)

type TestMysqlContext struct {
	User TableSet[UserPO] `data:"name=user"`
}

type UserPO struct {
	Id int `gorm:"primaryKey"`
	// 用户名称
	Name string
	// 用户年龄
	Age int
}

func TestNewContext(t *testing.T) {
	// 设置配置默认值，模拟配置文件
	configure.SetDefault("Database.test", "DataType=MySql,PoolMaxSize=50,PoolMinSize=1,ConnectionString=root:steden@123@tcp(mysql:3306)/test?charset=utf8&parseTime=True&loc=Local")

	context := NewContext[TestMysqlContext]("test")

	assert.Equal(t, "user", context.User.GetTableName())
}

func TestInitContext(t *testing.T) {
	// 设置配置默认值，模拟配置文件
	configure.SetDefault("Database.test", "DataType=MySql,PoolMaxSize=50,PoolMinSize=1,ConnectionString=root:steden@123@tcp(mysql:3306)/test?charset=utf8&parseTime=True&loc=Local")

	var context TestMysqlContext
	InitContext(&context, "test")
	assert.Equal(t, "user", context.User.GetTableName())

	InitContext(&context, "test")
	assert.Equal(t, "user", context.User.GetTableName())

	context2 := new(TestMysqlContext)
	InitContext(context2, "test")
	assert.Equal(t, "user", context2.User.GetTableName())

	InitContext(context2, "test")
	assert.Equal(t, "user", context2.User.GetTableName())
}
