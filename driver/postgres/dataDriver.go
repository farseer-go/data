package data_postgres

import (
	"fmt"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"strings"
)

type dataDriver struct {
}

func (receiver *dataDriver) GetDriver(connectionString string) gorm.Dialector {
	// host=127.0.0.1 user=user password=123456 dbname=dbname port=9920 sslmode=disable TimeZone=Asia/Shanghai
	return postgres.Open(connectionString)
}

func (receiver *dataDriver) CreateIndex(tableName, idxName string, fieldsName ...string) string {
	return fmt.Sprintf("CREATE INDEX %s ON %s (%s);", idxName, tableName, strings.Join(fieldsName, ","))
}
