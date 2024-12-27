package norm

import (
	"context"
	"fmt"
	"github.com/leisurelicht/norm/operator"
	"reflect"
	"strings"
)

var (
	ErrDuplicateKey = operator.ErrDuplicateKey
	ErrNotFound     = operator.ErrDuplicateKey
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

const (
	OrderKey = "~order~"
)

type (
	Cond map[string]any
	AND  map[string]any
	OR   map[string]any
)

func ToOR(key string) string {
	if key != "" && key != OrderKey && !strings.HasPrefix(key, orPrefix) {
		return orPrefix + strings.TrimSpace(key)
	}
	return key
}

func EachOR(conds any) any {
	switch conds.(type) {
	case Cond:
		conditions := conds.(Cond)
		for k, v := range conditions {
			if k == OrderKey {
				continue
			}
			if !strings.HasPrefix(k, orPrefix) {
				delete(conditions, k)
				conditions[orPrefix+k] = v
			}
		}
		return conditions
	case AND:
		conditions := conds.(AND)
		for k, v := range conditions {
			if k == OrderKey {
				continue
			}
			if !strings.HasPrefix(k, orPrefix) {
				delete(conditions, k)
				conditions[orPrefix+k] = v
			}
		}
		return conditions
	case OR:
		conditions := conds.(OR)
		for k, v := range conditions {
			if k == OrderKey {
				continue
			}
			if !strings.HasPrefix(k, orPrefix) {
				delete(conditions, k)
				conditions[orPrefix+k] = v
			}
		}
		return conditions
	default:
		fmt.Println(reflect.TypeOf(conds).String())
	}
	return conds
}
