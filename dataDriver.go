package data

import (
	"bytes"
	"fmt"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

type DataDriver struct {
}

func (receiver *DataDriver) GetDriver(connectionString string) gorm.Dialector {
	// user:123456@tcp(127.0.0.1:3306)/dbname?charset=utf8mb4&parseTime=True&loc=Local
	return mysql.Open(connectionString)
}

func (receiver *DataDriver) CreateIndex(tableName string, idxName string, idxField IdxField) string {
	var b bytes.Buffer
	b.WriteString("CREATE ")
	if idxField.IsUNIQUE {
		b.WriteString("UNIQUE ")
	}
	b.WriteString(fmt.Sprintf("INDEX %s ON %s (%s);", idxName, tableName, idxField.Fields))
	return b.String()
}
