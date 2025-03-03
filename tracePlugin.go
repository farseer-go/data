package data

import (
	"github.com/farseer-go/fs/parse"
	"github.com/farseer-go/fs/trace"
	"gorm.io/gorm"
)

type TracePlugin struct {
	traceManager trace.IManager
}

func (op *TracePlugin) Name() string {
	return "tracePlugin"
}

func (op *TracePlugin) Initialize(db *gorm.DB) (err error) {
	// 执行SQL前
	_ = db.Callback().Raw().Before("gorm:raw").Register("trace_before", op.traceBefore)
	_ = db.Callback().Create().Before("gorm:before_create").Register("trace_before", op.traceBefore)
	_ = db.Callback().Delete().Before("gorm:before_delete").Register("trace_before", op.traceBefore)
	_ = db.Callback().Update().Before("gorm:setup_reflect_value").Register("trace_before", op.traceBefore)
	_ = db.Callback().Query().Before("gorm:query").Register("trace_before", op.traceBefore)
	_ = db.Callback().Row().Before("gorm:row").Register("trace_before", op.traceBefore)

	// 执行完SQL后
	_ = db.Callback().Raw().After("gorm:raw").Register("trace_after", op.traceAfter)
	_ = db.Callback().Create().After("gorm:after_create").Register("trace_after", op.traceAfter)
	_ = db.Callback().Delete().After("gorm:after_delete").Register("trace_after", op.traceAfter)
	_ = db.Callback().Update().After("gorm:after_update").Register("trace_after", op.traceAfter)
	_ = db.Callback().Query().After("gorm:after_query").Register("trace_after", op.traceAfter)
	_ = db.Callback().Row().After("gorm:row").Register("trace_after", op.traceAfter)
	return
}

// 链路追踪记录
func (op *TracePlugin) traceBefore(db *gorm.DB) {
	detail := op.traceManager.TraceDatabase()
	db.InstanceSet("trace", detail)
}

// 链路追踪记录
func (op *TracePlugin) traceAfter(db *gorm.DB) {
	if result, exists := db.InstanceGet("trace"); exists {
		if detail, isOk := result.(*trace.TraceDetail); isOk {
			if db.DryRun {
				detail.Ignore()
			} else {
				connectionString, _ := db.InstanceGet("ConnectionString")
				dbName, _ := db.InstanceGet("DbName")
				detail.TraceDetailDatabase.SetSql(parse.ToString(connectionString), parse.ToString(dbName), db.Statement.Table, db.Dialector.Explain(db.Statement.SQL.String(), db.Statement.Vars...), db.Statement.RowsAffected)
			}
			detail.End(db.Error)
		}
	}
}
