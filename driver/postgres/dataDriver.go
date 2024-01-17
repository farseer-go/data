package data_postgres

import (
	"bytes"
	"fmt"
	"github.com/farseer-go/data"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type dataDriver struct {
}

func (receiver *dataDriver) GetDriver(connectionString string) gorm.Dialector {
	// host=127.0.0.1 user=user password=123456 dbname=dbname port=9920 sslmode=disable TimeZone=Asia/Shanghai
	return postgres.Open(connectionString)
}

func (receiver *dataDriver) CreateIndex(tableName string, idxName string, idxField data.IdxField) string {
	var b bytes.Buffer
	b.WriteString("CREATE ")
	if idxField.IsUNIQUE {
		b.WriteString("UNIQUE ")
	}
	b.WriteString(fmt.Sprintf("INDEX %s ON %s (%s);", idxName, tableName, idxField.Fields))
	return b.String()
}
