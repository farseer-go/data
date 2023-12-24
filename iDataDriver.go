package data

import "gorm.io/gorm"

type IDataDriver interface {
	GetDriver(connectionString string) gorm.Dialector
}
