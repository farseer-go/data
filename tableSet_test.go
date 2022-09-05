package data

import (
	"github.com/farseer-go/collections"
	"github.com/farseer-go/fs"
	"github.com/farseer-go/fs/configure"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestTableSet(t *testing.T) {
	// 设置配置默认值，模拟配置文件
	configure.SetDefault("Database.test", "DataType=MySql,PoolMaxSize=50,PoolMinSize=1,ConnectionString=root:steden@123@tcp(mysql:3306)/test?charset=utf8&parseTime=True&loc=Local")
	Module{}.Shutdown()
	fs.Initialize[Module]("test data")
	var context TestMysqlContext
	InitContext(&context, "test")

	t.Run("delete", func(t *testing.T) {
		// 先清空数据
		context.User.Where("1=1").Delete()
		// 此时的数据量应该为0
		assert.Equal(t, int64(0), context.User.Count())
	})

	t.Run("insert", func(t *testing.T) {
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
	})

	// 测试条件筛选、字段筛选
	t.Run("select", func(t *testing.T) {
		lst := context.User.Select("Age").Select("Name", "Id").Select([]string{"Name", "Id"}).Where("Age > ?", 34).Where("Name = ?", "steden").ToList()
		assert.Equal(t, 1, lst.Count())
		assert.Equal(t, "steden", lst.First().Name)
		assert.Equal(t, 36, lst.First().Age)
		assert.Equal(t, 0, lst.First().Attribute.Count())
		assert.Equal(t, "", lst.First().Fullname.FirstName)
		assert.Equal(t, "", lst.First().Fullname.LastName)
		assert.Equal(t, Man, lst.First().Gender)
		assert.Equal(t, 0, lst.First().Specialty.Count())
		assert.Less(t, 1, lst.First().Id)
	})

	// 测试排序
	t.Run("asc", func(t *testing.T) {
		lst := context.User.Asc("Age").ToList()
		assert.Equal(t, 2, lst.Count())
		assert.Equal(t, "harlen", lst.First().Name)
	})

	t.Run("desc", func(t *testing.T) {
		lst := context.User.Desc("Age").ToList()
		assert.Equal(t, 2, lst.Count())
		assert.Equal(t, "steden", lst.First().Name)
	})

	t.Run("limit", func(t *testing.T) {
		lst := context.User.Limit(1).ToList()
		assert.Equal(t, 1, lst.Count())
		assert.Equal(t, "steden", lst.First().Name)
		assert.Equal(t, 36, lst.First().Age)
		assert.Equal(t, 1, lst.First().Attribute.Count())
		assert.Equal(t, "he", lst.First().Fullname.FirstName)
		assert.Equal(t, "steden", lst.First().Fullname.LastName)
		assert.Equal(t, Man, lst.First().Gender)
		assert.Equal(t, 2, lst.First().Specialty.Count())
		assert.Less(t, 1, lst.First().Id)
	})

	t.Run("toArray", func(t *testing.T) {
		lst := context.User.ToArray()
		assert.Equal(t, 2, len(lst))
	})

	t.Run("ToPageList", func(t *testing.T) {
		lst := context.User.Where("Age > 10").Asc("Age").ToPageList(1, 1)
		assert.Equal(t, int64(2), lst.RecordCount)
		assert.Equal(t, 1, lst.List.Count())
		assert.Equal(t, "harlen", lst.List.First().Name)

		lst = context.User.Where("Age > 10").Asc("Age").ToPageList(1, 2)
		assert.Equal(t, int64(2), lst.RecordCount)
		assert.Equal(t, 1, lst.List.Count())
		assert.Equal(t, "steden", lst.List.First().Name)

		lst = context.User.Where("Age > 10").Desc("Age").ToPageList(1, 1)
		assert.Equal(t, int64(2), lst.RecordCount)
		assert.Equal(t, 1, lst.List.Count())
		assert.Equal(t, "steden", lst.List.First().Name)
	})

	t.Run("ToEntity", func(t *testing.T) {
		user := context.User.Where("Name = ?", "steden").Select("Id", "Name", "Age").ToEntity()
		assert.Equal(t, "steden", user.Name)
		assert.Equal(t, 36, user.Age)
		assert.Less(t, 1, user.Id)
		assert.Equal(t, 0, user.Attribute.Count())
		assert.Equal(t, "", user.Fullname.FirstName)
		assert.Equal(t, "", user.Fullname.LastName)
		assert.Equal(t, Man, user.Gender)
		assert.Equal(t, 0, user.Specialty.Count())
	})

	t.Run("IsExists", func(t *testing.T) {
		assert.True(t, context.User.Where("Name = ?", "steden").IsExists())
		assert.False(t, context.User.Where("Name = ?", "steden2").IsExists())
	})

	t.Run("GetString", func(t *testing.T) {
		assert.Equal(t, "steden", context.User.Where("Name = ?", "steden").GetString("Name"))
	})

	t.Run("GetInt", func(t *testing.T) {
		assert.Less(t, 1, context.User.Where("Name = ?", "steden").GetInt("Id"))
	})

	t.Run("GetLong", func(t *testing.T) {
		assert.Less(t, int64(1), context.User.Where("Name = ?", "steden").GetLong("Id"))
	})

	t.Run("GetFloat32", func(t *testing.T) {
		assert.Less(t, float32(1), context.User.Where("Name = ?", "steden").GetFloat32("Id"))
	})

	t.Run("GetFloat64", func(t *testing.T) {
		assert.Less(t, float64(1), context.User.Where("Name = ?", "steden").GetFloat64("Id"))
	})
}
