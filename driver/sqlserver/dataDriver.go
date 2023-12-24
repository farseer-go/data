package data_sqlserver

import (
	"gorm.io/driver/sqlserver"
	"gorm.io/gorm"
)

type dataDriver struct {
}

func (receiver *dataDriver) GetDriver(connectionString string) gorm.Dialector {
	// sqlserver://user:123456@127.0.0.1:9930?database=dbname
	return sqlserver.Open(connectionString)
}
