package data

type healthCheck struct {
	name  string
	Table TableSet[struct{}]
}

func (c *healthCheck) Check() (string, error) {
	InitContext(c, c.name, false)
	c.Table.getOrCreateSession()
	return "Database." + c.name, c.Table.err
}
