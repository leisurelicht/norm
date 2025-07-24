package operator

import (
	"context"
	"errors"
)

var (
	ErrDuplicateKey = errors.New("duplicate key")
	ErrNotFound     = errors.New("not found")
)

type Operator interface {
	OperatorSQL(operator, method string) string
	SetTableName(tableName string)
	Placeholder() string
	DBTag() string
	// Exec(ctx context.Context, conn any, sql string) (err error)
	Insert(ctx context.Context, conn any, query string, args ...any) (id int64, err error)
	BulkInsert(ctx context.Context, conn any, query string, args []string, data []map[string]any) (num int64, err error)
	Remove(ctx context.Context, conn any, query string, args ...any) (num int64, err error)
	Update(ctx context.Context, conn any, query string, args ...any) (num int64, err error)
	Count(ctx context.Context, conn any, condition string, args ...any) (num int64, err error)
	Exist(ctx context.Context, conn any, condition string, args ...any) (bool, error)
	FindOne(ctx context.Context, conn any, model any, query string, args ...any) (err error)
	FindAll(ctx context.Context, conn any, model any, query string, args ...any) (err error)
	// List(ctx context.Context, conn any, model any, query string, args ...any) (num int64, err error)
}
