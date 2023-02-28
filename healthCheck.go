package data

type healthCheck struct {
	name  string
	Table TableSet[struct{}]
}

func (c *healthCheck) Check() (string, error) {
	InitContext(c, c.name)
	c.Table.open()
	defer c.Table.close()
	return "Database." + c.name, c.Table.err
}
