package data

import (
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"gorm.io/hints"
)

type FinalHint struct {
}

func (indexHint FinalHint) Build(builder clause.Builder) {
	builder.WriteString(" FINAL ")
}
func (indexHint FinalHint) ModifyStatement(stmt *gorm.Statement) {
	for _, name := range []string{"FROM"} {
		clause := stmt.Clauses[name]

		if clause.AfterExpression == nil {
			clause.AfterExpression = indexHint
		} else {
			clause.AfterExpression = hints.Exprs{clause.AfterExpression, indexHint}
		}

		if name == "FROM" {
			clause.Builder = hints.IndexHintFromClauseBuilder
		}

		stmt.Clauses[name] = clause
	}
}
