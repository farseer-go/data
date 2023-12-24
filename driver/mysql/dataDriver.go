package mysql

import (
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

type DataDriver struct {
}

func (receiver *DataDriver) GetDriver(connectionString string) gorm.Dialector {
	// user:123456@tcp(127.0.0.1:3306)/dbname?charset=utf8mb4&parseTime=True&loc=Local
	return mysql.Open(connectionString)
}
