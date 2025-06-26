package clickhouse_go

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"reflect"

	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"
	"github.com/leisurelicht/norm/operator"
	ck "github.com/leisurelicht/norm/operator/clickhouse"
)

func Open(opt *clickhouse.Options) (driver.Conn, error) {
	return clickhouse.Open(opt)
}

func OpenDB(opt *clickhouse.Options) *sql.DB {
	return clickhouse.OpenDB(opt)
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

func (d *Operator) Insert(ctx context.Context, conn any, query string, args ...any) (id int64, err error) {
	if c, err := conn.(driver.DB); err == nil {
		return 0, nil
	} else if c, err := conn.(*sql.Db); err == nil {

	}
	return 0, nil
}

func (d *Operator) BulkInsert(ctx context.Context, conn any, query string, args []string, data []map[string]any) (num int64, err error) {
	return 0, nil
}

func (d *Operator) Remove(ctx context.Context, conn any, query string, args ...any) (num int64, err error) {
	return 0, nil
}

func (d *Operator) Update(ctx context.Context, conn any, query string, args ...any) (num int64, err error) {
	return 0, nil
}

func (d *Operator) Count(ctx context.Context, conn any, query string, args ...any) (num int64, err error) {
	return 0, nil
}

func (d *Operator) Exist(ctx context.Context, conn any, condition string, args ...any) (bool, error) {
	return false, nil
}

func (d *Operator) FindOne(ctx context.Context, conn any, model any, query string, args ...any) (err error) {
	err = conn.(driver.Conn).QueryRow(ctx, query, args...).ScanStruct(model)

	switch {
	case err == nil:
		return nil
	case errors.Is(err, sql.ErrNoRows):
		return operator.ErrNotFound
	default:
		return err
	}

}

func (d *Operator) FindAll(ctx context.Context, conn any, model any, query string, args ...any) (err error) {
	rows, err := conn.(driver.Conn).Query(ctx, query, args...)
	if err != nil {
		return err
	}

	defer func() { _ = rows.Close() }()

	rv := reflect.ValueOf(model)
	sliceValue := rv.Elem()
	elementType := sliceValue.Type().Elem()

	for rows.Next() {
		newElement := reflect.New(elementType).Interface()

		if err := rows.ScanStruct(newElement); err != nil {
			fmt.Printf("Scan Struct Failed. Error: %s", err)
			return err
		}
		sliceValue.Set(reflect.Append(sliceValue, reflect.ValueOf(newElement).Elem()))
	}

	if sliceValue.Len() == 0 {
		return operator.ErrNotFound
	}

	return nil
}
