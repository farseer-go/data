package data

import (
	"go/ast"
	"reflect"

	"github.com/bytedance/sonic"
	"gorm.io/gorm/schema"
)

// ToMap PO实体转map
func ToMap(entity any) map[string]any {
	dic := make(map[string]any)
	dicValue := reflect.ValueOf(dic)
	fsVal := reflect.Indirect(reflect.ValueOf(entity))
	namingStrategy := schema.NamingStrategy{IdentifierMaxLength: 64}

	// 遍历字段
	for i := 0; i < fsVal.NumField(); i++ {
		var fieldName string
		// 取出当前字段
		if field := fsVal.Type().Field(i); ast.IsExported(field.Name) {
			fieldName = field.Name
			// 取出gorm标签
			fieldTags := schema.ParseTagSetting(field.Tag.Get("gorm"), ";")
			// 取出Column标签
			if colName, existsColumn := fieldTags["COLUMN"]; existsColumn {
				fieldName = colName
			} else {
				fieldName = namingStrategy.ColumnName("", field.Name)
			}

			// 取出json标签
			_, isJsonField := fieldTags["JSON"]
			if !isJsonField {
				jsonTag, existsJsonTag := fieldTags["SERIALIZER"]
				isJsonField = existsJsonTag && jsonTag == "json"
			}

			itemValue := fsVal.Field(i)
			if isJsonField {
				marshal, _ := sonic.Marshal(itemValue.Interface())
				itemValue = reflect.ValueOf(marshal)
			}
			dicValue.SetMapIndex(reflect.ValueOf(fieldName), itemValue)
		}
	}
	return dic
}
