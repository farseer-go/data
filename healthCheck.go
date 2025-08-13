package data

import (
	"fmt"
	"time"
)

type healthCheck struct {
	name     string
	dataType string // 数据库类型
	IInternalContext
}

func (c *healthCheck) Check() (string, error) {
	InitContext(c, c.name)
	var dbAt time.Time
	var sql string
	switch c.dataType {
	case "sqlserver", "mssql":
		sql = "SELECT GETDATE()"
	default:
		sql = "SELECT now()"
	}
	original, err := c.Original()
	if err != nil {
		return fmt.Sprintf("Database.%s", c.name), err
	}

	tx := original.Raw(sql).Scan(&dbAt)
	return fmt.Sprintf("Database.%s => %s", c.name, dbAt.Format("2006-01-02 15:04:05")), tx.Error
}
