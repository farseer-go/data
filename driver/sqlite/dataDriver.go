package data_sqlite

import (
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type dataDriver struct {
}

func (receiver *dataDriver) GetDriver(connectionString string) gorm.Dialector {
	// gorm.db
	return sqlite.Open(connectionString)
}
