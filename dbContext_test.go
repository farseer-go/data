package data

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/farseer-go/collections"
	"github.com/farseer-go/fs"
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
	// 是否启用
	IsEnable bool
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

type GenderType int

const (
	Man GenderType = iota
	Woman
)

func TestNewContext(t *testing.T) {
	t.Run("withoutDbName", func(t *testing.T) {
		assert.Panics(t, func() {
			NewContext[TestMysqlContext]("")
		})
	})

	t.Run("NotSetConfig", func(t *testing.T) {
		assert.Panics(t, func() {
			NewContext[TestMysqlContext]("test")
		})
	})

	t.Run("errorPwd", func(t *testing.T) {
		assert.Panics(t, func() {
			configure.SetDefault("Database.test", "DataType=MySql,PoolMaxSize=50,PoolMinSize=1,ConnectionString=root:steden@tcp(mysql:3306)/test?charset=utf8&parseTime=True&loc=Local")
			NewContext[TestMysqlContext]("test").User.Count()
		})
	})

	t.Run("NewContext", func(t *testing.T) {
		// 设置配置默认值，模拟配置文件
		configure.SetDefault("Database.test", "DataType=MySql,PoolMaxSize=50,PoolMinSize=1,ConnectionString=root:steden@123@tcp(mysql:3306)/test?charset=utf8&parseTime=True&loc=Local")
		context := NewContext[TestMysqlContext]("test")
		assert.Equal(t, "user", context.User.GetTableName())
	})
}

func TestInitContext(t *testing.T) {
	// 设置配置默认值，模拟配置文件
	configure.SetDefault("Database.test", "DataType=MySql,PoolMaxSize=50,PoolMinSize=1,ConnectionString=root:steden@123@tcp(mysql:3306)/test?charset=utf8&parseTime=True&loc=Local")

	var context TestMysqlContext
	t.Run("zero value", func(t *testing.T) {
		InitContext(&context, "test")
		assert.Equal(t, "user", context.User.GetTableName())
	})

	t.Run("have value", func(t *testing.T) {
		InitContext(&context, "test")
		assert.Equal(t, "user", context.User.GetTableName())
	})

	t.Run("ptr", func(t *testing.T) {
		context2 := new(TestMysqlContext)
		InitContext(context2, "test")
		assert.Equal(t, "user", context2.User.GetTableName())

		InitContext(context2, "test")
		assert.Equal(t, "user", context2.User.GetTableName())
		context2.User.SetTableName("user2")
		assert.Equal(t, "user2", context2.User.GetTableName())
	})
}

func Test_checkConfig(t *testing.T) {
	t.Run("empty", func(t *testing.T) {
		assert.Panics(t, func() {
			configure.SetDefault("Database.test", "")
			fs.Initialize[Module]("test data")
		})
	})

	t.Run("emptyConnection", func(t *testing.T) {
		assert.Panics(t, func() {
			configure.SetDefault("Database.test", "DataType=MySql,PoolMaxSize=50,PoolMinSize=1")
			fs.Initialize[Module]("test data")
		})
	})

	t.Run("emptyDataType", func(t *testing.T) {
		assert.Panics(t, func() {
			configure.SetDefault("Database.test", "PoolMaxSize=50,PoolMinSize=1,ConnectionString=root:steden@123@tcp(mysql:3306)/test?charset=utf8&parseTime=True&loc=Local")
			fs.Initialize[Module]("test data")
		})
	})

	t.Run("unknownDataType", func(t *testing.T) {
		assert.Panics(t, func() {
			configure.SetDefault("Database.test", "DataType=oracle,PoolMaxSize=50,PoolMinSize=1,ConnectionString=root:steden@123@tcp(mysql:3306)/test?charset=utf8&parseTime=True&loc=Local")
			NewContext[TestMysqlContext]("test").User.Count()
		})
	})

	t.Run("postgresql", func(t *testing.T) {
		//configure.SetDefault("Database.test", "DataType=postgresql,PoolMaxSize=50,PoolMinSize=1,ConnectionString=root:steden@123@tcp(mysql:3306)/test?charset=utf8&parseTime=True&loc=Local")
		//NewContext[TestMysqlContext]("test")
	})

	t.Run("sqlite", func(t *testing.T) {
		//configure.SetDefault("Database.test", "DataType=sqlite,PoolMaxSize=50,PoolMinSize=1,ConnectionString=root:steden@123@tcp(mysql:3306)/test?charset=utf8&parseTime=True&loc=Local")
		//NewContext[TestMysqlContext]("test")
	})

	t.Run("sqlserver", func(t *testing.T) {
		//configure.SetDefault("Database.test", "DataType=sqlserver,PoolMaxSize=50,PoolMinSize=1,ConnectionString=root:steden@123@tcp(mysql:3306)/test?charset=utf8&parseTime=True&loc=Local")
		//NewContext[TestMysqlContext]("test")
	})
}
