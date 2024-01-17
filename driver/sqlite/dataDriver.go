package data_sqlite

import (
	"bytes"
	"fmt"
	"github.com/farseer-go/data"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type dataDriver struct {
}

func (receiver *dataDriver) GetDriver(connectionString string) gorm.Dialector {
	// gorm.db
	return sqlite.Open(connectionString)
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
