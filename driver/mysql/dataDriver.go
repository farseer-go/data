package mysql

import (
	"fmt"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"strings"
)

type DataDriver struct {
}

func (receiver *DataDriver) GetDriver(connectionString string) gorm.Dialector {
	// user:123456@tcp(127.0.0.1:3306)/dbname?charset=utf8mb4&parseTime=True&loc=Local
	return mysql.Open(connectionString)
}

func (receiver *DataDriver) CreateIndex(tableName, idxName string, fieldsName ...string) string {
	return fmt.Sprintf("CREATE INDEX %s ON %s (%s);", idxName, tableName, strings.Join(fieldsName, ","))
}
