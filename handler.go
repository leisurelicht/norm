package norm

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"strings"
)

const (
	Asterisk   = "*"
	SelectTemp = "SELECT %s FROM %s"
	InsertTemp = "INSERT INTO %s (%s) VALUES (%s)"
	UpdateTemp = "UPDATE %s SET %s"
	DeleteTemp = "DELETE FROM %s"
)

const (
	SelectColumsValidateError   = "Select columns validate error: %s"
	SelectColumnsTypeError      = "Select type should be string or string slice"
	OrderByColumnsValidateError = "OrderBy columns validate error: [%s] not exist"
	OrderByColumnsTypeError     = "OrderBy type should be string or string slice"
	GroupByColumnsValidateError = "GroupBy columns validate error: %s"
	GroupByColumnsTypeError     = "GroupBy type should be string or string slice"
	CreateDataEmptyError        = "create data is empty"
	CreateDataTypeError         = "create data type is wrong, should not be [%s]"
	UpdateDataEmptyError        = "update data is empty"
	DataEmptyError              = "data is empty"
	UpdateColumnNotExistError   = "update column [%s] not exist"
	ColumnNotExistError         = "column [%s] not exist"
	MustBeCalledError           = "[%s] must be called after [%s]"
)

const (
	UnsupportedControllerError = "%s not supported for %s"
)

type controllerCall struct {
	Name string
	Flag callFlag
}

var (
	ctlFilter  = controllerCall{Name: "Filter", Flag: qsFilter}
	ctlExclude = controllerCall{Name: "Exclude", Flag: qsExclude}
	ctlWhere   = controllerCall{Name: "Where", Flag: qsWhere}
	ctlSelect  = controllerCall{Name: "Select", Flag: qsSelect}
	ctlLimit   = controllerCall{Name: "Limit", Flag: qsLimit}
	ctlOrderBy = controllerCall{Name: "OrderBy", Flag: qsOrderBy}
	ctlGroupBy = controllerCall{Name: "GroupBy", Flag: qsGroupBy}
	ctlHaving  = controllerCall{Name: "Having", Flag: qsHaving}
)

var _ Controller = (*Impl)(nil)

type (
	Controller interface {
		ctx() context.Context
		setCalled(f controllerCall)
		checkCalled(f ...controllerCall) ([]string, bool)
		validateColumns(columns []string) (validatedColumns []string, err error)
		setError(format string, a ...any)
		haveError() error
		Reset() Controller
		Filter(filter ...any) Controller
		Exclude(exclude ...any) Controller
		Where(cond string, args ...any) Controller
		Select(columns any) Controller
		Limit(pageSize, pageNum int64) Controller
		OrderBy(orderBy any) Controller
		GroupBy(groupBy any) Controller
		Having(having string, args ...any) Controller
		Create(data any) (idOrNum int64, err error)
		Remove() (num int64, err error)
		Update(data map[string]any) (num int64, err error)
		Count() (num int64, err error)
		FindOne() (result map[string]any, err error)
		FindOneModel(modelPtr any) (err error)
		FindAll() (result []map[string]any, err error)
		FindAllModel(modelSlicePtr any) (err error)
		Delete() (num int64, err error)
		Exist() (exist bool, error error)
		List() (num int64, data []map[string]any, err error)
		GetOrCreate(data map[string]any) (result map[string]any, err error)
		CreateOrUpdate(data map[string]any) (created bool, idOrNum int64, err error)
		CreateIfNotExist(data map[string]any) (id int64, created bool, err error)
	}

	Impl struct {
		context        context.Context
		conn           any
		modelPtr       any
		modelSlicePtr  any
		operator       Operator
		tableName      string
		fieldNameMap   map[string]struct{}
		fieldNameSlice []string
		fieldRows      string
		mTag           string
		qs             QuerySet
		called         callFlag
	}
)

// NewController creates a new Controller instance with the provided connection, operator, and model.
// It initializes the controller with the model's field names and prepares it for database operations.
// The model can be a struct or a slice of structs, and the operator must implement the Operator interface.
// The connection is used to execute queries, and the operator provides methods for database operations.
// It returns a function that takes a context and returns a Controller instance.
func NewController(conn any, op Operator, m any) func(ctx context.Context) Controller {
	// createModelPointerAndSlice call must be at the beginning of this function,
	// for it will check type of the m(model) is a struct
	mPtr, mSlicePtr := createModelPointerAndSlice(m)

	fieldNameSlice := rawFieldNames(m, op.DBTag(), true)

	tableName := getTableName(m)

	op.SetTableName(tableName)

	return func(ctx context.Context) Controller {
		if ctx == nil {
			ctx = context.Background()
		}
		return &Impl{
			context:        ctx,
			conn:           conn,
			modelPtr:       mPtr,
			modelSlicePtr:  mSlicePtr,
			operator:       op,
			tableName:      tableName,
			fieldNameMap:   strSlice2Map(fieldNameSlice),
			fieldNameSlice: fieldNameSlice,
			fieldRows:      strings.Join(rawFieldNames(m, op.DBTag(), false), ","),
			mTag:           op.DBTag(),
			qs:             NewQuerySet(op),
			called:         0,
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
	var unknownColumns []string
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

func (m *Impl) setError(format string, a ...any) {
	m.qs.setError(format, a...)
}

func (m *Impl) haveError() error {
	if err := m.qs.Error(); err != nil {
		return err
	}
	return nil
}

func (m *Impl) reset() {
	m.qs.Reset()
	m.called = 0
}

// Reset resets the controller's state.
// It clears any previous filters, selections, and other query parameters.
// This allows the controller to be reused for a new query without needing to create a new instance.
func (m *Impl) Reset() Controller {
	m.reset()
	return m
}

// Filter adds a filter condition containing objects that match the given lookup parameters
// It only accepts norm.Cond / norm.AND / norm.OR.
func (m *Impl) Filter(filter ...any) Controller {
	m.setCalled(ctlFilter)

	m.qs.FilterToSQL(notNot, filter...)

	return m
}

// Exclude adds an exclusion condition containing objects that do not match the given lookup parameters.
// It only accepts norm.Cond / norm.AND/norm.OR.
func (m *Impl) Exclude(exclude ...any) Controller {
	m.setCalled(ctlExclude)

	m.qs.FilterToSQL(isNot, exclude...)

	return m
}

// Where adds a WHERE clause to the query.
// It accepts a condition string and optional arguments.
// The condition string should be a valid SQL WHERE clause, and the arguments will be used to replace placeholders in the condition.
func (m *Impl) Where(cond string, args ...any) Controller {
	m.setCalled(ctlWhere)

	m.qs.WhereToSQL(cond, args)

	return m
}

// Select adds a SELECT clause to the query.
// It accepts a string or a slice of strings for selecting columns.
// If you pass a string, it should be a comma-separated list of columns and will not be validated.
// If you pass a slice, it will validate each column against the model's field names.
func (m *Impl) Select(selects any) Controller {
	m.setCalled(ctlSelect)

	switch sel := selects.(type) {
	case string:
		if sel == "" {
			return m
		}
		m.qs.StrSelectToSQL(sel)
	case []string:
		if len(sel) == 0 {
			return m
		}

		validatedColumns, err := m.validateColumns(sel)
		if err != nil {
			m.setError(SelectColumsValidateError, err)
			return m
		}

		m.qs.SliceSelectToSQL(validatedColumns)
	default:
		m.setError(SelectColumnsTypeError)
	}

	return m
}

// Limit adds a LIMIT clause to the query.
// It accepts pageSize and pageNum to control the number of records returned.
// It requires OrderBy to be called first. If OrderBy has not been called, it will return an error.
// Both Page size and page number should be greater than 0.
func (m *Impl) Limit(pageSize, pageNum int64) Controller {
	m.setCalled(ctlLimit)

	if !m.hasCalled(ctlOrderBy) {
		m.setError(MustBeCalledError, "Limit", "OrderBy")
		return m
	}

	m.qs.LimitToSQL(pageSize, pageNum)
	return m
}

// OrderBy adds an ORDER BY clause to the query.
// It accepts a string or a slice of strings for ordering columns.
// If you pass a string, it should be a comma-separated list of columns and will not be validated. So never pass a parameter from user directly.
// If you pass a slice, it will validate each column against the model's field names.
func (m *Impl) OrderBy(orderBy any) Controller {
	m.setCalled(ctlOrderBy)

	var validatedOrderBy []string

	switch orderByVal := orderBy.(type) {
	case string:
		if orderByVal == "" {
			return m
		}
		m.qs.StrOrderByToSQL(orderByVal)
	case []string:
		if len(orderByVal) == 0 {
			return m
		}
		unknownColumns := []string{}
		for _, by := range orderByVal {
			if by == "" {
				continue
			}
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
			m.setError(OrderByColumnsValidateError, strings.Join(unknownColumns, "; "))
			return m
		}

		m.qs.OrderByToSQL(validatedOrderBy)
	default:
		m.setError(OrderByColumnsTypeError)
	}

	return m
}

// GroupBy adds a GROUP BY clause to the query.
// It accepts a string or a slice of strings for grouping columns.
// If you pass a string, it should be a comma-separated list of columns and will not be validated.
// If you pass a slice, it will validate each column against the model's field names.
func (m *Impl) GroupBy(groupBy any) Controller {
	m.setCalled(ctlGroupBy)

	switch gb := groupBy.(type) {
	case string:
		if gb == "" {
			return m
		}
		m.qs.StrGroupByToSQL(gb)
	case []string:
		if len(gb) == 0 {
			return m
		}

		validatedColumns, err := m.validateColumns(gb)
		if err != nil {
			m.setError(GroupByColumnsValidateError, err)
			return m
		}

		m.qs.SliceGroupByToSQL(validatedColumns)
	default:
		m.setError(GroupByColumnsTypeError)
		return m
	}

	return m
}

// Having adds a HAVING clause to the query.
func (m *Impl) Having(having string, args ...any) Controller {
	m.setCalled(ctlHaving)
	m.qs.HavingToSQL(having, args)
	return m
}

func (m *Impl) create(data map[string]any) (id int64, err error) {
	if len(data) == 0 {
		return 0, errors.New(CreateDataEmptyError)
	}

	var (
		rows []string
		args []any
	)

	for _, k := range m.fieldNameSlice {
		if _, ok := data[k]; !ok {
			continue
		}
		rows = append(rows, fmt.Sprintf("`%s`", k))
		args = append(args, data[k])
	}

	sql := fmt.Sprintf(InsertTemp, m.tableName, strings.Join(rows, ","), strings.Repeat("?,", len(rows)-1)+"?")

	return m.operator.Insert(m.ctx(), m.conn, sql, args...)
}

func (m *Impl) bulkCreate(data []map[string]any) (num int64, err error) {
	if methods, called := m.checkCalled(ctlFilter, ctlExclude, ctlWhere, ctlSelect, ctlGroupBy, ctlHaving, ctlOrderBy); called {
		return 0, fmt.Errorf(UnsupportedControllerError, methods, "BulkInsert")
	}

	if err = m.haveError(); err != nil {
		return 0, err
	}

	if len(data) == 0 {
		return 0, errors.New(CreateDataEmptyError)
	}

	var (
		rows []string
		args []string
	)

	for _, k := range m.fieldNameSlice {
		if _, ok := data[0][k]; !ok {
			continue
		}
		rows = append(rows, fmt.Sprintf("`%s`", k))
		args = append(args, k)
	}

	sql := fmt.Sprintf(InsertTemp, m.tableName, strings.Join(rows, ","), strings.Repeat("?,", len(rows)-1)+"?")

	return m.operator.BulkInsert(m.ctx(), m.conn, sql, args, data)
}

// Create creates a new record in the database with the provided data map.
// It returns the ID of the created record or the number of records inserted, and any error encountered.
func (m *Impl) Create(data any) (idOrNum int64, err error) {
	if methods, called := m.checkCalled(ctlFilter, ctlExclude, ctlWhere, ctlSelect, ctlOrderBy, ctlGroupBy, ctlHaving); called {
		return 0, fmt.Errorf(UnsupportedControllerError, methods, "Create")
	}

	switch d := data.(type) {
	case map[string]any:
		return m.create(d)
	case []map[string]any:
		return m.bulkCreate(d)
	default:
		v := reflect.ValueOf(data)
		if v.Kind() == reflect.Ptr {
			v = v.Elem()
		}
		if v.Kind() == reflect.Struct {
			return m.create(modelStruct2Map(data, m.mTag))
		} else if v.Kind() == reflect.Slice {
			return m.bulkCreate(modelStructSlice2MapSlice(data, m.mTag))
		}
	}
	return 0, fmt.Errorf(CreateDataTypeError, reflect.TypeOf(data).Kind())
}

// Remove deletes the records matching the current query set.
// It returns the number of records deleted and any error encountered.
// Note: This method will really remove records from the database
func (m *Impl) Remove() (num int64, err error) {
	if methods, called := m.checkCalled(ctlSelect, ctlGroupBy, ctlHaving); called {
		return 0, fmt.Errorf(UnsupportedControllerError, methods, "Remove")
	}

	if err = m.haveError(); err != nil {
		return num, err
	}

	sql := fmt.Sprintf(DeleteTemp, m.tableName)

	filterSQL, filterArgs := m.qs.GetQuerySet()
	sql += filterSQL

	return m.operator.Remove(m.ctx(), m.conn, sql, filterArgs...)
}

func (m *Impl) update(data map[string]any) (num int64, err error) {
	var (
		args       []any
		updateRows []string
		updateArgs []any
	)

	for k, v := range data {
		if _, ok := m.fieldNameMap[k]; !ok {
			return 0, fmt.Errorf(UpdateColumnNotExistError, k)
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

// Update updates the records matching the current query set with the provided data map.
// It returns the number of records updated and any error encountered.
func (m *Impl) Update(data map[string]any) (num int64, err error) {
	if methods, called := m.checkCalled(ctlSelect, ctlGroupBy, ctlHaving); called {
		return 0, fmt.Errorf(UnsupportedControllerError, methods, "Update")
	}

	if err = m.haveError(); err != nil {
		return num, err
	}

	if len(data) == 0 {
		return 0, errors.New(UpdateDataEmptyError)
	}

	return m.update(data)
}

// Count retrieves the total number of records matching the current query set.
// It returns the total count and any error encountered.
func (m *Impl) Count() (num int64, err error) {
	if err = m.haveError(); err != nil {
		return num, err
	}

	filterSQL, filterArgs := m.qs.GetQuerySet()

	return m.operator.Count(m.ctx(), m.conn, filterSQL, filterArgs...)
}

func (m *Impl) findOne() (result map[string]any, err error) {
	filterSQL, filterArgs := m.qs.GetQuerySet()

	query := SelectTemp
	query = fmt.Sprintf(query, m.fieldRows, m.tableName)
	query += filterSQL
	query += m.qs.GetGroupBySQL()
	havingSQL, havingArgs := m.qs.GetHavingSQL()
	query += havingSQL
	filterArgs = append(filterArgs, havingArgs...)
	query += m.qs.GetOrderBySQL()
	query += " LIMIT 1"

	res := deepCopyModelPtrStructure(m.modelPtr)

	err = m.operator.FindOne(m.ctx(), m.conn, res, query, filterArgs...)

	switch {
	case err == nil:
		return modelStruct2Map(res, m.mTag), nil
	case errors.Is(err, ErrNotFound):
		return map[string]any{}, nil
	default:
		return map[string]any{}, err
	}
}

// FindOne retrieves a single record matching the current query set into a map.
// It returns the data as a map, or an error if the operation fails.
func (m *Impl) FindOne() (result map[string]any, err error) {
	if methods, called := m.checkCalled(ctlSelect, ctlHaving); called {
		return result, fmt.Errorf(UnsupportedControllerError, methods, "FindOne")
	}

	if err = m.haveError(); err != nil {
		return result, err
	}

	return m.findOne()
}

// FindOneModel retrieves a single record matching the current query set into a model.
// It returns an error if the model type is not a pointer to a struct.
func (m *Impl) FindOneModel(modelPtr any) (err error) {
	if err = m.haveError(); err != nil {
		return err
	}

	rv := reflect.ValueOf(modelPtr)
	if rv.Kind() != reflect.Ptr || rv.Elem().Kind() != reflect.Struct {
		return fmt.Errorf("model must be a pointer to struct")
	}

	query := SelectTemp

	selectRows := m.qs.GetSelectSQL()
	if selectRows != Asterisk {
		query = fmt.Sprintf(query, selectRows, m.tableName)
	} else {
		query = fmt.Sprintf(query, m.fieldRows, m.tableName)
	}

	filterSQL, filterArgs := m.qs.GetQuerySet()

	query += filterSQL
	query += m.qs.GetGroupBySQL()
	havingSQL, havingArgs := m.qs.GetHavingSQL()
	query += havingSQL
	filterArgs = append(filterArgs, havingArgs...)
	query += m.qs.GetOrderBySQL()
	query += " LIMIT 1"

	return m.operator.FindOne(m.ctx(), m.conn, modelPtr, query, filterArgs...)
}

// FindAll retrieves all records matching the current query set into a slice of maps.
// It returns the data as a slice of maps, or an error if the operation fails.
func (m *Impl) FindAll() (result []map[string]any, err error) {
	if methods, called := m.checkCalled(ctlSelect, ctlHaving); called {
		return result, fmt.Errorf(UnsupportedControllerError, methods, "FindAll")
	}

	if err = m.haveError(); err != nil {
		return result, err
	}

	filterSQL, filterArgs := m.qs.GetQuerySet()

	query := SelectTemp
	query = fmt.Sprintf(query, m.fieldRows, m.tableName)
	query += filterSQL

	query += m.qs.GetGroupBySQL()
	query += m.qs.GetOrderBySQL()
	query += m.qs.GetLimitSQL()

	res := deepCopyModelPtrStructure(m.modelSlicePtr)

	err = m.operator.FindAll(m.ctx(), m.conn, res, query, filterArgs...)

	switch {
	case err == nil:
		return modelStructSlice2MapSlice(res, m.mTag), nil
	case errors.Is(err, ErrNotFound):
		return []map[string]any{}, nil
	default:
		return []map[string]any{}, err
	}
}

// FindAllModel retrieves all records matching the current query set into a slice of models.
// It returns an error if the model type is not a pointer to a slice.
func (m *Impl) FindAllModel(modelSlicePtr any) (err error) {
	if err = m.haveError(); err != nil {
		return err
	}

	rv := reflect.ValueOf(modelSlicePtr)
	if rv.Kind() != reflect.Ptr || rv.Elem().Kind() != reflect.Slice {
		return fmt.Errorf("model must be a pointer to slice")
	}

	query := SelectTemp

	selectRows := m.qs.GetSelectSQL()
	if selectRows != Asterisk {
		query = fmt.Sprintf(query, selectRows, m.tableName)
	} else {
		query = fmt.Sprintf(query, m.fieldRows, m.tableName)
	}

	filterSQL, filterArgs := m.qs.GetQuerySet()

	query += filterSQL
	query += m.qs.GetGroupBySQL()
	havingSQL, havingArgs := m.qs.GetHavingSQL()
	query += havingSQL
	filterArgs = append(filterArgs, havingArgs...)
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

// Delete marks the records as deleted by setting the 'is_deleted' field to true.
// It returns the number of records marked as deleted.
// Note: This method is not a true delete operation; it only marks records as deleted.
func (m *Impl) Delete() (int64, error) {
	if methods, called := m.checkCalled(ctlGroupBy, ctlSelect, ctlOrderBy); called {
		return 0, fmt.Errorf(UnsupportedControllerError, methods, "Delete")
	}

	data := map[string]any{"is_deleted": true}

	return m.Update(data)
}

func (m *Impl) exist() (exist bool, err error) {
	filterSQL, filterArgs := m.qs.GetQuerySet()

	return m.operator.Exist(m.ctx(), m.conn, filterSQL, filterArgs...)
}

// Exist checks if any record exists that matches the current query set.
func (m *Impl) Exist() (exist bool, err error) {
	if methods, called := m.checkCalled(ctlGroupBy, ctlSelect); called {
		return false, fmt.Errorf(UnsupportedControllerError, methods, "Exist")
	}

	if err = m.haveError(); err != nil {
		return exist, err
	}

	return m.exist()
}

// List retrieves the total count and all data matching the current query set.
// It returns the total count, data as a slice of maps, and any error encountered.
func (m *Impl) List() (total int64, data []map[string]any, err error) {
	if methods, called := m.checkCalled(ctlSelect, ctlHaving); called {
		return total, data, fmt.Errorf(UnsupportedControllerError, methods, "List")
	}

	if total, err = m.Count(); err != nil {
		return
	}

	if data, err = m.FindAll(); err != nil {
		return
	}

	return total, data, nil
}

// GetOrCreate creates a new record if it does not already exist, or returns the existing record.
func (m *Impl) GetOrCreate(data map[string]any) (res map[string]any, err error) {
	if methods, called := m.checkCalled(ctlSelect, ctlGroupBy, ctlHaving); called {
		return res, fmt.Errorf(UnsupportedControllerError, methods, "GetOrCreate")
	}

	if err = m.haveError(); err != nil {
		return res, err
	}

	if len(data) == 0 {
		return res, errors.New(DataEmptyError)
	}

	if _, err := m.create(data); err != nil {
		if !errors.Is(err, ErrDuplicateKey) {
			return res, err
		}
	}

	m.setCalled(ctlFilter)
	m.qs.FilterToSQL(notNot, Cond(data))

	return m.findOne()
}

// CreateOrUpdate creates a new record if it does not already exist, or updates the existing record.
func (m *Impl) CreateOrUpdate(data map[string]any) (created bool, numOrID int64, err error) {
	if methods, called := m.checkCalled(ctlSelect, ctlGroupBy, ctlHaving); called {
		return false, 0, fmt.Errorf(UnsupportedControllerError, methods, "CreateOrUpdate")
	}

	if err = m.haveError(); err != nil {
		return false, 0, err
	}

	if len(data) == 0 {
		return false, 0, errors.New(DataEmptyError)
	}

	if exist, err := m.exist(); err != nil {
		return false, 0, err
	} else if exist {
		if num, err := m.update(data); err != nil {
			return false, 0, err
		} else {
			return false, num, nil
		}
	}

	m.reset()
	id, err := m.create(data)
	if err != nil {
		return false, 0, err
	}
	return true, id, nil
}

// CreateIfNotExist creates a new record if it does not already exist.
// If data not exist, it will create a new record and return the 'id' column value(if exists) and created true.
// if data exist, it will return 0 and created false
func (m *Impl) CreateIfNotExist(data map[string]any) (id int64, created bool, err error) {
	if methods, called := m.checkCalled(ctlSelect, ctlGroupBy, ctlHaving); called {
		return 0, false, fmt.Errorf(UnsupportedControllerError, methods, "CreateIfNotExist")
	}

	if err = m.haveError(); err != nil {
		return 0, false, err
	}

	if len(data) == 0 {
		return 0, false, errors.New(DataEmptyError)
	}

	m.setCalled(ctlFilter)
	m.qs.FilterToSQL(notNot, Cond(data))

	if exist, err := m.exist(); err != nil {
		return 0, false, err
	} else if exist {
		return 0, false, nil
	}

	id, err = m.create(data)
	if err != nil {
		return 0, false, err
	}

	return id, true, nil
}
