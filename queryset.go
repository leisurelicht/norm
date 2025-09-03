package norm

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"sync"
)

// String pool for common SQL components to reduce allocations
var (
	stringPool = sync.Pool{
		New: func() interface{} {
			return make(map[string]string, 16) // Pool of string maps
		},
	}
)

// Cached string constants for better performance
var (
	cachedAndConjunction = " " + conjunctions[0] + " "
	cachedOrConjunction  = " " + conjunctions[1] + " "
)

// Optimized function to build repeated operator strings
func buildRepeatedOperatorSQL(op, fieldName string, valueLen int) string {
	if valueLen <= 1 {
		return fmt.Sprintf(op, fieldName)
	}
	
	// Pre-calculate total length needed
	baseSQL := fmt.Sprintf(op, fieldName)
	totalLen := len(baseSQL)*valueLen + len(cachedAndConjunction)*(valueLen-1)
	
	var sqlBuilder strings.Builder
	sqlBuilder.Grow(totalLen)
	sqlBuilder.WriteString(baseSQL)
	
	for i := 1; i < valueLen; i++ {
		sqlBuilder.WriteString(cachedAndConjunction)
		sqlBuilder.WriteString(baseSQL)
	}
	
	return sqlBuilder.String()
}

// Optimized placeholder building
func buildPlaceholders(placeholder string, count int) string {
	if count <= 1 {
		return placeholder
	}
	
	// Pre-calculate exact capacity needed
	totalLen := len(placeholder)*count + (count-1) // count-1 commas
	
	var builder strings.Builder
	builder.Grow(totalLen)
	builder.WriteString(placeholder)
	
	for i := 1; i < count; i++ {
		builder.WriteByte(',')
		builder.WriteString(placeholder)
	}
	
	return builder.String()
}

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
	
	// Pre-allocated string constants for better performance
	whereClause      = " WHERE "
	leftParen        = "("
	rightParen       = ")"
	spaceAnd         = " AND "
	spaceOr          = " OR "
	spaceAndNot      = " AND NOT "
	spaceOrNot       = " OR NOT "
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
	// Optimized conjunction map with pre-allocated strings
	conjunctionMap = map[reflect.Type]int{
		reflect.TypeOf(Cond{}): andTag,
		reflect.TypeOf(AND{}):  andTag,
		reflect.TypeOf(OR{}):   orTag,
	}
	// Pre-allocated space strings for better performance
	spaceConjunctions = [4]string{" AND ", " OR ", " AND NOT ", " OR NOT "}
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
		return whereClause + p.whereCond.SQL, p.whereCond.Args
	}

	// Early return for no filter conditions
	if len(p.filterConds) == 0 {
		return "", nil
	}

	// Pre-calculate more accurate capacity for string builder and args slice
	totalConditions := 0
	maxSQLLength := len(whereClause) // Start with WHERE clause length
	
	for _, filterList := range p.filterConds {
		totalConditions += len(filterList)
		// Estimate SQL length based on existing conditions
		for _, cond := range filterList {
			maxSQLLength += len(cond.SQL) + 20 // 20 chars for conjunctions and parens
		}
	}

	// Build SQL with pre-allocated builder
	outerSQL := strings.Builder{}
	outerSQL.Grow(maxSQLLength)
	outerSQL.WriteString(whereClause)

	// Pre-allocate args slice with accurate capacity
	args = make([]any, 0, totalConditions*2)

	for i, condList := range p.filterConds {
		// Add conjunction between filter groups
		if i > 0 {
			outerSQL.WriteString(spaceConjunctions[p.filterConjTag[i]])
		}

		// Check if this is a NOT condition by examining the first filter in the group
		isNot := strings.Contains(condList[0].Conj, "NOT")
		if isNot {
			outerSQL.WriteString("NOT ")
		}

		// Single condition doesn't need inner parentheses
		if len(condList) == 1 {
			outerSQL.WriteString(leftParen)
			outerSQL.WriteString(condList[0].SQL)
			outerSQL.WriteString(rightParen)
			args = append(args, condList[0].Args...)
			continue
		}

		// Multiple conditions in this group
		outerSQL.WriteString(leftParen)

		// First condition
		outerSQL.WriteString(leftParen)
		outerSQL.WriteString(condList[0].SQL)
		outerSQL.WriteString(rightParen)
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
			outerSQL.WriteString(" ")
			outerSQL.WriteString(leftParen)
			outerSQL.WriteString(filter.SQL)
			outerSQL.WriteString(rightParen)
			args = append(args, filter.Args...)
		}

		outerSQL.WriteString(rightParen)
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
		// Reset flags for each iteration
		andOrFlag = andTag
		method = ""
		
		if strings.HasPrefix(fieldLookup, orPrefix) {
			fieldLookup = fieldLookup[len(orPrefix):] // More efficient than TrimPrefix
			andOrFlag = orTag
		}

		if methodPos := strings.Index(fieldLookup, methodJoiner); methodPos != -1 {
			method = fieldLookup[methodPos+len(methodJoiner):]
			fieldLookup = fieldLookup[:methodPos]
		}

		// Optimize lookup parsing with fewer allocations
		if operatorPos := strings.Index(fieldLookup, operatorJoiner); operatorPos == -1 {
			operator = _exact
			notFlag = notNot
			fieldName = fieldLookup
		} else {
			parts := fieldLookup[operatorPos+len(operatorJoiner):]
			if strings.HasPrefix(parts, notPrefix) {
				operator = parts[len(notPrefix):]
				notFlag = isNot
			} else {
				operator = parts
				notFlag = notNot
			}
			fieldName = fieldLookup[:operatorPos]
		}

		op := p.OperatorSQL(operator, method)
		if op == "" {
			p.setError(unknownOperatorError, operator)
			return
		}

		fCond := newCond()
		fCond.SetConj(conjunctions[andOrFlag])

		// Cache reflection to reduce overhead
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
				valueLen := valueOf.Len()
				if valueLen == 0 {
					p.setError(operatorValueLenLessError, operator, 0)
					return
				}
				
				// Pre-allocate args slice for better performance
				args := make([]any, valueLen)
				for i := 0; i < valueLen; i++ {
					args[i] = valueOf.Index(i).Interface()
				}
				
				// Use optimized function to build repeated SQL
				sql := buildRepeatedOperatorSQL(op, fieldName, valueLen)
				if len(filter) > 1 {
					sql = leftParen + sql + rightParen
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
				sql := fmt.Sprintf(op, fieldName, not[notFlag]) + leftParen + filedValue.(string) + rightParen
				filterConds[fieldName] = fCond.SetSQL(sql, []any{})
				continue
			}

			if !isListKind(valueKind) {
				p.setError(unsupportedValueError, operator, valueKind.String())
				return
			}
			valueLen := valueOf.Len()
			if valueLen == 0 {
				p.setError(operatorValueLenLessError, operator, 0)
				return
			}
			
			// Use optimized placeholder building
			placeholders := buildPlaceholders(p.Placeholder(), valueLen)
			sql := fmt.Sprintf(op, fieldName, not[notFlag]) + leftParen + placeholders + rightParen
			
			args := make([]any, valueLen)
			for i := 0; i < valueLen; i++ {
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

	// More accurate capacity estimation
	totalLen := 0
	for _, col := range columns {
		totalLen += len(col) + 4 // Add 4 for backticks and comma+space
	}

	var result strings.Builder
	result.Grow(totalLen)
	
	// First column
	result.WriteString(wrapWithBackticks(columns[0]))
	
	// Remaining columns
	for i := 1; i < len(columns); i++ {
		result.WriteString(", ")
		result.WriteString(wrapWithBackticks(columns[i]))
	}

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

	if len(orderBy) == 0 {
		return p
	}

	// Estimate capacity more accurately
	estimatedLen := 0
	for _, by := range orderBy {
		estimatedLen += len(by) + 10 // 10 chars for backticks, DESC/ASC, comma, space
	}

	var result strings.Builder
	result.Grow(estimatedLen)

	for i, by := range orderBy {
		if i > 0 {
			result.WriteString(", ")
		}
		
		by = strings.TrimSpace(by)
		if strings.HasPrefix(by, descPrefix) {
			result.WriteByte('`')
			result.WriteString(by[1:])
			result.WriteString("` DESC")
		} else {
			result.WriteByte('`')
			result.WriteString(by)
			result.WriteString("` ASC")
		}
	}

	p.orderBySQL = result.String()
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

	// Estimate capacity: each field needs backticks, trimming, commas
	estimatedLen := 0
	for _, by := range groupBy {
		estimatedLen += len(strings.TrimSpace(by)) + 4 // 4 for backticks and comma+space
	}

	var b strings.Builder
	b.Grow(estimatedLen)
	
	// First field
	b.WriteByte('`')
	b.WriteString(strings.TrimSpace(groupBy[0]))
	b.WriteByte('`')
	
	// Remaining fields
	for _, by := range groupBy[1:] {
		b.WriteString("`, `")
		b.WriteString(strings.TrimSpace(by))
	}
	b.WriteByte('`')

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
