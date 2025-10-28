package norm

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

type callFlag int64

const (
	defaultOuterFilterCondsLen = 10
	defaultInnerFilterCondsLen = 10
	orPrefix                   = "| "
	notPrefix                  = "not_"
	descPrefix                 = "-"
	operatorJoiner             = "__"
	plural                     = "~"
	methodJoiner               = "##"
)

const (
	_exact   = "exact"
	_exclude = "exclude"
	_iexact  = "iexact"
	_gt      = "gt"
	_gte     = "gte"
	_lt      = "lt"
	_lte     = "lte"
	_len     = "len"

	_in          = "in"
	_between     = "between"
	_contains    = "contains"
	_icontains   = "icontains"
	_startswith  = "startswith"
	_istartswith = "istartswith"
	_endswith    = "endswith"
	_iendswith   = "iendswith"

	_isNull = "is_null"
)

const (
	argsLenError          = "args length must be equal to ? number"
	orderKeyTypeError     = "order key value must be a list of string"
	orderKeyLenError      = "order key length must be equal to filter key length"
	isNotValueError       = "isNot value must be 0 or 1"
	paramTypeError        = "param type must be string or slice of string"
	pageSizeORNumberError = "page size and page number must be positive"

	filterOrWhereError          = "[%s] or [Where] can not be called at the same time"
	fieldLookupError            = "field lookups [%s] is invalid"
	unknownOperatorError        = "unknown operator [%s]"
	notImplementedOperatorError = "not implemented operator [%s]"
	unsupportedValueError       = "operator [%s] unsupported value type [%s]"
	operatorValueLenError       = "operator [%s] value length must be [%d]"
	operatorValueLenLessError   = "operator [%s] value length must greater than [%d]"
	operatorValueTypeError      = "operator [%s] value must be string list"
	unsupportedFilterTypeError  = "unsupported filter type [%s], Please use be [Cond | AND | OR]"
	operatorValueEmptyError     = "operator [%s] unsupported value empty"
)

const (
	qsFilter callFlag = 1 << iota
	qsExclude
	qsWhere
	qsOrderBy
	qsLimit
	qsSelect
	qsGroupBy
	qsHaving
)

const (
	isFilter, isExclude                = 0, 1
	notNot, isNot                      = 0, 1
	andTag, orTag, andNotTag, orNotTag = 0, 1, 2, 3
)

var (
	filterAndExclude = []string{"Filter", "Exclude"}
	not              = [2]string{"", " NOT"}
	conjunctions     = [4]string{"AND", "OR", "AND NOT", "OR NOT"}
	// Define a map for conjunction types and their tags
	conjunctionMap = map[reflect.Type]int{
		reflect.TypeOf(Cond{}): andTag,
		reflect.TypeOf(AND{}):  andTag,
		reflect.TypeOf(OR{}):   orTag,
	}
)

type cond struct {
	Conj string
	SQL  string
	Args []any
}

func newCond() *cond {
	return &cond{}
}

func (q *cond) SetConj(conj string) *cond {
	q.Conj = conj
	return q
}

func (q *cond) SetSQL(sql string, args []any) *cond {
	q.SQL = sql
	q.Args = args
	return q
}

func newCondByValue(conj, sql string, args []any) *cond {
	return &cond{
		Conj: conj,
		SQL:  sql,
		Args: args,
	}
}

type QuerySet interface {
	setCalled(f callFlag)
	hasCalled(f callFlag) bool
	setError(format string, a ...any)
	Error() error
	Reset()
	GetQuerySet() (string, []any)
	FilterToSQL(notTag int, filter ...any) QuerySet
	WhereToSQL(cond string, args ...any) QuerySet
	GetSelectSQL() string
	SelectToSQL(columns any) QuerySet
	StrSelectToSQL(columns string) QuerySet
	SliceSelectToSQL(columns []string) QuerySet
	GetLimitSQL() string
	LimitToSQL(pageSize, pageNum int64) QuerySet
	GetOrderBySQL() string
	OrderByToSQL(orderBy any) QuerySet
	StrOrderByToSQL(orderBy string) QuerySet
	SliceOrderByToSQL(orderBy []string) QuerySet
	GetGroupBySQL() string
	GroupByToSQL(groupBy any) QuerySet
	StrGroupByToSQL(groupBy string) QuerySet
	SliceGroupByToSQL(groupBy []string) QuerySet
	GetHavingSQL() (string, []any)
	HavingToSQL(having string, args ...any) QuerySet
}

type QuerySetImpl struct {
	Operator
	selectColumn  string
	whereCond     cond
	filterConds   [][]cond
	filterConjTag []int
	orderBySQL    string
	limitSQL      string
	groupSQL      string
	havingSQL     cond
	err           error
	called        callFlag
}

var _ QuerySet = (*QuerySetImpl)(nil)

func NewQuerySet(op Operator) QuerySet {
	return &QuerySetImpl{
		Operator:      op,
		selectColumn:  "*",
		whereCond:     cond{},
		filterConds:   make([][]cond, 0, defaultOuterFilterCondsLen),
		filterConjTag: make([]int, 0, defaultOuterFilterCondsLen),
		orderBySQL:    "",
		limitSQL:      "",
		groupSQL:      "",
		err:           nil,
	}
}

func (p *QuerySetImpl) setCalled(f callFlag) {
	p.called = p.called | f
}

func (p *QuerySetImpl) hasCalled(f callFlag) bool {
	return p.called&f == f
}

func (p *QuerySetImpl) setError(format string, a ...any) {
	p.err = fmt.Errorf(format, a...)
}

func (p *QuerySetImpl) Error() error {
	return p.err
}

func (p *QuerySetImpl) Reset() {
	p.selectColumn = "*"
	p.whereCond = cond{}
	p.filterConds = make([][]cond, 0, defaultOuterFilterCondsLen)
	p.filterConjTag = make([]int, 0, defaultOuterFilterCondsLen)
	p.orderBySQL = ""
	p.limitSQL = ""
	p.groupSQL = ""
	p.havingSQL = cond{}
	p.err = nil
	p.called = 0
}

func (p *QuerySetImpl) GetQuerySet() (sql string, args []any) {
	// Handle the case with direct WHERE condition
	if p.whereCond.SQL != "" {
		return " WHERE " + p.whereCond.SQL, p.whereCond.Args
	}

	// Early return for no filter conditions
	if len(p.filterConds) == 0 {
		return "", nil
	}

	// Pre-calculate approximate capacity for string builder
	totalConditions := 0
	for _, filterList := range p.filterConds {
		totalConditions += len(filterList)
	}

	// Initial capacity estimate: 8 chars per condition + conjunctions
	outerSQL := strings.Builder{}
	outerSQL.Grow(totalConditions*20 + len(p.filterConds)*10)
	outerSQL.WriteString(" WHERE ")

	args = make([]any, 0, totalConditions*2) // Estimate arg count

	for i, condList := range p.filterConds {
		// Add conjunction between filter groups
		if i > 0 {
			outerSQL.WriteString(" ")
			outerSQL.WriteString(conjunctions[p.filterConjTag[i]])
			outerSQL.WriteString(" ")
		}

		// Check if this is a NOT condition by examining the first filter in the group
		// The isNot flag affects the entire filter group
		isNot := strings.Contains(condList[0].Conj, "NOT")
		if isNot {
			outerSQL.WriteString("NOT ")
		}

		// Single condition doesn't need inner parentheses
		if len(condList) == 1 {
			outerSQL.WriteString("(")
			outerSQL.WriteString(condList[0].SQL)
			outerSQL.WriteString(")")
			args = append(args, condList[0].Args...)
			continue
		}

		// Multiple conditions in this group
		outerSQL.WriteString("(")

		// First condition
		outerSQL.WriteString("(")
		outerSQL.WriteString(condList[0].SQL)
		outerSQL.WriteString(")")
		args = append(args, condList[0].Args...)

		// Remaining conditions with their conjunctions
		for _, filter := range condList[1:] {
			// Extract the base conjunction without NOT suffix for inner conditions
			baseConj := filter.Conj
			if isNot {
				baseConj = strings.ReplaceAll(baseConj, " NOT", "")
			}

			outerSQL.WriteString(" ")
			outerSQL.WriteString(baseConj)
			outerSQL.WriteString(" (")
			outerSQL.WriteString(filter.SQL)
			outerSQL.WriteString(")")
			args = append(args, filter.Args...)
		}

		outerSQL.WriteString(")")
	}

	return outerSQL.String(), args
}

func (p *QuerySetImpl) filterHandler(filter map[string]any) (filterSql string, filterArgs []any) {
	if len(filter) == 0 {
		return
	}

	var (
		filterConds = make(map[string]*cond, len(filter))
		skList      []string
		isOrder     = false
		fieldName   string
		operator    string
		andOrFlag   = andTag
		notFlag     = notNot
		method      string
	)

	if sk, ok := filter[SortKey]; ok {
		delete(filter, SortKey)
		skList, ok = sk.([]string)
		if !ok {
			p.setError(orderKeyTypeError)
			return
		}
		if len(skList) != len(filter) {
			p.setError(orderKeyLenError)
			return
		}
		isOrder = true
	}

	for fieldLookup, filedValue := range filter {
		if strings.HasPrefix(fieldLookup, orPrefix) {
			fieldLookup = strings.TrimPrefix(fieldLookup, orPrefix)
			andOrFlag = orTag
		} else {
			andOrFlag = andTag
		}

		if pos := strings.Index(fieldLookup, methodJoiner); pos != -1 {
			method = fieldLookup[pos+len(methodJoiner):]
			fieldLookup = fieldLookup[:pos]
		} else {
			method = ""
		}

		lookup := strings.Split(fieldLookup, operatorJoiner)
		if len(lookup) == 1 {
			operator = _exact
			notFlag = notNot
		} else if len(lookup) == 2 {
			operator = strings.ToLower(strings.TrimSpace(lookup[1]))
			if strings.HasPrefix(operator, notPrefix) {
				operator = strings.TrimPrefix(operator, notPrefix)
				notFlag = isNot
			} else {
				notFlag = notNot
			}
		} else {
			p.setError(fieldLookupError, fieldLookup)
			return
		}

		op := p.OperatorSQL(operator, method)
		if op == "" {
			p.setError(unknownOperatorError, operator)
			return
		}

		fieldName = lookup[0]

		fCond := newCond()
		fCond.SetConj(conjunctions[andOrFlag])

		valueOf := reflect.ValueOf(filedValue)
		valueKind := valueOf.Kind()
		switch operator {
		case _exact:
			if filedValue == nil {
				// should generate sql like "fieldName IS NULL"
				filterConds[fieldName] = fCond.SetSQL(fmt.Sprintf(p.OperatorSQL(_isNull, ""), fieldName), []any{})
				break
			}
			// the value arrived here is not nil, so go to the next case for processing
			fallthrough
		case _exclude, _iexact:
			if isStringKind(valueKind) || isBoolKind(valueKind) || isNumericKind(valueKind) {
				filterConds[fieldName] = fCond.SetSQL(fmt.Sprintf(op, fieldName), []any{filedValue})
			} else if isListKind(valueKind) {
				if valueOf.Len() == 0 {
					p.setError(operatorValueLenLessError, operator, 0)
					return
				}
				opStr := " " + conjunctions[0] + " " + fmt.Sprintf(op, fieldName)
				sql := fmt.Sprintf(op, fieldName) + strings.Repeat(opStr, valueOf.Len()-1)
				if len(filter) > 1 {
					sql = "(" + sql + ")"
				}
				args := make([]any, valueOf.Len())
				for i := range valueOf.Len() {
					args[i] = valueOf.Index(i).Interface()
				}
				filterConds[fieldName] = fCond.SetSQL(sql, args)
			} else {
				p.setError(unsupportedValueError, operator, valueKind.String())
				return
			}
		case _gt, _gte, _lt, _lte, _len:
			if !isNumericKind(valueKind) {
				p.setError(unsupportedValueError, operator, valueKind.String())
				return
			}
			filterConds[fieldName] = fCond.SetSQL(fmt.Sprintf(op, fieldName), []any{filedValue})
		case _in:
			if isStringKind(valueKind) {
				sql := fmt.Sprintf(op, fieldName, not[notFlag]) + " (" + filedValue.(string) + ")"
				filterConds[fieldName] = fCond.SetSQL(sql, []any{})
				continue
			}

			if !isListKind(valueKind) {
				p.setError(unsupportedValueError, operator, valueKind.String())
				return
			}
			if valueOf.Len() == 0 {
				p.setError(operatorValueLenLessError, operator, 0)
				return
			}
			sql := fmt.Sprintf(op, fieldName, not[notFlag]) + " (" + p.GetPlaceholder() + strings.Repeat(","+p.GetPlaceholder(), valueOf.Len()-1) + ")"
			args := make([]any, valueOf.Len())
			for i := 0; i < valueOf.Len(); i++ {
				args[i] = valueOf.Index(i).Interface()
			}
			filterConds[fieldName] = fCond.SetSQL(sql, args)
		case _between:
			if !isListKind(valueKind) {
				p.setError(unsupportedValueError, operator, valueKind.String())
				return
			}
			if valueOf.Len() != 2 {
				p.setError(operatorValueLenError, operator, 2)
				return
			}
			sql := fmt.Sprintf(op, fieldName, not[notFlag])
			args := make([]any, valueOf.Len())
			for i := 0; i < valueOf.Len(); i++ {
				args[i] = valueOf.Index(i).Interface()
			}
			filterConds[fieldName] = fCond.SetSQL(sql, args)
		case _contains, _icontains, _startswith, _istartswith, _endswith, _iendswith:
			valueFormat := "%%%v%%"
			if operator == _startswith || operator == _istartswith {
				valueFormat = "%v%%"
			} else if operator == _endswith || operator == _iendswith {
				valueFormat = "%%%v"
			}

			if isStringKind(valueKind) {
				if valueOf.IsZero() {
					p.setError(unsupportedValueError, operator, "blank string")
					return
				}
				filterConds[fieldName] = fCond.SetSQL(fmt.Sprintf(op, fieldName, not[notFlag]), []any{fmt.Sprintf(valueFormat, filedValue)})
			} else if isListKind(valueKind) {
				if valueOf.Len() == 0 {
					p.setError(operatorValueLenLessError, operator, 0)
					return
				}
				if !isStrList(filedValue) {
					p.setError(operatorValueTypeError, operator)
					return
				}
				genStrListValueLikeSQL(p, filterConds, fieldName, valueOf, notFlag, operator, valueFormat)
			} else {
				p.setError(unsupportedValueError, operator, valueKind.String())
				return
			}
		default:
			p.setError(notImplementedOperatorError, op)
			continue
		}
	}

	if len(filterConds) == 0 {
		return filterSql, filterArgs
	}

	if isOrder {
		for index, key := range skList {
			if condition, ok := filterConds[key]; ok {
				joinSQL(&filterSql, &filterArgs, index, condition)
			}
		}
	} else {
		index := 0
		for _, condition := range filterConds {
			joinSQL(&filterSql, &filterArgs, index, condition)
			index++
		}
	}

	return filterSql, filterArgs
}

func (p *QuerySetImpl) FilterToSQL(state int, filter ...any) QuerySet {
	if !p.hasCalled(qsWhere) {
		if state == isFilter {
			p.setCalled(qsFilter)
		} else if state == isExclude {
			p.setCalled(qsExclude)
		} else {
			p.setError(isNotValueError)
			return p
		}
	} else {
		p.setError(filterOrWhereError, filterAndExclude[state])
		return p
	}

	if len(filter) == 0 {
		return p
	}

	var (
		arg         map[string]any
		conjFlag    = andTag
		filterConds = make([]cond, 0, defaultInnerFilterCondsLen)
	)

	for i, f := range filter {
		if f == nil {
			p.setError(unsupportedFilterTypeError, "nil")
			return p
		}

		// Set the conjunction tag for the first filter
		if i == 0 {
			// Use the map to determine the conjunction tag
			conjTag, ok := conjunctionMap[reflect.TypeOf(f)]
			if !ok {
				p.setError(unsupportedFilterTypeError, reflect.TypeOf(f).String())
				return p // Return immediately if there's an error
			}
			p.filterConjTag = append(p.filterConjTag, conjTag)
		}

		switch v := f.(type) {
		case Cond:
			arg, conjFlag = v, andTag
		case AND:
			arg, conjFlag = v, andTag
		case OR:
			arg, conjFlag = v, orTag
		default:
			p.setError(unsupportedFilterTypeError, reflect.TypeOf(f).String())
		}

		if filterSQL, filterArgs := p.filterHandler(arg); filterSQL == "" {
			continue
		} else {
			// Only add "NOT" to the conjunction for the first condition
			// The rest will be handled in GetQuerySet
			conjStr := conjunctions[conjFlag]
			if state == 1 && len(filterConds) == 0 {
				conjStr = conjunctions[conjFlag+2] // Use AND NOT or OR NOT
			}
			filterConds = append(filterConds, *newCondByValue(conjStr, filterSQL, filterArgs))
		}
	}

	if len(filterConds) > 0 {
		p.filterConds = append(p.filterConds, filterConds)
	}

	return p
}

func (p *QuerySetImpl) WhereToSQL(cond string, args ...any) QuerySet {
	if !p.hasCalled(qsFilter) && !p.hasCalled(qsExclude) {
		p.setCalled(qsWhere)
	} else if p.hasCalled(qsFilter) {
		p.setError(filterOrWhereError, filterAndExclude[isFilter])
		return p
	} else if p.hasCalled(qsExclude) {
		p.setError(filterOrWhereError, filterAndExclude[isExclude])
		return p
	}

	num := strings.Count(cond, "?")
	if num > 0 && len(args) != num {
		p.setError(argsLenError)
		return p
	}
	p.whereCond.SQL = cond
	p.whereCond.Args = args

	return p
}

func (p *QuerySetImpl) GetSelectSQL() string {
	return p.selectColumn
}

func (p *QuerySetImpl) SelectToSQL(columns any) QuerySet {
	p.setCalled(qsSelect)

	switch cols := columns.(type) {
	case string:
		p.StrSelectToSQL(cols)
	case []string:
		p.SliceSelectToSQL(cols)
	default:
		p.setError(paramTypeError)
	}

	return p
}

func (p *QuerySetImpl) StrSelectToSQL(columns string) QuerySet {
	p.setCalled(qsSelect)

	p.selectColumn = columns
	return p
}

func (p *QuerySetImpl) SliceSelectToSQL(columns []string) QuerySet {
	p.setCalled(qsSelect)

	if len(columns) == 0 {
		return p
	}

	var result strings.Builder
	result.Grow(len(columns) * 10) // Pre-allocate space for performance
	for i := 0; i < len(columns)-1; i++ {
		result.WriteString(wrapWithBackticks(columns[i]))
		result.WriteString(", ")
	}
	result.WriteString(wrapWithBackticks(columns[len(columns)-1]))

	p.selectColumn = result.String()

	return p
}

func (p *QuerySetImpl) GetOrderBySQL() string {
	if p.orderBySQL == "" {
		return ""
	}
	return " ORDER BY " + p.orderBySQL
}

func (p *QuerySetImpl) OrderByToSQL(orderBy any) QuerySet {
	p.setCalled(qsOrderBy)

	switch o := orderBy.(type) {
	case string:
		p.StrOrderByToSQL(o)
	case []string:
		p.SliceOrderByToSQL(o)
	default:
		p.setError(paramTypeError)
		return p
	}

	return p
}

func (p *QuerySetImpl) StrOrderByToSQL(orderBy string) QuerySet {
	p.setCalled(qsOrderBy)

	// If the orderBy string is empty, just return
	if orderBy == "" {
		return p
	}

	p.orderBySQL = orderBy

	return p
}

func (p *QuerySetImpl) SliceOrderByToSQL(orderBy []string) QuerySet {
	p.setCalled(qsOrderBy)

	orderByList := orderBy

	if len(orderByList) == 0 {
		return p
	}

	for _, by := range orderByList {
		by = strings.TrimSpace(by)
		switch strings.HasPrefix(by, descPrefix) {
		case true:
			p.orderBySQL += "`" + by[1:] + "` DESC"
		case false:
			p.orderBySQL += "`" + by + "` ASC"
		}
		p.orderBySQL += ", "
	}

	p.orderBySQL = p.orderBySQL[:len(p.orderBySQL)-2]

	return p
}

func (p *QuerySetImpl) GetLimitSQL() string {
	return p.limitSQL
}

func (p *QuerySetImpl) LimitToSQL(pageSize, pageNum int64) QuerySet {
	p.setCalled(qsLimit)

	if pageSize > 0 && pageNum > 0 {
		var offset, limit int64
		offset = (pageNum - 1) * pageSize
		limit = pageSize
		p.limitSQL = " LIMIT " + strconv.FormatInt(limit, 10) + " OFFSET " + strconv.FormatInt(offset, 10)
	} else if pageSize < 0 || pageNum < 0 {
		p.setError(pageSizeORNumberError)
		return p
	}

	return p
}

func (p *QuerySetImpl) GetGroupBySQL() string {
	if p.groupSQL == "" {
		return ""
	}

	return " GROUP BY " + p.groupSQL
}

func (p *QuerySetImpl) GroupByToSQL(groupBy any) QuerySet {
	switch v := groupBy.(type) {
	case string:
		p.StrGroupByToSQL(v)
	case []string:
		p.SliceGroupByToSQL(v)
	default:
		p.setError(paramTypeError)
	}
	return p
}

func (p *QuerySetImpl) StrGroupByToSQL(groupBy string) QuerySet {
	p.setCalled(qsGroupBy)

	p.groupSQL = groupBy

	return p
}

func (p *QuerySetImpl) SliceGroupByToSQL(groupBy []string) QuerySet {
	p.setCalled(qsGroupBy)

	if len(groupBy) == 0 {
		return p
	}

	var b strings.Builder
	b.WriteString("`")
	b.WriteString(strings.TrimSpace(groupBy[0]))
	for _, by := range groupBy[1:] {
		b.WriteString("`, `")
		b.WriteString(strings.TrimSpace(by))
	}
	b.WriteString("`")

	p.groupSQL = b.String()

	return p
}

func (p *QuerySetImpl) GetHavingSQL() (string, []any) {
	if p.havingSQL.SQL == "" {
		return "", []any{}
	}

	return " HAVING " + p.havingSQL.SQL, p.havingSQL.Args
}

func (p *QuerySetImpl) HavingToSQL(having string, args ...any) QuerySet {
	p.setCalled(qsHaving)

	p.havingSQL.SetSQL(having, args)

	return p
}
