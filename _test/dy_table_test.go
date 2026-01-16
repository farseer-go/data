package test

import (
	"github.com/farseer-go/collections"
	"github.com/farseer-go/data"
	"github.com/stretchr/testify/assert"
	"strconv"
	"testing"
)

// 动态创建表
func TestDyTable(t *testing.T) {
	var context TestMysqlContext
	data.InitContext(&context, "test")

	// 创建10张表
	for i := 0; i < 10; i++ {
		tableName := "user_" + strconv.Itoa(i)
		ts := context.User.SetTableName(tableName)
		ts.CreateTable("")
		ts.CreateIndex()

		ts.Insert(&UserPO{
			Name: "steden",
			Age:  36,
			Fullname: FullNameVO{
				FirstName: "he",
				LastName:  "steden",
			},
			Specialty: collections.NewList("go", "net"),
			Attribute: collections.NewDictionaryFromMap(map[string]string{"work-year": "15"}),
			Gender:    Man,
			IsEnable:  true,
		})
		ts.Insert(&UserPO{
			Name: "harlen",
			Age:  34,
			Fullname: FullNameVO{
				FirstName: "lee",
				LastName:  "harlen",
			},
			Specialty: collections.NewList("go", "net"),
			Attribute: collections.NewDictionaryFromMap(map[string]string{"work-year": "10"}),
			Gender:    Woman,
			IsEnable:  false,
		})

		// 此时的数据量应该为2
		assert.Equal(t, int64(2), ts.Count())
		assert.Equal(t, tableName, ts.GetTableName())

		_, err := context.User.ExecuteSql("drop table " + tableName)
		assert.Nil(t, err)
	}
}
