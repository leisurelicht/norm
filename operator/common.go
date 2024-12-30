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
	OperatorSQL(operator string) string
	Insert(ctx context.Context, conn any, sql string, args ...any) (id int64, err error)
	BulkInsert(ctx context.Context, conn any, sql string, args ...any) (err error)
	Remove(ctx context.Context, conn any, sql string, args ...any) (num int64, err error)
	Update(ctx context.Context, conn any, sql string, args ...any) (num int64, err error)
	Count(ctx context.Context, conn any, sql string, args ...any) (num int64, err error)
	FindOne(ctx context.Context, conn any, model any, sql string, args ...any) (err error)
	FindAll(ctx context.Context, conn any, model any, sql string, args ...any) (err error)
}
