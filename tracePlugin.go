package data

import (
	"github.com/farseer-go/linkTrace"
	"gorm.io/gorm"
)

type TracePlugin struct{}

func (op *TracePlugin) Name() string {
	return "tracePlugin"
}

func (op *TracePlugin) Initialize(db *gorm.DB) (err error) {
	// 执行SQL前
	_ = db.Callback().Raw().Before("trace_before_raw").Register("trace_before", op.traceBefore)
	_ = db.Callback().Create().Before("trace_before_create").Register("trace_before", op.traceBefore)
	_ = db.Callback().Delete().Before("trace_before_delete").Register("trace_before", op.traceBefore)
	_ = db.Callback().Update().Before("trace_before_update").Register("trace_before", op.traceBefore)
	_ = db.Callback().Query().Before("trace_before_query").Register("trace_before", op.traceBefore)
	_ = db.Callback().Row().Before("trace_before_row").Register("trace_before", op.traceBefore)

	// 执行完SQL后
	_ = db.Callback().Raw().After("trace_after_raw").Register("trace_after", op.traceAfter)
	_ = db.Callback().Create().After("trace_after_create").Register("trace_after", op.traceAfter)
	_ = db.Callback().Delete().After("trace_after_delete").Register("trace_after", op.traceAfter)
	_ = db.Callback().Update().After("trace_after_update").Register("trace_after", op.traceAfter)
	_ = db.Callback().Query().After("trace_after_query").Register("trace_after", op.traceAfter)
	_ = db.Callback().Row().After("trace_after_row").Register("trace_after", op.traceAfter)
	return
}

// 链路追踪记录
func (op *TracePlugin) traceBefore(db *gorm.DB) {
	sql := db.Dialector.Explain(db.Statement.SQL.String(), db.Statement.Vars...)
	detail := linkTrace.TraceDatabase(db.Statement.DB.Name(), db.Statement.Table, sql)
	db.InstanceSet("trace", detail)
}

// 链路追踪记录
func (op *TracePlugin) traceAfter(db *gorm.DB) {
	if result, exists := db.InstanceGet("trace"); exists {
		if detail, isOk := result.(*linkTrace.TraceDatabaseDetail); isOk {
			//sqlInfo.Stack = utils.FileWithLineNum()
			//sqlInfo.Rows = db.Statement.RowsAffected
			detail.End(db.Error)
		}
	}
}
