module github.com/farseer-go/data/driver/clickhouse

go 1.19

require (
	github.com/farseer-go/data v0.13.0
	github.com/farseer-go/fs v0.13.0
	gorm.io/driver/clickhouse v0.6.0
	gorm.io/gorm v1.25.5
)

// 原库404
exclude github.com/mitchellh/osext v0.0.0-20151018003038-5e2d6d41470f

//replace github.com/mitchellh/osext v0.0.0-20151018003038-5e2d6d41470f => github.com/farseer-go/osext v0.1.0

require (
	github.com/ClickHouse/ch-go v0.61.0 // indirect
	github.com/ClickHouse/clickhouse-go/v2 v2.17.1 // indirect
	github.com/andybalholm/brotli v1.0.6 // indirect
	github.com/farseer-go/collections v0.13.0 // indirect
	github.com/farseer-go/mapper v0.13.0 // indirect
	github.com/go-faster/city v1.0.1 // indirect
	github.com/go-faster/errors v0.7.1 // indirect
	github.com/go-sql-driver/mysql v1.7.1 // indirect
	github.com/google/uuid v1.5.0 // indirect
	github.com/hashicorp/go-version v1.6.0 // indirect
	github.com/jinzhu/inflection v1.0.0 // indirect
	github.com/jinzhu/now v1.1.5 // indirect
	github.com/klauspost/compress v1.17.4 // indirect
	github.com/paulmach/orb v0.10.0 // indirect
	github.com/pierrec/lz4/v4 v4.1.19 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/safchain/ethtool v0.3.0 // indirect
	github.com/segmentio/asm v1.2.0 // indirect
	github.com/shopspring/decimal v1.3.1 // indirect
	github.com/timandy/routine v1.1.2 // indirect
	go.opentelemetry.io/otel v1.21.0 // indirect
	go.opentelemetry.io/otel/trace v1.21.0 // indirect
	golang.org/x/sys v0.16.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
	gorm.io/driver/mysql v1.5.2 // indirect
)
