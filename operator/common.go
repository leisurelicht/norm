package operator

import (
	"context"
	"errors"
)

var (
	ErrDuplicateKey = errors.New("duplicate key")
	ErrNotFound     = errors.New("not found")
	ErrResultIsNil  = errors.New("result is nil")
)

type Operator interface {
	OperatorSQL(operator, method string) string
	GetPlaceholder() string
	GetDBTag() string
	GetTableName() string
	SetTableName(tableName string) Operator
	Insert(ctx context.Context, query string, args ...any) (id int64, err error)
	BulkInsert(ctx context.Context, query string, args []string, data []map[string]any) (num int64, err error)
	Remove(ctx context.Context, query string, args ...any) (num int64, err error)
	Update(ctx context.Context, query string, args ...any) (num int64, err error)
	Count(ctx context.Context, condition string, args ...any) (num int64, err error)
	Exist(ctx context.Context, condition string, args ...any) (bool, error)
	FindOne(ctx context.Context, model any, query string, args ...any) (err error)
	FindAll(ctx context.Context, model any, query string, args ...any) (err error)
}

const (
	placeholder = "?"
)

type AddOptions struct {
	TableName   string
	Placeholder string
	DBTag       string
}

type AddFunc func(params *AddOptions)

func DefaultAddOptions(dbTag string) AddOptions {
	return AddOptions{
		Placeholder: placeholder,
		DBTag:       dbTag,
	}
}

func WithTableName(tableName string) AddFunc {
	return func(op *AddOptions) {
		op.TableName = tableName
	}
}

func WithPlaceholder(placeholder string) AddFunc {
	return func(op *AddOptions) {
		op.Placeholder = placeholder
	}
}

func WithDBTag(dbTag string) AddFunc {
	return func(op *AddOptions) {
		op.DBTag = dbTag
	}
}
