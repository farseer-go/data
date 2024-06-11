module github.com/farseer-go/data

go 1.21

toolchain go1.22.0

require (
	github.com/farseer-go/collections v0.13.0
	github.com/farseer-go/fs v0.14.0
	github.com/farseer-go/mapper v0.13.0
	github.com/shopspring/decimal v1.4.0
	github.com/stretchr/testify v1.8.4
	gorm.io/driver/mysql v1.5.6
	gorm.io/gorm v1.25.10
)

require (
	filippo.io/edwards25519 v1.1.0 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/go-sql-driver/mysql v1.8.1 // indirect
	github.com/jinzhu/inflection v1.0.0 // indirect
	github.com/jinzhu/now v1.1.5 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/timandy/routine v1.1.2 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

exclude github.com/timandy/routine v1.1.3
