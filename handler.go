package norm

import (
	"context"
	"errors"
	"fmt"
	"log"
	"reflect"
	"strings"

	"github.com/zeromicro/go-zero/core/logc"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/builder"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

const (
	SelectTemp      = "SELECT %s FROM %s"
	InsertTemp      = "INSERT INTO %s (%s) VALUES (%s)"
	UpdateTemp      = "UPDATE %s SET %s"
	DeleteTemp      = "DELETE FROM %s"
	DefaultModelTag = "db"
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
	ctlWhere   = controllerCall{Name: "Where", Flag: callWhere}
	ctlOrderBy = controllerCall{Name: "OrderBy", Flag: callOrderBy}
	ctlLimit   = controllerCall{Name: "Limit", Flag: callLimit}
	ctlSelect  = controllerCall{Name: "Select", Flag: callSelect}
	ctlGroupBy = controllerCall{Name: "GroupBy", Flag: callGroupBy}
)

var _ Controller = (*Impl)(nil)

type (
	Controller interface {
		ctx() context.Context
		hasCalled(f controllerCall) bool
		checkCalled(f ...controllerCall) ([]string, bool)
		values(values []string) string
		WithSession(session sqlx.Session) Controller
		Reset() Controller
		Filter(filter ...any) Controller
		Exclude(exclude ...any) Controller
		OrderBy(orderBy any) Controller
		Limit(pageSize, pageNum int64) Controller
		Select(columns any) Controller
		Where(cond string, args ...any) Controller
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

func (m *Impl) values(values []string) string {
	valueRows := ""

	for _, v := range values {
		if _, ok := m.fieldNameMap[v]; !ok {
			logc.Errorf(m.ctx(), "Key [%s] not exist.", v)
			continue
		}
		valueRows += fmt.Sprintf("`%s`,", v)
	}
	valueRows = strings.TrimRight(valueRows, ",")

	return valueRows
}

func (m *Impl) checkQuerySetError() {
	if m.qs.Error() != nil {
		logx.WithCallerSkip(3).WithContext(m.ctx()).Errorf("sql conditions queryset error: %s", m.qs.Error())
	}
}

func (m *Impl) querySetError() error {
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

func (m *Impl) Where(cond string, args ...any) Controller {
	m.setCalled(ctlWhere)

	m.qs.WhereToSQL(cond, args)
	return m
}

func (m *Impl) Exclude(exclude ...any) Controller {
	m.qs.FilterToSQL(1, exclude...)

	m.checkQuerySetError()

	return m
}

func (m *Impl) OrderBy(orderBy any) Controller {
	m.setCalled(ctlOrderBy)

	var orderByChecked []string

	v := reflect.ValueOf(orderBy)
	switch v.Kind() {
	case reflect.String:
		if orderBy.(string) == "" {
			return m
		}
		m.qs.OrderByToSQL(orderBy)
	case reflect.Slice, reflect.Array:
		orderByList, ok := orderBy.([]string)
		if !ok {
			logc.Error(m.ctx(), "Order by slice type should be string slice or string array")
			return m
		}
		if len(orderByList) == 0 {
			return m
		}

		for _, by := range orderByList {
			by = strings.TrimSpace(by)
			if strings.HasPrefix(by, "-") {
				if _, ok := m.fieldNameMap[by[1:]]; ok {
					orderByChecked = append(orderByChecked, by)
				} else {
					logc.Errorf(m.ctx(), "Order by key [%s] not exist.", by[1:])
					continue
				}
			} else {
				if _, ok := m.fieldNameMap[by]; ok {
					orderByChecked = append(orderByChecked, by)
				} else {
					logc.Errorf(m.ctx(), "Order by key [%s] not exist.", by)
					continue
				}
			}
		}
		m.qs.OrderByToSQL(orderByChecked)
	default:
		logc.Error(m.ctx(), "Order by type should be string, string slice or string array .")
		return m
	}

	return m
}

func (m *Impl) Limit(pageSize, pageNum int64) Controller {
	m.setCalled(ctlLimit)

	m.qs.LimitToSQL(pageSize, pageNum)
	return m
}

func (m *Impl) Select(columns any) Controller {
	m.setCalled(ctlSelect)

	m.qs.SelectToSQL(columns)
	return m
}

func (m *Impl) GroupBy(groupBy any) Controller {
	m.setCalled(ctlGroupBy)

	var groupBySliceChecked []string

	v := reflect.ValueOf(groupBy)
	switch v.Kind() {
	case reflect.String:
		if groupBy.(string) == "" {
			return m
		}
		m.qs.GroupByToSQL(groupBy)
	case reflect.Slice, reflect.Array:
		groupByList, ok := groupBy.([]string)
		if !ok {
			logc.Error(m.ctx(), "Group by type should be string slice or string array")
			return m
		}
		if len(groupByList) == 0 {
			return m
		}

		for _, by := range groupByList {
			by = strings.TrimSpace(by)
			if _, ok := m.fieldNameMap[by]; ok {
				groupBySliceChecked = append(groupBySliceChecked, by)
			} else {
				logc.Errorf(m.ctx(), "Group by key [%s] not exist.", by)
				continue
			}
		}
		m.qs.GroupByToSQL(groupBy)
	default:
		logc.Error(m.ctx(), "Group by type should be string, string slice or string array .")
		return m
	}

	return m
}

func (m *Impl) Insert(data map[string]any) (id int64, err error) {
	if methods, called := m.checkCalled(ctlFilter, ctlWhere, ctlOrderBy, ctlGroupBy, ctlSelect); called {
		return 0, fmt.Errorf(UnsupportedControllerError, methods, "Insert")
	}

	if err = m.querySetError(); err != nil {
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

	if err = m.querySetError(); err != nil {
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

	if err = m.querySetError(); err != nil {
		return num, err
	}

	var (
		args       []any
		updateRows []string
		updateArgs []any
	)

	for k, v := range data {
		if _, ok := m.fieldNameMap[k]; !ok {
			logc.Errorf(m.ctx(), "Key [%s] not exist.", k)
			continue
		}
		updateRows = append(updateRows, fmt.Sprintf("`%s`", k))
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
	if err = m.querySetError(); err != nil {
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

	if err = m.querySetError(); err != nil {
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
	if err = m.querySetError(); err != nil {
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

	if err = m.querySetError(); err != nil {
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
	if err = m.querySetError(); err != nil {
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
	if err = m.querySetError(); err != nil {
		return res, err
	}

	if _, ok := m.fieldNameMap[column1]; !ok {
		err = fmt.Errorf("column [%s] not exist", column1)
		logc.Errorf(m.ctx(), err.Error())
		return res, err
	}
	if _, ok := m.fieldNameMap[column2]; !ok {
		err = fmt.Errorf("column [%s] not exist", column2)
		logc.Errorf(m.ctx(), err.Error())
		return res, err
	}

	query := fmt.Sprintf("SELECT `%s`,`%s` FROM %s ", column1, column2, m.tableName)

	filterSQL, filterArgs := m.qs.GetQuerySet()

	query += filterSQL
	query += m.qs.GetOrderBySQL()
	query += m.qs.GetLimitSQL()

	result, _ := DeepCopy(m.modelSlice)

	if err = m.operator.FindAll(m.ctx(), m.conn, res, query, filterArgs...); err != nil {
		logc.Errorf(m.ctx(), "GetC2CMap querySetError: %+v", err)
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
