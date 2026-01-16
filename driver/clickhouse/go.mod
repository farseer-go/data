module github.com/farseer-go/data/driver/clickhouse

go 1.24.0

require (
	github.com/farseer-go/data v0.17.3
	github.com/farseer-go/fs v0.17.3
	gorm.io/driver/clickhouse v0.7.0
	gorm.io/gorm v1.31.1
)

require (
	filippo.io/edwards25519 v1.1.0 // indirect
	github.com/ClickHouse/ch-go v0.69.0 // indirect
	github.com/ClickHouse/clickhouse-go/v2 v2.42.0 // indirect
	github.com/andybalholm/brotli v1.2.0 // indirect
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/farseer-go/collections v0.17.3 // indirect
	github.com/farseer-go/mapper v0.17.3 // indirect
	github.com/go-faster/city v1.0.1 // indirect
	github.com/go-faster/errors v0.7.1 // indirect
	github.com/go-sql-driver/mysql v1.9.3 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/govalues/decimal v0.1.36 // indirect
	github.com/hashicorp/go-version v1.8.0 // indirect
	github.com/jinzhu/inflection v1.0.0 // indirect
	github.com/jinzhu/now v1.1.5 // indirect
	github.com/json-iterator/go v1.1.12 // indirect
	github.com/klauspost/compress v1.18.3 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.2 // indirect
	github.com/paulmach/orb v0.12.0 // indirect
	github.com/pierrec/lz4/v4 v4.1.23 // indirect
	github.com/segmentio/asm v1.2.1 // indirect
	github.com/shopspring/decimal v1.4.0 // indirect
	github.com/timandy/routine v1.1.6 // indirect
	go.opentelemetry.io/otel v1.39.0 // indirect
	go.opentelemetry.io/otel/trace v1.39.0 // indirect
	go.yaml.in/yaml/v3 v3.0.4 // indirect
	golang.org/x/sys v0.40.0 // indirect
	golang.org/x/text v0.33.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
	gorm.io/driver/mysql v1.6.0 // indirect
	gorm.io/hints v1.1.2 // indirect
)

// 原库404
exclude github.com/mitchellh/osext v0.0.0-20151018003038-5e2d6d41470f

// 使用支持string decimal的库
//replace github.com/ClickHouse/clickhouse-go/v2 v2.33.1 => github.com/farseer-go/clickhouse-go/v2 v2.0.0
