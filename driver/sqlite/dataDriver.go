package data_sqlite

import (
	"fmt"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"strings"
)

type dataDriver struct {
}

func (receiver *dataDriver) GetDriver(connectionString string) gorm.Dialector {
	// gorm.db
	return sqlite.Open(connectionString)
}

func (receiver *dataDriver) CreateIndex(tableName, idxName string, fieldsName ...string) string {
	return fmt.Sprintf("CREATE INDEX %s ON %s (%s);", idxName, tableName, strings.Join(fieldsName, ","))
}
