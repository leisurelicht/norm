package norm

import (
	"database/sql"
	"errors"
	"fmt"
	"reflect"
	"strings"
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

func rawFieldNames(in any, tag string, pg bool) []string {
	if in == nil {
		panic(errors.New("model is nil"))
	}

	v := reflect.ValueOf(in)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	// we only accept structs
	if v.Kind() != reflect.Struct {
		panic(fmt.Errorf("model only can be a struct; got %s", v.Kind()))
	}

	out := make([]string, 0, v.NumField())

	typ := v.Type()
	for i := range v.NumField() {
		// gets us a StructField
		fi := typ.Field(i)
		tagv := fi.Tag.Get(tag)
		switch tagv {
		case "-":
			continue
		case "":
			if pg {
				out = append(out, fi.Name)
			} else {
				out = append(out, fmt.Sprintf("`%s`", fi.Name))
			}
		default:
			// get tag name with the tag option, e.g.:
			// `db:"id"`
			// `db:"id,type=char,length=16"`
			// `db:",type=char,length=16"`
			// `db:"-,type=char,length=16"`
			if strings.Contains(tagv, ",") {
				tagv = strings.TrimSpace(strings.Split(tagv, ",")[0])
			}
			if tagv == "-" {
				continue
			}
			if len(tagv) == 0 {
				tagv = fi.Name
			}
			if pg {
				out = append(out, tagv)
			} else {
				out = append(out, fmt.Sprintf("`%s`", tagv))
			}
		}
	}

	return out
}

// strSlice2Map convert string slice to map
func strSlice2Map(s []string) (res map[string]struct{}) {
	res = make(map[string]struct{})
	for _, e := range s {
		res[e] = struct{}{}
	}
	return res
}

// modelStruct2Map convert struct to map, tag is the tag name of struct
func modelStruct2Map(obj any, tag string) map[string]any {
	t := reflect.TypeOf(obj)
	v := reflect.ValueOf(obj)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
		v = v.Elem()
	}

	data := make(map[string]any, t.NumField())

	for i := 0; i < t.NumField(); i++ {
		if t.Field(i).Tag.Get(tag) == "" || t.Field(i).Tag.Get(tag) == "-" {
			// skip empty tag
			// skip "-" tag
			continue
		}

		if v.Field(i).Kind() != reflect.Struct {
			data[t.Field(i).Tag.Get(tag)] = v.Field(i).Interface()
		}

		value := v.Field(i).Interface()
		switch reflect.TypeOf(value) {
		case reflect.TypeOf(sql.NullByte{}):
			if value.(sql.NullByte).Valid {
				value = value.(sql.NullByte).Byte
			} else {
				value = nil
			}
		case reflect.TypeOf(sql.NullBool{}):
			if value.(sql.NullBool).Valid {
				value = value.(sql.NullBool).Bool
			} else {
				value = nil
			}
		case reflect.TypeOf(sql.NullFloat64{}):
			if value.(sql.NullFloat64).Valid {
				value = value.(sql.NullFloat64).Float64
			} else {
				value = nil
			}
		case reflect.TypeOf(sql.NullInt16{}):
			if value.(sql.NullInt16).Valid {
				value = value.(sql.NullInt16).Int16
			} else {
				value = nil
			}
		case reflect.TypeOf(sql.NullInt32{}):
			if value.(sql.NullInt32).Valid {
				value = value.(sql.NullInt32).Int32
			} else {
				value = nil
			}
		case reflect.TypeOf(sql.NullInt64{}):
			if value.(sql.NullInt64).Valid {
				value = value.(sql.NullInt64).Int64
			} else {
				value = nil
			}
		case reflect.TypeOf(sql.NullString{}):
			if value.(sql.NullString).Valid {
				value = value.(sql.NullString).String
			} else {
				value = nil
			}
		case reflect.TypeOf(sql.NullTime{}):
			if value.(sql.NullTime).Valid {
				value = value.(sql.NullTime).Time
			} else {
				value = nil
			}
		}
		data[t.Field(i).Tag.Get(tag)] = value
	}
	return data
}

// modelStructSlice2MapSlice convert struct slice to map slice, tag is the tag name of struct
func modelStructSlice2MapSlice(obj any, tag string) []map[string]any {
	t := reflect.TypeOf(obj)
	v := reflect.ValueOf(obj)
	if t.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	data := make([]map[string]any, 0, v.Len())

	for i := 0; i < v.Len(); i++ {
		data = append(data, modelStruct2Map(v.Index(i).Interface(), tag))
	}
	return data
}

func createModelPointerAndSlice(input any) (any, any) {
	inputType := reflect.ValueOf(input).Type()

	inputPointer := reflect.New(inputType).Interface()

	slicePointer := reflect.New(reflect.SliceOf(inputType)).Interface()

	return inputPointer, slicePointer
}

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
	op := p.OperatorSQL(operator)

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

// deepCopyModelPtrStructure deep copy a pointer struct model or a pointer slice model.
//
// The input must be a pointer to a struct or a pointer to a slice of structs, if not, return nil.
// I didn't check src type whether it is struct because it should be generated by the createModelPointerAndSlice function.
// So if this function return nil, check createModelPointerAndSlice first, make sure it can pass its unit test.
func deepCopyModelPtrStructure(src any) any {
	srcValue := reflect.ValueOf(src)

	if srcValue.Kind() != reflect.Ptr {
		return src
	}

	// create a new value of the same type as src
	return reflect.New(srcValue.Elem().Type()).Interface()
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

func processSQL(sqlParts []string, isKeyWord func(word string) bool) string {
	var result strings.Builder

	// 遍历每个子字符串
	for i, part := range sqlParts {
		// 分割子字符串为单词
		words := strings.Fields(part)

		// 遍历每个单词
		for j, word := range words {
			// 如果是 SQL 关键字，直接输出，不包裹反引号
			if isKeyWord(word) {
				result.WriteString(word)
			} else {
				// 如果不是关键字，包裹反引号
				result.WriteString(wrapWithBackticks(word))
			}

			// 添加空格分隔
			if j < len(words)-1 {
				result.WriteString(" ")
			}
		}

		// 如果不是最后一个部分，添加逗号和空格
		if i < len(sqlParts)-1 {
			result.WriteString(", ")
		}
	}

	return result.String()
}
