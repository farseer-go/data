package data_postgres

import (
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type dataDriver struct {
}

func (receiver *dataDriver) GetDriver(connectionString string) gorm.Dialector {
	// host=127.0.0.1 user=user password=123456 dbname=dbname port=9920 sslmode=disable TimeZone=Asia/Shanghai
	return postgres.Open(connectionString)
}
