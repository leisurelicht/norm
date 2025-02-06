package clickhouse

import (
	"context"
	"strings"
)

var operators = map[string]string{
	"exact":   "`%s` = ?",
	"exclude": "`%s` != ?",
	"iexact":  "`%s` LIKE ?",
	"gt":      "`%s` > ?",
	"gte":     "`%s` >= ?",
	"lt":      "`%s` < ?",
	"lte":     "`%s` <= ?",

	"in":              "`%s`%s IN",
	"not_in":          "`%s`%s IN",
	"contains":        "`%s`%s LIKE ?",
	"not_contains":    "`%s`%s LIKE ?",
	"icontains":       "`%s`%s ILIKE ?",
	"not_icontains":   "`%s`%s ILIKE ?",
	"startswith":      "`%s`%s LIKE ?",
	"not_startswith":  "`%s`%s LIKE ?",
	"istartswith":     "`%s`%s ILIKE ?",
	"not_istartswith": "`%s`%s ILIKE ?",
	"endswith":        "`%s`%s LIKE ?",
	"not_endswith":    "`%s`%s LIKE ?",
	"iendswith":       "`%s`%s ILIKE ?",
	"not_iendswith":   "`%s`%s ILIKE ?",
}

var selectKeys = map[string]struct{}{
	"DISTINCT": {},
	"AS":       {},
}

type Operator struct{}

func NewOperator() *Operator {
	return &Operator{}
}

func (d *Operator) OperatorSQL(operator string) string {
	return operators[operator]
}

func (d *Operator) IsSelectKey(word string) bool {
	_, exists := selectKeys[strings.ToUpper(word)] // 判断关键字（区分大小写）
	return exists
}

func (d *Operator) Insert(ctx context.Context, conn any, sql string, args ...any) (id int64, err error) {
	return id, err
}

func (d *Operator) BulkInsert(ctx context.Context, conn any, sql string, args ...any) (err error) {
	return err
}

func (d *Operator) Remove(ctx context.Context, conn any, sql string, args ...any) (num int64, err error) {
	return num, err
}

func (d *Operator) Update(ctx context.Context, conn any, sql string, args ...any) (num int64, err error) {
	return num, err
}

func (d *Operator) Count(ctx context.Context, conn any, sql string, args ...any) (num int64, err error) {
	return num, err
}

func (d *Operator) FindOne(ctx context.Context, conn any, model any, sql string, args ...any) (err error) {
	return err
}

func (d *Operator) FindAll(ctx context.Context, conn any, model any, sql string, args ...any) (err error) {
	return err
}
