package data

import (
	"github.com/farseer-go/collections"
	"github.com/farseer-go/fs"
	"github.com/farseer-go/fs/configure"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestTableSet_ToList(t *testing.T) {
	// 设置配置默认值，模拟配置文件
	configure.SetDefault("Database.test", "DataType=MySql,PoolMaxSize=50,PoolMinSize=1,ConnectionString=root:steden@123@tcp(mysql:3306)/test?charset=utf8&parseTime=True&loc=Local")
	Module{}.Shutdown()
	fs.Initialize[Module]("test data")
	var context TestMysqlContext
	InitContext(&context, "test")

	// 先清空数据
	context.User.Where("1=1").Delete()
	// 此时的数据量应该为0
	assert.Equal(t, int64(0), context.User.Count())

	context.User.Insert(&UserPO{
		Name: "steden",
		Age:  36,
		Fullname: FullNameVO{
			FirstName: "he",
			LastName:  "steden",
		},
		Specialty: collections.NewList("go", "net"),
		Attribute: collections.NewDictionaryFromMap(map[string]string{"work-year": "15"}),
		Gender:    Man,
	})

	context.User.Insert(&UserPO{
		Name: "harlen",
		Age:  34,
		Fullname: FullNameVO{
			FirstName: "lee",
			LastName:  "harlen",
		},
		Specialty: collections.NewList("go", "net"),
		Attribute: collections.NewDictionaryFromMap(map[string]string{"work-year": "10"}),
		Gender:    Woman,
	})

	// 此时的数据量应该为2
	assert.Equal(t, int64(2), context.User.Count())
	assert.Equal(t, int64(2), context.User.SetTableName("user").Count())

	// 测试条件筛选、字段筛选
	list := context.User.Select("Age").Select("Name", "Id").Where("Age > ?", 34).Where("Name = ?", "steden").ToList()
	assert.Equal(t, 1, list.Count())
	assert.Equal(t, "steden", list.First().Name)
	assert.Equal(t, 36, list.First().Age)
	assert.Less(t, 1, list.First().Id)

	// 测试排序
	lst := context.User.Asc("Age").ToList()
	assert.Equal(t, 2, lst.Count())
	assert.Equal(t, "harlen", lst.First().Name)

	lst = context.User.Desc("Age").ToList()
	assert.Equal(t, 2, lst.Count())
	assert.Equal(t, "steden", lst.First().Name)

	lst = context.User.Limit(1).ToList()
	assert.Equal(t, 1, lst.Count())
}

func TestTableSet_Limit(t *testing.T) {

}
