package norm

import (
	"fmt"
	"reflect"
)

func isStrList(v any) bool {
	if v == nil {
		return false
	}

	switch reflect.TypeOf(v).Kind() {
	case reflect.Slice:
		slice, ok := v.([]string)
		return ok && slice != nil
	case reflect.Array:
		value := reflect.ValueOf(v)
		if value.Len() > 0 && value.Index(0).Kind() == reflect.String {
			return true
		}
	default:
		return false
	}
	return false
}

// numericKinds is a set of all reflect.Kind values that represent numeric types
var numericKinds = map[reflect.Kind]struct{}{
	reflect.Int:     {},
	reflect.Int8:    {},
	reflect.Int16:   {},
	reflect.Int32:   {},
	reflect.Int64:   {},
	reflect.Uint:    {},
	reflect.Uint8:   {},
	reflect.Uint16:  {},
	reflect.Uint32:  {},
	reflect.Uint64:  {},
	reflect.Float32: {},
	reflect.Float64: {},
}

func isNumericKind(kind reflect.Kind) bool {
	_, ok := numericKinds[kind]
	return ok
}

func isStringKind(kind reflect.Kind) bool {
	return kind == reflect.String
}

func isBoolKind(kind reflect.Kind) bool {
	return kind == reflect.Bool
}

func isListKind(kind reflect.Kind) bool {
	return kind == reflect.Slice || kind == reflect.Array
}

func genStrListValueLikeSQL(p *QuerySetImpl, filterConditions map[string]*cond, fieldName string, valueOf reflect.Value, notFlag int, operator, valueFormat string) {
	op := p.OperatorSQL(operator, "")

	filterConditions[fieldName] = newCondByValue("", fmt.Sprintf(op, fieldName, not[notFlag]), []any{fmt.Sprintf(valueFormat, valueOf.Index(0).Interface())})
	for i := 1; i < valueOf.Len(); i++ {
		if valueOf.Index(i).IsZero() {
			p.setError(operatorValueEmptyError, operator)
			return
		}

		filterConditions[fieldName].SQL += fmt.Sprintf(" "+conjunctions[notFlag^1]+" "+op, fieldName, not[notFlag])
		filterConditions[fieldName].Args = append(filterConditions[fieldName].Args, fmt.Sprintf(valueFormat, valueOf.Index(i).Interface()))
	}
}

func joinSQL(filterSql *string, filterArgs *[]any, index int, condition *cond) {
	if filterSql == nil || filterArgs == nil || condition == nil {
		return
	}
	if index == 0 {
		*filterSql += condition.SQL
	} else {
		*filterSql += " " + condition.Conj + " " + condition.SQL
	}
	*filterArgs = append(*filterArgs, condition.Args...)
}

// 用反引号包裹字段
func wrapWithBackticks(str string) string {
	if str == "" {
		return str
	}
	// 如果已经被包裹，不做修改
	if len(str) > 1 && str[0] == '`' && str[len(str)-1] == '`' {
		return str
	}
	// 包裹字段
	return "`" + str + "`"
}
