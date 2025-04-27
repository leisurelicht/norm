package mysql

import (
	"context"
	"database/sql"
	"errors"
	"strings"

	"github.com/zeromicro/go-zero/core/logc"
	"github.com/zeromicro/go-zero/core/stores/sqlx"

	"github.com/leisurelicht/norm/operator"
)

var operators = map[string]string{
	"exact":   "`%s` = ?",
	"exclude": "`%s` != ?",
	"iexact":  "`%s` LIKE ?",
	"gt":      "`%s` > ?",
	"gte":     "`%s` >= ?",
	"lt":      "`%s` < ?",
	"lte":     "`%s` <= ?",
	"len":     "LENGTH(`%s`) = ?",

	"in":          "`%s`%s IN",
	"between":     "`%s`%s BETWEEN ? AND ?",
	"contains":    "`%s`%s LIKE BINARY ?",
	"icontains":   "`%s`%s LIKE ?",
	"startswith":  "`%s`%s LIKE BINARY ?",
	"istartswith": "`%s`%s LIKE ?",
	"endswith":    "`%s`%s LIKE BINARY ?",
	"iendswith":   "`%s`%s LIKE ?",

	"unimplemented": "UNIMPLEMENTED", // Placeholder for unimplemented operators

	"is_null": "`%s` IS NULL",
}

type Operator struct {
	tableName string
}

func NewOperator() *Operator {
	return &Operator{}
}

func (d *Operator) SetTableName(tableName string) {
	d.tableName = tableName
}

func (d *Operator) OperatorSQL(operator string) string {
	return operators[operator]
}

func (d *Operator) Insert(ctx context.Context, conn any, query string, args ...any) (id int64, err error) {
	res, err := conn.(sqlx.SqlConn).ExecCtx(ctx, query, args...)
	if err != nil {
		if strings.Contains(err.Error(), "1062") {
			return 0, operator.ErrDuplicateKey
		}
		logc.Errorf(ctx, "Insert Error: %s", err)
		return 0, err
	}

	id, err = res.LastInsertId()
	if err != nil {
		logc.Errorf(ctx, "Get last insert id error: %s", err)
	}

	return id, err
}

func (d *Operator) BulkInsert(ctx context.Context, conn any, query string, args []string, data []map[string]any) (num int64, err error) {
	blk, err := sqlx.NewBulkInserter(conn.(sqlx.SqlConn), query)
	if err != nil {
		panic(err)
	}

	values := make([]any, len(args))
	for _, row := range data {
		for i, arg := range args {
			if val, ok := row[arg]; ok {
				values[i] = val
			}
		}
		blk.Insert(values...)
	}

	blk.SetResultHandler(func(result sql.Result, err error) {
		if err != nil {
			logc.Errorf(ctx, "Bulk insert error: %s", err)
			return
		}

		num, err = result.RowsAffected()
		if err != nil {
			logc.Errorf(ctx, "Bulk insert rows affected error: %s", err)
			return
		}

		logc.Infof(ctx, "Inserted %d rows", num)
	})

	blk.Flush()

	return num, err
}

func (d *Operator) Remove(ctx context.Context, conn any, query string, args ...any) (num int64, err error) {
	res, err := conn.(sqlx.SqlConn).ExecCtx(ctx, query, args...)
	if err != nil {
		logc.Errorf(ctx, "Remove error: %+v", err)
		return 0, err
	}

	num, err = res.RowsAffected()
	if err != nil {
		logc.Errorf(ctx, "Remove rows affected error: %+v", err)
		return 0, err
	}
	return num, nil
}

func (d *Operator) Update(ctx context.Context, conn any, query string, args ...any) (num int64, err error) {
	res, err := conn.(sqlx.SqlConn).Exec(query, args...)
	if err != nil {
		logc.Errorf(ctx, "Update error: %+v", err)
		return 0, err
	}

	num, err = res.RowsAffected()
	if err != nil {
		logc.Errorf(ctx, "Update rows affected error: %+v", err)
		return 0, err
	}
	return num, nil
}

func (d *Operator) Count(ctx context.Context, conn any, query string, args ...any) (num int64, err error) {
	err = conn.(sqlx.SqlConn).QueryRowCtx(ctx, &num, query, args...)

	switch {
	case err == nil:
		return num, nil
	case errors.Is(err, sqlx.ErrNotFound):
		return 0, nil
	default:
		logc.Errorf(ctx, "Count error: %+v. ", err)
		return 0, err
	}
}

func (d *Operator) Exist(ctx context.Context, conn any, condition string, args ...any) (bool, error) {
	query := "SELECT count(1) FROM " + d.tableName + condition

	var num int64
	err := conn.(sqlx.SqlConn).QueryRowCtx(ctx, &num, query, args...)

	switch {
	case err == nil:
		return num > 0, nil
	case errors.Is(err, sqlx.ErrNotFound):
		return false, nil
	default:
		logc.Errorf(ctx, "Exist error: %+v", err)
		return false, err
	}
}

func (d *Operator) FindOne(ctx context.Context, conn any, model any, query string, args ...any) (err error) {
	err = conn.(sqlx.SqlConn).QueryRowPartialCtx(ctx, model, query, args...)

	switch {
	case err == nil:
		return nil
	case errors.Is(err, sqlx.ErrNotFound):
		return operator.ErrNotFound
	default:
		logc.Errorf(ctx, "FindOne error: %+v", err)
		return err
	}
}

func (d *Operator) FindAll(ctx context.Context, conn any, model any, query string, args ...any) (err error) {
	err = conn.(sqlx.SqlConn).QueryRowsPartialCtx(ctx, model, query, args...)

	switch {
	case err == nil:
		return nil
	case errors.Is(err, sqlx.ErrNotFound):
		return operator.ErrNotFound
	default:
		logc.Errorf(ctx, "FindAll error: %+v", err)
		return err
	}
}
