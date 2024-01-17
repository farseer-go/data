package mysql

import (
	"bytes"
	"fmt"
	"github.com/farseer-go/data"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

type DataDriver struct {
}

func (receiver *DataDriver) GetDriver(connectionString string) gorm.Dialector {
	// user:123456@tcp(127.0.0.1:3306)/dbname?charset=utf8mb4&parseTime=True&loc=Local
	return mysql.Open(connectionString)
}

func (receiver *DataDriver) CreateIndex(tableName string, idxField data.IdxField) string {
	var b bytes.Buffer
	b.WriteString("CREATE ")
	if idxField.IsUNIQUE {
		b.WriteString("UNIQUE ")
	}
	b.WriteString(fmt.Sprintf("INDEX %s ON %s (%s);", idxField.IdxName, tableName, idxField.Fields))
	return b.String()
}
