package data_clickhouse

import (
	"bytes"
	"fmt"
	"github.com/farseer-go/data"
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

func (receiver *dataDriver) CreateIndex(tableName string, idxField data.IdxField) string {
	var b bytes.Buffer
	b.WriteString("CREATE ")
	if idxField.IsUNIQUE {
		b.WriteString("UNIQUE ")
	}
	b.WriteString(fmt.Sprintf("INDEX %s ON %s (%s);", idxField.IdxName, tableName, idxField.Fields))
	return b.String()
}
