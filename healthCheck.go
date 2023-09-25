package data

import (
	"fmt"
	"time"
)

type healthCheck struct {
	name string
	IInternalContext
}

func (c *healthCheck) Check() (string, error) {
	InitContext(c, c.name, false)
	var dbAt time.Time
	tx := c.Original().Raw("select now()").Scan(&dbAt)
	return fmt.Sprintf("Database.%s => %s", c.name, dbAt.Format(time.DateTime)), tx.Error
}
