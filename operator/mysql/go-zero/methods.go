package go_zero

import (
	"context"
	"errors"
	"fmt"
	"strings"

	mysqlDriver "github.com/go-sql-driver/mysql"
	"github.com/zeromicro/go-zero/core/logc"
	"github.com/zeromicro/go-zero/core/stores/sqlx"

	"github.com/leisurelicht/norm/internal/operator"
	mysqlOp "github.com/leisurelicht/norm/internal/operator/mysql"
)

// NewMysql returns a mysql connection.
func NewMysql(datasource string, opts ...sqlx.SqlOption) sqlx.SqlConn {
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

func (d OperatorImpl) WithSession(session any) operator.Operator {
	if sqlSession, ok := session.(sqlx.Session); ok {
		d.conn = sqlx.NewSqlConnFromSession(sqlSession)
	}
	return d
}

func (d OperatorImpl) OperatorSQL(operator, method string) string {
	op, ok := mysqlOp.Operators[operator]
	if !ok {
		return ""
	}
	if method == "" {
		return op
	}
	if methodSQL, ok := mysqlOp.Methods[method]; ok {
		op = strings.ReplaceAll(op, "?", methodSQL)
	}
	return op
}

func (d OperatorImpl) Insert(ctx context.Context, query string, args ...any) (id int64, err error) {
	res, err := d.conn.ExecCtx(ctx, query, args...)
	if err != nil {
		if isDuplicateKeyError(err) {
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
	query, err = buildBulkInsertQuery(query, len(data))
	if err != nil {
		logc.Errorf(ctx, "Build bulk insert query error: %s", err)
		return 0, err
	}

	values := make([]any, 0, len(args)*len(data))
	for _, row := range data {
		for _, arg := range args {
			if val, ok := row[arg]; ok {
				values = append(values, val)
				continue
			}
			values = append(values, nil)
		}
	}

	result, err := d.conn.ExecCtx(ctx, query, values...)
	if err != nil {
		if isDuplicateKeyError(err) {
			return 0, operator.ErrDuplicateKey
		}
		logc.Errorf(ctx, "Bulk insert error: %s", err)
		return 0, err
	}

	num, err = result.RowsAffected()
	if err != nil {
		logc.Errorf(ctx, "Bulk insert rows affected error: %s", err)
		return 0, err
	}

	logc.Infof(ctx, "Inserted %d rows", num)
	return num, err
}

func buildBulkInsertQuery(query string, rows int) (string, error) {
	if rows <= 1 {
		return query, nil
	}

	lower := strings.ToLower(query)

	// Limit VALUES keyword search to the main INSERT ... VALUES clause.
	// If the query contains an "ON DUPLICATE KEY" section, ignore everything after it
	// so we don't accidentally match the MySQL VALUES() function there.
	searchEnd := strings.Index(lower, " on duplicate key")
	searchLower := lower
	if searchEnd > 0 {
		searchLower = lower[:searchEnd]
	}

	// Use the last occurrence so identifiers like `values_log` or `default_values`
	// in table/column names don't confuse us.
	pos := strings.LastIndex(searchLower, "values")
	if pos < 0 {
		return "", fmt.Errorf("invalid insert query, missing VALUES: %q", query)
	}

	head := strings.TrimSpace(query[:pos+len("values")])
	tail := strings.TrimSpace(query[pos+len("values"):])
	left := strings.IndexByte(tail, '(')
	right := strings.IndexByte(tail, ')')
	if left < 0 || right <= left {
		return "", fmt.Errorf("invalid insert query, bad values placeholder: %q", query)
	}

	rowTemplate := tail[left : right+1]
	suffix := strings.TrimSpace(tail[right+1:])

	valueTemplates := make([]string, rows)
	for i := range valueTemplates {
		valueTemplates[i] = rowTemplate
	}

	bulkQuery := head + " " + strings.Join(valueTemplates, ",")
	if suffix != "" {
		bulkQuery += " " + suffix
	}
	return bulkQuery, nil
}

func isDuplicateKeyError(err error) bool {
	var mysqlErr *mysqlDriver.MySQLError
	return errors.As(err, &mysqlErr) && mysqlErr.Number == 1062
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

	if err != nil {
		logc.Errorf(ctx, "FindAll error: %s", err)
		return err
	}

	return nil
}
