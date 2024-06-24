module github.com/farseer-go/data/driver/clickhouse

go 1.21

toolchain go1.22.0

require (
	github.com/farseer-go/data v0.14.0
	github.com/farseer-go/fs v0.14.0
	gorm.io/driver/clickhouse v0.6.1
	gorm.io/gorm v1.25.10
)

// 原库404
exclude github.com/mitchellh/osext v0.0.0-20151018003038-5e2d6d41470f

//replace github.com/mitchellh/osext v0.0.0-20151018003038-5e2d6d41470f => github.com/farseer-go/osext v0.1.0

exclude github.com/timandy/routine v1.1.3

require (
	filippo.io/edwards25519 v1.1.0 // indirect
	github.com/ClickHouse/ch-go v0.61.5 // indirect
	github.com/ClickHouse/clickhouse-go/v2 v2.25.0 // indirect
	github.com/andybalholm/brotli v1.1.0 // indirect
	github.com/farseer-go/collections v0.14.0 // indirect
	github.com/farseer-go/mapper v0.14.0 // indirect
	github.com/go-faster/city v1.0.1 // indirect
	github.com/go-faster/errors v0.7.1 // indirect
	github.com/go-sql-driver/mysql v1.8.1 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/hashicorp/go-version v1.7.0 // indirect
	github.com/jinzhu/inflection v1.0.0 // indirect
	github.com/jinzhu/now v1.1.5 // indirect
	github.com/klauspost/compress v1.17.8 // indirect
	github.com/paulmach/orb v0.11.1 // indirect
	github.com/pierrec/lz4/v4 v4.1.21 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/segmentio/asm v1.2.0 // indirect
	github.com/shopspring/decimal v1.4.0 // indirect
	github.com/timandy/routine v1.1.2 // indirect
	go.opentelemetry.io/otel v1.27.0 // indirect
	go.opentelemetry.io/otel/trace v1.27.0 // indirect
	golang.org/x/sys v0.21.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
	gorm.io/driver/mysql v1.5.6 // indirect
)
