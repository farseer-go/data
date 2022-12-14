package test

import (
	"github.com/farseer-go/collections"
	"github.com/farseer-go/data"
	"github.com/farseer-go/fs"
	"github.com/farseer-go/fs/configure"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestTableSet(t *testing.T) {
	// 设置配置默认值，模拟配置文件
	configure.SetDefault("Database.test", "DataType=MySql,PoolMaxSize=50,PoolMinSize=1,ConnectionString=root:steden@123@tcp(mysql:3306)/test?charset=utf8&parseTime=True&loc=Local")
	data.Module{}.Shutdown()
	fs.Initialize[data.Module]("test data")
	var context TestMysqlContext
	data.InitContext(&context, "test")

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
			IsEnable:  true,
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
			IsEnable:  false,
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
		assert.Equal(t, 0, lst.First().Specialty.Count())
		assert.Less(t, 1, lst.First().Id)
	})

	// 测试排序
	t.Run("asc", func(t *testing.T) {
		lst := context.User.Asc("Age").ToList()
		assert.Equal(t, 2, lst.Count())
		assert.Equal(t, "harlen", lst.First().Name)
		assert.Equal(t, 34, lst.First().Age)
		assert.Equal(t, 1, lst.First().Attribute.Count())
		assert.Equal(t, "10", lst.First().Attribute.GetValue("work-year"))
		assert.Equal(t, "lee", lst.First().Fullname.FirstName)
		assert.Equal(t, "harlen", lst.First().Fullname.LastName)
		assert.Equal(t, Woman, lst.First().Gender)
		assert.Equal(t, 2, lst.First().Specialty.Count())
		assert.True(t, lst.First().Specialty.Contains("go"))
		assert.True(t, lst.First().Specialty.Contains("net"))
		assert.False(t, lst.First().IsEnable)
		assert.Less(t, 1, lst.First().Id)
	})

	t.Run("desc", func(t *testing.T) {
		lst := context.User.Desc("Age").ToList()
		assert.Equal(t, 2, lst.Count())
		assert.Equal(t, "steden", lst.First().Name)
		assert.Equal(t, 36, lst.First().Age)
		assert.Equal(t, 1, lst.First().Attribute.Count())
		assert.Equal(t, "15", lst.First().Attribute.GetValue("work-year"))
		assert.Equal(t, "he", lst.First().Fullname.FirstName)
		assert.Equal(t, "steden", lst.First().Fullname.LastName)
		assert.Equal(t, Man, lst.First().Gender)
		assert.Equal(t, 2, lst.First().Specialty.Count())
		assert.True(t, lst.First().Specialty.Contains("go"))
		assert.True(t, lst.First().Specialty.Contains("net"))
		assert.True(t, lst.First().IsEnable)
		assert.Less(t, 1, lst.First().Id)
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
		assert.True(t, lst.First().Specialty.Contains("go"))
		assert.True(t, lst.First().Specialty.Contains("net"))
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

	t.Run("GetFloat64", func(t *testing.T) {
		assert.True(t, context.User.Where("Name = ?", "steden").GetBool("Is_Enable"))
		assert.False(t, context.User.Where("Name = ?", "harlen").GetBool("Is_Enable"))
	})

	t.Run("UpdateValue", func(t *testing.T) {
		context.User.Where("Name = ?", "steden").UpdateValue("age", 18)
		user := context.User.Where("Name = ?", "steden").ToEntity()
		assert.Equal(t, 18, user.Age)

		context.User.Where("Name = ?", "steden").UpdateValue("Fullname", FullNameVO{
			FirstName: "lao_he",
			LastName:  "niao",
		})

		user = context.User.Where("Name = ?", "steden").ToEntity()
		assert.Equal(t, "lao_he", user.Fullname.FirstName)
		assert.Equal(t, "niao", user.Fullname.LastName)
	})

	t.Run("Update", func(t *testing.T) {
		user := context.User.Where("Name = ?", "steden").ToEntity()
		user.Age = 15
		user.Fullname = FullNameVO{
			FirstName: "lao",
			LastName:  "siji",
		}
		user.IsEnable = false

		context.User.Where("Name = ?", "steden").Update(user)
		user = context.User.Where("Name = ?", "steden").ToEntity()
		assert.Equal(t, 15, user.Age)
		assert.Equal(t, "lao", user.Fullname.FirstName)
		assert.Equal(t, "siji", user.Fullname.LastName)
		assert.Equal(t, false, user.IsEnable)
	})
}
