package clickhouse

import (
	"context"

	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"
	ck "github.com/leisurelicht/norm/operator/clickhouse"
)

func Open(opt *clickhouse.Options) (driver.Conn, error) {
	return clickhouse.Open(opt)
}

const (
	placeholder = "?"
)

type Operator struct {
	tableName   string
	placeholder string
}

func NewOperator() *Operator {
	return &Operator{
		placeholder: placeholder,
	}
}

func (d *Operator) SetTableName(tableName string) {
	d.tableName = tableName
}

func (d *Operator) Placeholder() string {
	return d.placeholder
}

func (d *Operator) OperatorSQL(operator string) string {
	return ck.Operators[operator]
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
