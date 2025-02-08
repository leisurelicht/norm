package norm

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

const (
	defaultOuterFilterCondsLen = 10
	defaultInnerFilterCondsLen = 10
	orPrefix                   = "| "
	notPrefix                  = "not_"
	descPrefix                 = "-"
	operatorJoiner             = "__"
	plural                     = "~"
)

const (
	_exact       = "exact"
	_exclude     = "exclude"
	_iexact      = "iexact"
	_gt          = "gt"
	_gte         = "gte"
	_lt          = "lt"
	_lte         = "lte"
	_len         = "len"
	_in          = "in"
	_between     = "between"
	_contains    = "contains"
	_icontains   = "icontains"
	_startswith  = "startswith"
	_istartswith = "istartswith"
	_endswith    = "endswith"
	_iendswith   = "iendswith"
)

const (
	argsLenError       = "args length must be equal to ? number"
	orderKeyTypeError  = "order key value must be a list of string"
	orderKeyLenError   = "order key length must be equal to filter key length"
	isNotValueError    = "isNot value must be 0 or 1"
	paramTypeError     = "param type must be string or slice of string"
	filterOrWhereError = "filter or where can not be called at the same time"

	fieldLookupError            = "field lookups [%s] is invalid"
	unknownOperatorError        = "unknown operator [%s]"
	notImplementedOperatorError = "not implemented operator [%s]"
	unsupportedValueError       = "operator [%s] unsupported value type [%s]"
	operatorValueLenError       = "operator [%s] value length must be [%d]"
	operatorValueLenLessError   = "operator [%s] value length must greater than [%d]"
	operatorValueTypeError      = "operator [%s] value must be string list"
	unsupportedFilterTypeError  = "unsupported filter type [%s], Please use be [Cond | AND | OR]"
)

const (
	emptyTag, notTag                   = 0, 1
	andTag, orTag, andNotTag, orNotTag = 0, 1, 2, 3
)

var (
	not          = [2]string{"", " NOT"}
	conjunctions = [4]string{"AND", "OR", "AND NOT", "OR NOT"}
)

type cond struct {
	Conj string
	SQL  string
	Args []any
}

type callFlag int64

const (
	callFilter callFlag = 1 << iota
	callExclude
	callWhere
	callOrderBy
	callLimit
	callSelect
	callGroupBy
)

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
	p.err = nil
}

func (p *QuerySetImpl) GetQuerySet() (sql string, args []any) {
	where := " WHERE "

	if p.whereCond.SQL != "" {
		return where + p.whereCond.SQL, p.whereCond.Args
	}

	if len(p.filterConds) == 0 {
		return "", []any{}
	}

	conj := ""
	for i, filterList := range p.filterConds {
		if i > 0 && len(p.filterConds) > 1 {
			conj = conjunctions[p.filterConjTag[i]]
		} else {
			conj = ""
		}

		if len(filterList) == 1 {
			tempConj := filterList[0].Conj
			if i > 0 {
				tempConj = " " + tempConj
			}
			sql += tempConj + " (" + filterList[0].SQL + ")"
			args = append(args, filterList[0].Args...)
		} else if len(filterList) > 1 {
			sql += "(" + filterList[0].SQL + ")"
			args = append(args, filterList[0].Args...)

			for _, filter := range filterList[1:] {
				sql += " " + filter.Conj + " (" + filter.SQL + ")"
				args = append(args, filter.Args...)
			}
			sql = conj + "(" + sql + ")"
		}
	}

	if sql[0:3] == "AND" {
		sql = sql[4:]
	} else if sql[0:2] == "OR" {
		sql = sql[3:]
	}

	return where + sql, args
}

func (p *QuerySetImpl) filterHandler(filter map[string]any) (filterSql string, filterArgs []any) {
	if len(filter) == 0 {
		return
	}

	var (
		filterConds = make(map[string]*cond)
		skList      []string
		isOrder     = false
		fieldName   string
		operator    string
		andOrFlag   = andTag
		notFlag     = emptyTag
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

		lookup := strings.Split(fieldLookup, operatorJoiner)
		if len(lookup) == 1 {
			operator = _exact
			notFlag = emptyTag
		} else if len(lookup) == 2 {
			operator = strings.ToLower(strings.TrimSpace(lookup[1]))
			if strings.HasPrefix(operator, notPrefix) {
				operator = strings.TrimPrefix(operator, notPrefix)
				notFlag = notTag
			} else {
				notFlag = emptyTag
			}
		} else {
			p.setError(fieldLookupError, fieldLookup)
			return
		}

		op := p.OperatorSQL(operator)
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
		case _exact, _exclude, _iexact:
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
				for i := 0; i < valueOf.Len(); i++ {
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
			sql := fmt.Sprintf(op, fieldName, not[notFlag]) + " (?" + strings.Repeat(",?", valueOf.Len()-1) + ")"
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

func (p *QuerySetImpl) FilterToSQL(isNot int, filter ...any) QuerySet {
	if !p.hasCalled(callWhere) {
		p.setCalled(callFilter)
	} else {
		p.setError(filterOrWhereError)
		return p
	}

	if len(filter) == 0 {
		return p
	}
	if isNot != 0 && isNot != 1 {
		p.setError(isNotValueError)
		return p
	}

	if conj, ok := filter[0].(string); !ok {
		p.filterConjTag = append(p.filterConjTag, andTag)
	} else {
		if res := indexConjunctions(conj); res > 0 {
			p.filterConjTag = append(p.filterConjTag, res)
		} else {
			p.filterConjTag = append(p.filterConjTag, andTag)
		}
		filter = filter[1:]
	}

	var (
		arg         map[string]any
		conjFlag    = andTag
		filterConds = make([]cond, 0, defaultInnerFilterCondsLen)
	)

	for _, f := range filter {
		switch f.(type) {
		case Cond:
			arg, conjFlag = f.(Cond), andTag
		case AND:
			arg, conjFlag = f.(AND), andTag
		case OR:
			arg, conjFlag = f.(OR), orTag
		default:
			p.setError(unsupportedFilterTypeError, reflect.TypeOf(f).String())
		}
		if filterSQL, filterArgs := p.filterHandler(arg); filterSQL == "" {
			continue
		} else {
			filterConds = append(filterConds, *newCondByValue(conjunctions[conjFlag]+not[isNot], filterSQL, filterArgs))
		}
	}

	if len(filterConds) > 0 {
		p.filterConds = append(p.filterConds, filterConds)
	}

	return p
}

func (p *QuerySetImpl) WhereToSQL(cond string, args ...any) QuerySet {
	if !p.hasCalled(callFilter) {
		p.setCalled(callWhere)
	} else {
		p.setError(filterOrWhereError)
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
	p.setCalled(callSelect)

	switch columns.(type) {
	case string:
		p.StrSelectToSQL(columns.(string))
	case []string:
		p.SliceSelectToSQL(columns.([]string))
	default:
		p.setError(paramTypeError)
	}

	return p
}

func (p *QuerySetImpl) StrSelectToSQL(columns string) QuerySet {
	p.setCalled(callSelect)

	p.selectColumn = columns
	return p
}

func (p *QuerySetImpl) SliceSelectToSQL(columns []string) QuerySet {
	p.setCalled(callSelect)

	if len(columns) == 0 {
		return p
	}

	p.selectColumn = processSQL(columns, p.IsSelectKey)
	return p
}

func (p *QuerySetImpl) GetOrderBySQL() string {
	if p.orderBySQL == "" {
		return ""
	}
	return " ORDER BY " + p.orderBySQL
}

func (p *QuerySetImpl) OrderByToSQL(orderBy any) QuerySet {
	p.setCalled(callOrderBy)

	switch orderBy.(type) {
	case string:
		p.StrOrderByToSQL(orderBy.(string))
	case []string:
		p.SliceOrderByToSQL(orderBy.([]string))
	default:
		p.setError(paramTypeError)
		return p
	}

	return p
}

func (p *QuerySetImpl) StrOrderByToSQL(orderBy string) QuerySet {
	p.setCalled(callOrderBy)

	p.orderBySQL = orderBy

	return p
}

func (p *QuerySetImpl) SliceOrderByToSQL(orderBy []string) QuerySet {
	p.setCalled(callOrderBy)

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
	p.setCalled(callLimit)

	if pageSize > 0 && pageNum > 0 {
		var offset, limit int64
		offset = (pageNum - 1) * pageSize
		limit = pageSize
		p.limitSQL = " LIMIT " + strconv.FormatInt(limit, 10) + " OFFSET " + strconv.FormatInt(offset, 10)
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
	switch groupBy.(type) {
	case string:
		p.StrGroupByToSQL(groupBy.(string))
	case []string:
		p.SliceGroupByToSQL(groupBy.([]string))
	default:
		p.setError(paramTypeError)
	}
	return p
}

func (p *QuerySetImpl) StrGroupByToSQL(groupBy string) QuerySet {
	p.setCalled(callGroupBy)

	p.groupSQL = groupBy

	return p
}

func (p *QuerySetImpl) SliceGroupByToSQL(groupBy []string) QuerySet {
	p.setCalled(callGroupBy)

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
