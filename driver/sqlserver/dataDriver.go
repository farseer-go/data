package data_sqlserver

import (
	"fmt"
	"gorm.io/driver/sqlserver"
	"gorm.io/gorm"
	"strings"
)

type dataDriver struct {
}

func (receiver *dataDriver) GetDriver(connectionString string) gorm.Dialector {
	// sqlserver://user:123456@127.0.0.1:9930?database=dbname
	return sqlserver.Open(connectionString)
}

func (receiver *dataDriver) CreateIndex(tableName, idxName string, fieldsName ...string) string {
	return fmt.Sprintf("CREATE NONCLUSTERED INDEX %s ON %s (%s);", idxName, tableName, strings.Join(fieldsName, ","))
}
