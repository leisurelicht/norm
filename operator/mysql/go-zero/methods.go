package go_zero

import (
	"context"
	"database/sql"
	"errors"
	"github.com/leisurelicht/norm/internal/operator"
	"github.com/leisurelicht/norm/internal/operator/mysql"
	"strings"

	"github.com/leisurelicht/norm/internal/config"
	"github.com/zeromicro/go-zero/core/logc"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

// NewMysql returns a mysql connection.
func NewMysql(datasource string, opts ...sqlx.SqlOption) sqlx.SqlConn {
	switch config.Get().Level {
	case config.Debug:
		logx.SetLevel(logx.DebugLevel)
	case config.Info:
		logx.SetLevel(logx.InfoLevel)
		logx.DisableStat()
	case config.Warn:
		logx.SetLevel(logx.SevereLevel)
		logx.DisableStat()
	case config.Error:
		logx.SetLevel(logx.ErrorLevel)
		logx.DisableStat()
	default:
		logx.SetLevel(logx.InfoLevel)
		logx.DisableStat()
	}
	return sqlx.NewMysql(datasource, opts...)
}

const dbTag = "db"

type OperatorImpl struct {
	conn sqlx.SqlConn
	operator.AddOptions
}

func NewOperator(conn sqlx.SqlConn, opts ...operator.AddFunc) OperatorImpl {
	addOptions := operator.DefaultAddOptions(dbTag)
	for _, opt := range opts {
		opt(&addOptions)
	}
	return OperatorImpl{
		conn:       conn,
		AddOptions: addOptions,
	}
}

func WithTableName(tableName string) operator.AddFunc {
	return operator.WithTableName(tableName)
}

func (d OperatorImpl) SetTableName(tableName string) operator.Operator {
	if d.TableName == "" {
		d.TableName = tableName
	}
	return d
}

func (d OperatorImpl) GetTableName() string {
	return d.TableName
}

func (d OperatorImpl) GetPlaceholder() string {
	return d.Placeholder
}

func (d OperatorImpl) GetDBTag() string {
	return d.DBTag
}

func (d OperatorImpl) OperatorSQL(operator, method string) string {
	op, ok := mysql.Operators[operator]
	if !ok {
		return ""
	}
	if method == "" {
		return op
	}
	if methodSQL, ok := mysql.Methods[method]; ok {
		op = strings.ReplaceAll(op, "?", methodSQL)
	}
	return op
}

func (d OperatorImpl) Insert(ctx context.Context, query string, args ...any) (id int64, err error) {
	res, err := d.conn.ExecCtx(ctx, query, args...)
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

func (d OperatorImpl) BulkInsert(ctx context.Context, query string, args []string, data []map[string]any) (num int64, err error) {
	blk, err := sqlx.NewBulkInserter(d.conn, query)
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

func (d OperatorImpl) Remove(ctx context.Context, query string, args ...any) (num int64, err error) {
	res, err := d.conn.ExecCtx(ctx, query, args...)
	if err != nil {
		logc.Errorf(ctx, "Remove error: %s", err)
		return 0, err
	}

	num, err = res.RowsAffected()
	if err != nil {
		logc.Errorf(ctx, "Remove rows affected error: %s", err)
		return 0, err
	}
	return num, nil
}

func (d OperatorImpl) Update(ctx context.Context, query string, args ...any) (num int64, err error) {
	res, err := d.conn.Exec(query, args...)
	if err != nil {
		logc.Errorf(ctx, "Update error: %s", err)
		return 0, err
	}

	num, err = res.RowsAffected()
	if err != nil {
		logc.Errorf(ctx, "Update rows affected error: %s", err)
		return 0, err
	}
	return num, nil
}

func (d OperatorImpl) Count(ctx context.Context, condition string, args ...any) (num int64, err error) {
	query := "SELECT count(1) FROM " + d.TableName + condition

	err = d.conn.QueryRowCtx(ctx, &num, query, args...)

	switch {
	case err == nil:
		return num, nil
	case errors.Is(err, sqlx.ErrNotFound):
		return 0, nil
	default:
		logc.Errorf(ctx, "Count error: %s. ", err)
		return 0, err
	}
}

func (d OperatorImpl) Exist(ctx context.Context, condition string, args ...any) (exist bool, err error) {
	query := "SELECT count(1) FROM " + d.TableName + condition

	var num int64
	err = d.conn.QueryRowCtx(ctx, &num, query, args...)

	switch {
	case err == nil:
		return num > 0, nil
	case errors.Is(err, sqlx.ErrNotFound):
		return false, nil
	default:
		logc.Errorf(ctx, "Exist error: %s", err)
		return false, err
	}
}

func (d OperatorImpl) FindOne(ctx context.Context, model any, query string, args ...any) (err error) {
	err = d.conn.QueryRowPartialCtx(ctx, model, query, args...)

	switch {
	case err == nil:
		return nil
	case errors.Is(err, sqlx.ErrNotFound):
		return operator.ErrNotFound
	default:
		logc.Errorf(ctx, "FindOne error: %s", err)
		return err
	}
}

func (d OperatorImpl) FindAll(ctx context.Context, model any, query string, args ...any) (err error) {
	err = d.conn.QueryRowsPartialCtx(ctx, model, query, args...)

	switch {
	case err == nil:
		return nil
	case errors.Is(err, sqlx.ErrNotFound):
		return operator.ErrNotFound
	default:
		logc.Errorf(ctx, "FindAll error: %s", err)
		return err
	}
}
