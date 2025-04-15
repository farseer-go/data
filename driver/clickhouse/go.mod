module github.com/farseer-go/data/driver/clickhouse

go 1.23.0

toolchain go1.23.3

require (
	github.com/farseer-go/data v0.16.5
	github.com/farseer-go/fs v0.16.6
	gorm.io/driver/clickhouse v0.6.1
	gorm.io/gorm v1.25.12
)

require (
	filippo.io/edwards25519 v1.1.0 // indirect
	github.com/ClickHouse/ch-go v0.65.1 // indirect
	github.com/ClickHouse/clickhouse-go/v2 v2.34.0 // indirect
	github.com/andybalholm/brotli v1.1.1 // indirect
	github.com/farseer-go/collections v0.16.4 // indirect
	github.com/farseer-go/mapper v0.16.5 // indirect
	github.com/go-faster/city v1.0.1 // indirect
	github.com/go-faster/errors v0.7.1 // indirect
	github.com/go-sql-driver/mysql v1.9.2 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/govalues/decimal v0.1.36 // indirect
	github.com/hashicorp/go-version v1.7.0 // indirect
	github.com/jinzhu/inflection v1.0.0 // indirect
	github.com/jinzhu/now v1.1.5 // indirect
	github.com/json-iterator/go v1.1.12 // indirect
	github.com/klauspost/compress v1.18.0 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.2 // indirect
	github.com/paulmach/orb v0.11.1 // indirect
	github.com/pierrec/lz4/v4 v4.1.22 // indirect
	github.com/segmentio/asm v1.2.0 // indirect
	github.com/shopspring/decimal v1.4.0 // indirect
	github.com/timandy/routine v1.1.5 // indirect
	go.opentelemetry.io/otel v1.35.0 // indirect
	go.opentelemetry.io/otel/trace v1.35.0 // indirect
	golang.org/x/sys v0.32.0 // indirect
	golang.org/x/text v0.24.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
	gorm.io/driver/mysql v1.5.7 // indirect
	gorm.io/hints v1.1.2 // indirect
)

// 原库404
exclude github.com/mitchellh/osext v0.0.0-20151018003038-5e2d6d41470f

// 使用支持string decimal的库
//replace github.com/ClickHouse/clickhouse-go/v2 v2.33.1 => github.com/farseer-go/clickhouse-go/v2 v2.0.0
