package test

import "github.com/zeromicro/go-zero/core/stores/sqlx"

var _ SourceModel = (*customSourceModel)(nil)

type (
	// SourceModel is an interface to be customized, add more methods here,
	// and implement the added methods in customSourceModel.
	SourceModel interface {
		sourceModel
	}

	customSourceModel struct {
		*defaultSourceModel
	}
)

// NewSourceModel returns a model for the database table.
func NewSourceModel(conn sqlx.SqlConn) SourceModel {
	return &customSourceModel{
		defaultSourceModel: newSourceModel(conn),
	}
}
