package data_clickhouse

import (
	"gorm.io/driver/clickhouse"
	"gorm.io/gorm"
)

type dataDriver struct {
}

func (receiver *dataDriver) GetDriver(connectionString string) gorm.Dialector {
	// clickhouse://user:123456@127.0.0.1:9942/dbname?dial_timeout=10s&read_timeout=20s
	//return clickhouse.Open(receiver.ConnectionString)
	return clickhouse.New(clickhouse.Config{
		DSN:                          connectionString,
		DisableDatetimePrecision:     true,     // disable datetime64 precision, not supported before clickhouse 20.4
		DontSupportRenameColumn:      true,     // rename column not supported before clickhouse 20.4
		DontSupportEmptyDefaultValue: false,    // do not consider empty strings as valid default values
		SkipInitializeWithVersion:    false,    // smart configure based on used version
		DefaultGranularity:           3,        // 1 granule = 8192 rows
		DefaultCompression:           "LZ4",    // default compression algorithm. LZ4 is lossless
		DefaultIndexType:             "minmax", // index stores extremes of the expression
		DefaultTableEngineOpts:       "ENGINE=MergeTree() ORDER BY tuple()",
	})
}

func (receiver *DataDriver) CreateIndex(idxName string, fieldsName ...string) string {
	return fmt.Sprintf("CREATE INDEX %s ON table_name (%s);", idxName, strings.Join(fieldsName, ","))
}
