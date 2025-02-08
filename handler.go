package norm

import (
	"context"
	"errors"
	"fmt"
	"log"
	"reflect"
	"strings"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/builder"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

const (
	DefaultModelTag = "db"
	SelectTemp      = "SELECT %s FROM %s"
	InsertTemp      = "INSERT INTO %s (%s) VALUES (%s)"
	UpdateTemp      = "UPDATE %s SET %s"
	DeleteTemp      = "DELETE FROM %s"
)

const (
	UnsupportedControllerError = "%s not supported for %s"
)

type controllerCall struct {
	Name string
	Flag callFlag
}

var (
	ctlFilter  = controllerCall{Name: "Filter", Flag: callFilter}
	ctlExclude = controllerCall{Name: "Exclude", Flag: callExclude}
	ctlWhere   = controllerCall{Name: "Where", Flag: callWhere}
	ctlSelect  = controllerCall{Name: "Select", Flag: callSelect}
	ctlLimit   = controllerCall{Name: "Limit", Flag: callLimit}
	ctlOrderBy = controllerCall{Name: "OrderBy", Flag: callOrderBy}
	ctlGroupBy = controllerCall{Name: "GroupBy", Flag: callGroupBy}
)

var _ Controller = (*Impl)(nil)

type (
	Controller interface {
		ctx() context.Context
		hasCalled(f controllerCall) bool
		checkCalled(f ...controllerCall) ([]string, bool)
		WithSession(session sqlx.Session) Controller
		Reset() Controller
		Filter(filter ...any) Controller
		Exclude(exclude ...any) Controller
		Where(cond string, args ...any) Controller
		Select(columns any) Controller
		Limit(pageSize, pageNum int64) Controller
		OrderBy(orderBy any) Controller
		GroupBy(groupBy any) Controller
		Insert(data map[string]any) (id int64, err error)
		InsertModel(model any) (id int64, err error)
		BulkInsert(data []map[string]any, handler sqlx.ResultHandler) (err error)
		BulkInsertModel(modelSlice any, handler sqlx.ResultHandler) (err error)
		Remove() (num int64, err error)
		Update(data map[string]any) (num int64, err error)
		Count() (num int64, err error)
		FindOne() (result map[string]any, err error)
		FindOneModel(modelPtr any) (err error)
		FindAll() (result []map[string]any, err error)
		FindAllModel(modelSlicePtr any) (err error)
		Delete() (num int64, err error)
		Modify(data map[string]any) (num int64, err error)
		Exist() (exist bool, error error)
		List() (num int64, data []map[string]any, err error)
		GetOrCreate(data map[string]any) (result map[string]any, err error)
		CreateOrUpdate(data map[string]any, filter ...any) (created bool, num int64, err error)
		GetC2CMap(column1, column2 string) (res map[any]any, err error)
		CreateIfNotExist(data map[string]any) (id int64, created bool, err error)
	}

	Impl struct {
		context      context.Context
		conn         any
		model        any
		modelSlice   any
		operator     Operator
		tableName    string
		fieldNameMap map[string]struct{}
		fieldRows    string
		mTag         string
		qs           QuerySet
		called       callFlag
	}
)

// shiftName shift name like DevicePolicyMap to device_policy_map
func shiftName(s string) string {
	res := ""
	for i, c := range s {
		if c >= 'A' && c <= 'Z' {
			if i != 0 {
				res += "_"
			}
			res += string(c + 32)
		} else {
			res += string(c)
		}
	}
	return "`" + res + "`"
}

func NewController(conn any, op Operator, m any) func(ctx context.Context) Controller {
	t := reflect.TypeOf(m)
	if t.Kind() != reflect.Struct {
		log.Panicf("model [%s] must be a struct", t.Name())
		return nil
	}

	mPtr, mSlicePtr := CreatePointerAndSlice(m)

	return func(ctx context.Context) Controller {
		if ctx == nil {
			ctx = context.Background()
		}
		return &Impl{
			context:      ctx,
			conn:         conn,
			model:        mPtr,
			modelSlice:   mSlicePtr,
			operator:     op,
			tableName:    shiftName(t.Name()),
			fieldNameMap: StrSlice2Map(builder.RawFieldNames(m, true)),
			fieldRows:    strings.Join(builder.RawFieldNames(m), ","),
			mTag:         DefaultModelTag,
			qs:           NewQuerySet(op),
			called:       0,
		}
	}
}

func (m *Impl) ctx() context.Context {
	return m.context
}

func (m *Impl) setCalled(f controllerCall) {
	m.called = m.called | f.Flag
}

func (m *Impl) hasCalled(f controllerCall) bool {
	return m.called&f.Flag == f.Flag
}

func (m *Impl) checkCalled(f ...controllerCall) ([]string, bool) {
	var calledMethod []string
	for _, v := range f {
		if m.called&v.Flag == v.Flag {
			calledMethod = append(calledMethod, v.Name)
		}
	}
	return calledMethod, len(calledMethod) != 0
}

func (m *Impl) validateColumns(columns []string) (validatedColumns []string, err error) {
	unknownColumns := []string{}
	for _, v := range columns {
		if _, ok := m.fieldNameMap[v]; !ok {
			unknownColumns = append(unknownColumns, v)
			continue
		}
		validatedColumns = append(validatedColumns, v)
	}

	if len(unknownColumns) > 0 {
		return nil, errors.New("[" + strings.Join(unknownColumns, "; ") + "] not exist")
	}

	return
}

func (m *Impl) validateColumns2Str(columns []string) string {
	columnsRows := ""

	for _, v := range columns {
		if _, ok := m.fieldNameMap[v]; !ok {
			continue
		}
		columnsRows += "`" + v + "`,"
	}
	columnsRows = columnsRows[:len(columnsRows)-1]

	return columnsRows
}

func (m *Impl) checkQuerySetError() {
	if m.qs.Error() != nil {
		logx.WithCallerSkip(3).WithContext(m.ctx()).Errorf("sql conditions queryset error: %s", m.qs.Error())
	}
}

func (m *Impl) setError(format string, a ...any) {
	m.qs.setError(format, a...)
}

func (m *Impl) HaveError() error {
	if err := m.qs.Error(); err != nil {
		return err
	}
	return nil
}

func (m *Impl) WithSession(session sqlx.Session) Controller {
	m.conn = sqlx.NewSqlConnFromSession(session)
	return m
}

func (m *Impl) Reset() Controller {
	m.qs.Reset()
	return m
}

func (m *Impl) Filter(filter ...any) Controller {
	m.setCalled(ctlFilter)

	m.qs.FilterToSQL(0, filter...)

	m.checkQuerySetError()

	return m
}

func (m *Impl) Exclude(exclude ...any) Controller {
	m.setCalled(ctlExclude)

	m.qs.FilterToSQL(1, exclude...)

	m.checkQuerySetError()

	return m
}

func (m *Impl) Where(cond string, args ...any) Controller {
	m.setCalled(ctlWhere)

	m.qs.WhereToSQL(cond, args)
	return m
}

func (m *Impl) Select(selects any) Controller {
	m.setCalled(ctlSelect)

	switch selects.(type) {
	case string:
		if selects.(string) == "" {
			return m
		}
		m.qs.StrSelectToSQL(selects.(string))
	case []string:
		selectList, _ := selects.([]string)

		if len(selectList) == 0 {
			return m
		}

		validatedColumns, err := m.validateColumns(selectList)

		if err != nil {
			m.setError("Select columns validate error: %s", err)
			return m
		}

		m.qs.SliceSelectToSQL(validatedColumns)
	default:
		m.setError("Select type should be string or string slice")
	}

	return m
}

func (m *Impl) Limit(pageSize, pageNum int64) Controller {
	m.setCalled(ctlLimit)

	m.qs.LimitToSQL(pageSize, pageNum)
	return m
}

func (m *Impl) OrderBy(orderBy any) Controller {
	m.setCalled(ctlOrderBy)

	var validatedOrderBy []string

	switch orderBy.(type) {
	case string:
		if orderBy.(string) == "" {
			return m
		}
		m.qs.StrOrderByToSQL(orderBy.(string))
	case []string:
		orderByList, _ := orderBy.([]string)

		if len(orderByList) == 0 {
			return m
		}

		unknownColumns := []string{}
		for _, by := range orderByList {
			needValidate := by
			if strings.HasPrefix(by, "-") {
				needValidate = by[1:]
			}
			if _, ok := m.fieldNameMap[needValidate]; ok {
				validatedOrderBy = append(validatedOrderBy, by)
			} else {
				unknownColumns = append(unknownColumns, by)
				continue
			}
		}

		if len(unknownColumns) > 0 {
			m.setError("OrderBy columns validate error: [%s] not exist", strings.Join(unknownColumns, "; "))
			return m
		}

		m.qs.OrderByToSQL(validatedOrderBy)
	default:
		m.setError("OrderBy type should be string or string slice")
	}

	return m
}

func (m *Impl) GroupBy(groupBy any) Controller {
	m.setCalled(ctlGroupBy)

	switch groupBy.(type) {
	case string:
		if groupBy.(string) == "" {
			return m
		}
		m.qs.StrGroupByToSQL(groupBy.(string))
	case []string:
		groupByList, _ := groupBy.([]string)

		if len(groupByList) == 0 {
			return m
		}

		validatedColumns, err := m.validateColumns(groupByList)

		if err != nil {
			m.setError("GroupBy columns validate error: %s", err)
			return m
		}

		m.qs.SliceGroupByToSQL(validatedColumns)
	default:
		m.setError("GroupBy type should be string or string slice")
		return m
	}

	return m
}

func (m *Impl) Insert(data map[string]any) (id int64, err error) {
	if methods, called := m.checkCalled(ctlFilter, ctlExclude, ctlWhere, ctlOrderBy, ctlGroupBy, ctlSelect); called {
		return 0, fmt.Errorf(UnsupportedControllerError, methods, "Insert")
	}

	if err = m.HaveError(); err != nil {
		return 0, err
	}

	var (
		rows []string
		args []any
	)

	for k := range m.fieldNameMap {
		if _, ok := data[k]; !ok {
			continue
		}
		rows = append(rows, fmt.Sprintf("`%s`", k))
		args = append(args, data[k])
	}

	sql := fmt.Sprintf(InsertTemp, m.tableName, strings.Join(rows, ","), strings.Repeat("?,", len(rows)-1)+"?")

	return m.operator.Insert(m.ctx(), m.conn, sql, args...)
}

func (m *Impl) InsertModel(model any) (id int64, err error) {
	return m.Insert(Struct2Map(model, m.mTag))
}

func (m *Impl) BulkInsert(data []map[string]any, handler sqlx.ResultHandler) (err error) {
	return m.operator.BulkInsert(m.ctx(), m.conn, m.tableName, data, handler)
}

func (m *Impl) BulkInsertModel(modelSlice any, handler sqlx.ResultHandler) (err error) {
	return m.BulkInsert(StructSlice2MapSlice(modelSlice, m.mTag), handler)
}

func (m *Impl) Remove() (num int64, err error) {
	if methods, called := m.checkCalled(ctlGroupBy, ctlSelect, ctlOrderBy); called {
		return 0, fmt.Errorf(UnsupportedControllerError, methods, "Remove")
	}

	if err = m.HaveError(); err != nil {
		return num, err
	}

	sql := fmt.Sprintf(DeleteTemp, m.tableName)

	filterSQL, filterArgs := m.qs.GetQuerySet()
	sql += filterSQL

	return m.operator.Remove(m.ctx(), m.conn, sql, filterArgs...)
}

func (m *Impl) Update(data map[string]any) (num int64, err error) {
	if methods, called := m.checkCalled(ctlGroupBy, ctlSelect, ctlOrderBy); called {
		return 0, fmt.Errorf(UnsupportedControllerError, methods, "Update")
	}

	if err = m.HaveError(); err != nil {
		return num, err
	}

	var (
		args       []any
		updateRows []string
		updateArgs []any
	)

	for k, v := range data {
		if _, ok := m.fieldNameMap[k]; !ok {
			return 0, errors.New("update column [" + k + "] not exist")
		}
		updateRows = append(updateRows, "`"+k+"`")
		updateArgs = append(updateArgs, v)
	}

	sql := fmt.Sprintf(UpdateTemp, m.tableName, strings.Join(updateRows, "=?,")+"=?")
	args = append(args, updateArgs...)

	filterSQL, filterArgs := m.qs.GetQuerySet()
	sql += filterSQL
	args = append(args, filterArgs...)

	return m.operator.Update(m.ctx(), m.conn, sql, args...)
}

func (m *Impl) Count() (num int64, err error) {
	if err = m.HaveError(); err != nil {
		return num, err
	}

	query := fmt.Sprintf("SELECT count(1) FROM %s", m.tableName)

	filterSQL, filterArgs := m.qs.GetQuerySet()

	query += filterSQL

	return m.operator.Count(m.ctx(), m.conn, query, filterArgs...)
}

func (m *Impl) FindOne() (result map[string]any, err error) {
	if methods, called := m.checkCalled(ctlGroupBy, ctlSelect); called {
		return result, fmt.Errorf(UnsupportedControllerError, methods, "FindOne")
	}

	if err = m.HaveError(); err != nil {
		return result, err
	}

	filterSQL, filterArgs := m.qs.GetQuerySet()

	query := SelectTemp
	query = fmt.Sprintf(query, m.fieldRows, m.tableName)
	query += filterSQL
	query += m.qs.GetOrderBySQL()
	query += " LIMIT 1"

	res, _ := DeepCopy(m.model)

	err = m.operator.FindOne(m.ctx(), m.conn, res, query, filterArgs...)

	switch {
	case err == nil:
		return Struct2Map(res, m.mTag), nil
	case errors.Is(err, ErrNotFound):
		return map[string]any{}, nil
	default:
		return map[string]any{}, err
	}
}

func (m *Impl) FindOneModel(modelPtr any) (err error) {
	if err = m.HaveError(); err != nil {
		return err
	}

	query := SelectTemp

	selectRows := m.qs.GetSelectSQL()
	if selectRows != "*" {
		query = fmt.Sprintf(query, selectRows, m.tableName)
	} else {
		query = fmt.Sprintf(query, m.fieldRows, m.tableName)
	}

	filterSQL, filterArgs := m.qs.GetQuerySet()

	query += filterSQL
	query += m.qs.GetGroupBySQL()
	query += m.qs.GetOrderBySQL()
	query += " LIMIT 1"

	return m.operator.FindOne(m.ctx(), m.conn, modelPtr, query, filterArgs...)
}

func (m *Impl) FindAll() (result []map[string]any, err error) {
	if methods, called := m.checkCalled(ctlGroupBy, ctlSelect); called {
		return result, fmt.Errorf(UnsupportedControllerError, methods, "FindAll")
	}

	if err = m.HaveError(); err != nil {
		return result, err
	}

	filterSQL, filterArgs := m.qs.GetQuerySet()

	query := SelectTemp
	query = fmt.Sprintf(query, m.fieldRows, m.tableName)
	query += filterSQL

	query += m.qs.GetOrderBySQL()
	query += m.qs.GetLimitSQL()

	res, _ := DeepCopy(m.modelSlice)

	err = m.operator.FindAll(m.ctx(), m.conn, res, query, filterArgs...)

	switch {
	case err == nil:
		return StructSlice2MapSlice(res, m.mTag), nil
	case errors.Is(err, ErrNotFound):
		return []map[string]any{}, nil
	default:
		return []map[string]any{}, err
	}
}

func (m *Impl) FindAllModel(modelSlicePtr any) (err error) {
	if err = m.HaveError(); err != nil {
		return err
	}

	query := SelectTemp

	selectRows := m.qs.GetSelectSQL()
	if selectRows != "*" {
		query = fmt.Sprintf(query, selectRows, m.tableName)
	} else {
		query = fmt.Sprintf(query, m.fieldRows, m.tableName)
	}

	filterSQL, filterArgs := m.qs.GetQuerySet()

	query += filterSQL
	query += m.qs.GetGroupBySQL()
	query += m.qs.GetOrderBySQL()
	query += m.qs.GetLimitSQL()

	err = m.operator.FindAll(m.ctx(), m.conn, modelSlicePtr, query, filterArgs...)

	switch {
	case err == nil:
		return nil
	case reflect.ValueOf(modelSlicePtr).Elem().Len() == 0:
		return ErrNotFound
	default:
		return err
	}
}

func (m *Impl) Delete() (int64, error) {
	if methods, called := m.checkCalled(ctlGroupBy, ctlSelect, ctlOrderBy); called {
		return 0, fmt.Errorf(UnsupportedControllerError, methods, "Delete")
	}

	data := map[string]any{"is_deleted": true}

	return m.Update(data)
}

func (m *Impl) Modify(data map[string]any) (num int64, err error) {
	return m.Exclude(data).Update(data)
}

func (m *Impl) Exist() (exist bool, err error) {
	if num, err := m.Count(); err != nil {
		return false, err
	} else if num > 0 {
		return true, nil
	}

	return false, nil
}

func (m *Impl) List() (total int64, data []map[string]any, err error) {
	if total, err = m.Count(); err != nil {
		return
	}

	if data, err = m.FindAll(); err != nil {
		return
	}

	return total, data, nil
}

func (m *Impl) GetOrCreate(data map[string]any) (map[string]any, error) {
	if _, err := m.Insert(data); err != nil {
		if !errors.Is(err, ErrDuplicateKey) {
			return nil, err
		}
	}

	return m.Filter(data).FindOne()
}

func (m *Impl) CreateOrUpdate(data map[string]any, filter ...any) (bool, int64, error) {
	if exist, err := m.Filter(filter...).Exist(); err != nil {
		return false, 0, err
	} else if exist {
		if num, err := m.Reset().Filter(filter...).Update(data); err != nil {
			return false, 0, err
		} else {
			return false, num, nil
		}
	}

	id, err := m.Insert(data)
	if err != nil {
		return false, 0, err
	}
	return true, id, nil
}

func (m *Impl) GetC2CMap(column1, column2 string) (res map[any]any, err error) {
	if err = m.HaveError(); err != nil {
		return res, err
	}

	if _, ok := m.fieldNameMap[column1]; !ok {
		return res, fmt.Errorf("column [%s] not exist", column1)
	}
	if _, ok := m.fieldNameMap[column2]; !ok {
		return res, fmt.Errorf("column [%s] not exist", column2)
	}

	query := fmt.Sprintf("SELECT `%s`,`%s` FROM %s ", column1, column2, m.tableName)

	filterSQL, filterArgs := m.qs.GetQuerySet()

	query += filterSQL
	query += m.qs.GetOrderBySQL()
	query += m.qs.GetLimitSQL()

	result, _ := DeepCopy(m.modelSlice)

	if err = m.operator.FindAll(m.ctx(), m.conn, res, query, filterArgs...); err != nil {
		return res, err
	}

	res = make(map[any]any)
	for _, v := range StructSlice2MapSlice(result, m.mTag) {
		res[v[column1]] = v[column2]
	}

	return res, nil
}

func (m *Impl) CreateIfNotExist(data map[string]any) (id int64, created bool, err error) {
	if exist, err := m.Filter(data).Exist(); err != nil {
		return 0, false, err
	} else if exist {
		return 0, false, nil
	}

	id, err = m.Insert(data)
	if err != nil {
		return 0, false, err
	}

	return id, true, nil
}
