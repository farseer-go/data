package test

import (
	"database/sql/driver"
	"errors"
	"fmt"
	"testing"

	"github.com/farseer-go/collections"
	"github.com/farseer-go/data"
	"github.com/farseer-go/data/decimal"
	"github.com/farseer-go/fs/snc"
	"github.com/stretchr/testify/assert"
)

type TestMysqlContext struct {
	User data.TableSet[UserPO] `data:"name=user;migrate"`
}

type UserPO struct {
	Id        int                                    `gorm:"primaryKey"`
	Name      string                                 // 用户名称
	Age       int                                    // 用户年龄
	Fullname  FullNameVO                             // 用户全称
	Specialty collections.List[string]               // 特长
	Attribute collections.Dictionary[string, string] // 自定义属性
	Gender    GenderType                             // 性别
	IsEnable  bool                                   // 是否启用
	Weight    decimal.Decimal                        // 体重
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
	ba, err := snc.Marshal(receiver)
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
	err := snc.Unmarshal(ba, &t)
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
			data.NewContext[TestMysqlContext]("")
		})
	})

	t.Run("NotSetConfig", func(t *testing.T) {
		assert.Panics(t, func() {
			data.NewContext[TestMysqlContext]("test2")
		})
	})

	t.Run("NewContext", func(t *testing.T) {
		context := data.NewContext[TestMysqlContext]("test")
		assert.Equal(t, "user", context.User.GetTableName())
	})
}

func TestInitContext(t *testing.T) {
	var context TestMysqlContext
	t.Run("zero value", func(t *testing.T) {
		data.InitContext(&context, "test")
		assert.Equal(t, "user", context.User.GetTableName())
	})

	t.Run("have value", func(t *testing.T) {
		data.InitContext(&context, "test")
		assert.Equal(t, "user", context.User.GetTableName())
	})

	t.Run("ptr", func(t *testing.T) {
		context2 := new(TestMysqlContext)
		data.InitContext(context2, "test")
		assert.Equal(t, "user", context2.User.GetTableName())

		data.InitContext(context2, "test")
		assert.Equal(t, "user", context2.User.GetTableName())
		assert.Equal(t, "user2", context2.User.SetTableName("user2").GetTableName())
	})
}

func Test_checkConfig(t *testing.T) {
	t.Run("unknownDataType", func(t *testing.T) {
		assert.Panics(t, func() {
			data.RegisterInternalContext("Database.test_oracle", "DataType=oracle,PoolMaxSize=5,PoolMinSize=1,ConnectionString=root:steden@123@tcp(192.168.1.8:3306)/test?charset=utf8&parseTime=True&loc=Local")
			data.NewContext[TestMysqlContext]("test_oracle").User.Count()
		})
	})

	t.Run("postgresql", func(t *testing.T) {
	})

	t.Run("sqlite", func(t *testing.T) {
	})

	t.Run("sqlserver", func(t *testing.T) {
	})
}

func TestExecuteSql(t *testing.T) {
	var context TestMysqlContext
	data.InitContext(&context, "test")
}
