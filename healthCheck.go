package data

type healthCheck struct {
	name  string
	Table TableSet[struct{}]
}

func (c *healthCheck) Check() (string, error) {
	InitContext(c, c.name)
	c.Table.open()
	defer c.Table.close()
	return "数据库 " + c.name, c.Table.err
}
