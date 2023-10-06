package data

import (
	"go/ast"
	"gorm.io/gorm/schema"
	"reflect"
)

func ToMap(entity any) map[string]any {
	dic := make(map[string]any)
	dicValue := reflect.ValueOf(dic)
	fsVal := reflect.Indirect(reflect.ValueOf(entity))
	namingStrategy := schema.NamingStrategy{IdentifierMaxLength: 64}

	for i := 0; i < fsVal.NumField(); i++ {
		var fieldName string
		if field := fsVal.Type().Field(i); ast.IsExported(field.Name) {
			fieldName = field.Name
			fieldTags := schema.ParseTagSetting(field.Tag.Get("gorm"), ";")
			if c, existsColumn := fieldTags["COLUMN"]; existsColumn {
				fieldName = c
				continue
			}
			fieldName = namingStrategy.ColumnName("", field.Name)
		}
		itemValue := fsVal.Field(i)
		dicValue.SetMapIndex(reflect.ValueOf(fieldName), itemValue)
	}
	return dic
}
