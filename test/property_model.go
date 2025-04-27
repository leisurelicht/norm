package test

import "github.com/zeromicro/go-zero/core/stores/sqlx"

var _ PropertyModel = (*customPropertyModel)(nil)

type (
	// PropertyModel is an interface to be customized, add more methods here,
	// and implement the added methods in customPropertyModel.
	PropertyModel interface {
		propertyModel
	}

	customPropertyModel struct {
		*defaultPropertyModel
	}
)

// NewPropertyModel returns a model for the database table.
func NewPropertyModel(conn sqlx.SqlConn) PropertyModel {
	return &customPropertyModel{
		defaultPropertyModel: newPropertyModel(conn),
	}
}
