package data

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/farseer-go/collections"
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
	// 用户全称
	Fullname FullNameVO
	// 特长
	Specialty collections.List[string]
	// 自定义属性
	Attribute collections.Dictionary[string, string]
	// 性别
	Gender GenderType
}

// 全称
type FullNameVO struct {
	// 姓氏
	FirstName string
	// 名称
	LastName string
}

// Value return json value, implement driver.Valuer interface
func (receiver FullNameVO) Value() (driver.Value, error) {
	ba, err := json.Marshal(receiver)
	return string(ba), err
}

// Scan scan value into Jsonb, implements sql.Scanner interface
func (receiver *FullNameVO) Scan(val any) error {
	if val == nil {
		*receiver = FullNameVO{}
		return nil
	}
	var ba []byte
	switch v := val.(type) {
	case []byte:
		ba = v
	case string:
		ba = []byte(v)
	default:
		return errors.New(fmt.Sprint("Failed to unmarshal JSONB value:", val))
	}

	t := FullNameVO{}
	err := json.Unmarshal(ba, &t)
	*receiver = t
	return err
}

// MarshalJSON to output non base64 encoded []byte
//func (receiver FullNameVO) MarshalJSON() ([]byte, error) {
//	return json.Marshal(receiver)
//}

// UnmarshalJSON to deserialize []byte
//func (receiver *FullNameVO) UnmarshalJSON(b []byte) error {
//	return json.Unmarshal(b, &receiver)
//}

type GenderType int

const (
	Man GenderType = iota
	Woman
)

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
	context2.User.SetTableName("user2")
	assert.Equal(t, "user2", context2.User.GetTableName())
}
