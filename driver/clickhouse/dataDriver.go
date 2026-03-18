package data_clickhouse

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/farseer-go/data"
	"gorm.io/driver/clickhouse"
	"gorm.io/gorm"
)

type dataDriver struct {
}

func (receiver *dataDriver) GetDriver(connectionString string) gorm.Dialector {
	// clickhouse://user:123456@127.0.0.1:9942/dbname?dial_timeout=10s&read_timeout=60s
	// 注意：为了避免 code: 101 错误，建议在连接字符串中添加以下参数：
	if !strings.Contains(connectionString, "max_execution_time") {
		// - max_execution_time=60 (查询最大执行时间)
		connectionString += "&max_execution_time=60"
	}

	// if !strings.Contains(connectionString, "connection_max_life_time") {
	// 	// - connection_max_life_time=300 (连接最大生命周期，秒)
	// 	connectionString += "&connection_max_life_time=300"
	// }

	if !strings.Contains(connectionString, "&async_insert=") {
		connectionString += "&async_insert=1"
	}

	if !strings.Contains(connectionString, "&wait_for_async_insert=") {
		connectionString += "&wait_for_async_insert=1"
	}

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

func (receiver *dataDriver) CreateIndex(tableName string, idxName string, idxField data.IdxField) string {
	var b bytes.Buffer
	b.WriteString("CREATE ")
	if idxField.IsUNIQUE {
		b.WriteString("UNIQUE ")
	}
	b.WriteString(fmt.Sprintf("INDEX %s ON %s (%s);", idxName, tableName, idxField.Fields))
	return b.String()
}
