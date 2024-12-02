package norm

import (
	"database/sql"
	"fmt"
	"reflect"
	"strings"
)

// Struct2Map convert struct to map, tag is the tag name of struct
func Struct2Map(obj any, tag string) map[string]any {
	t := reflect.TypeOf(obj)
	v := reflect.ValueOf(obj)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
		v = v.Elem()
	}

	data := make(map[string]any, t.NumField())

	for i := 0; i < t.NumField(); i++ {
		if t.Field(i).Tag.Get(tag) == "" {
			continue
		}

		if v.Field(i).Kind() != reflect.Struct {
			data[t.Field(i).Tag.Get(tag)] = v.Field(i).Interface()
		}

		value := v.Field(i).Interface()
		switch reflect.TypeOf(value) {
		case reflect.TypeOf(sql.NullString{}):
			if value.(sql.NullString).Valid {
				value = value.(sql.NullString).String
			} else {
				value = nil
			}
		case reflect.TypeOf(sql.NullInt64{}):
			if value.(sql.NullInt64).Valid {
				value = value.(sql.NullInt64).Int64
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

// StructSlice2MapSlice convert struct slice to map slice, tag is the tag name of struct
func StructSlice2MapSlice(obj any, tag string) []map[string]any {
	t := reflect.TypeOf(obj)
	v := reflect.ValueOf(obj)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
		v = v.Elem()
	}

	data := make([]map[string]any, 0, v.Len())

	for i := 0; i < v.Len(); i++ {
		data = append(data, Struct2Map(v.Index(i).Interface(), tag))
	}
	return data
}

func CreatePointerAndSlice(input interface{}) (interface{}, interface{}) {
	// 获取输入值的反射值
	inputValue := reflect.ValueOf(input)

	// 获取输入值的类型
	inputType := inputValue.Type()

	// 创建输入值的指针
	inputPointer := reflect.New(inputType).Interface()
	reflect.ValueOf(inputPointer).Elem().Set(inputValue)

	// 创建一个包含输入值的切片
	sliceType := reflect.SliceOf(inputType)
	slice := reflect.MakeSlice(sliceType, 1, 1)
	slice.Index(0).Set(inputValue)

	// 创建包含该切片的指针
	slicePointer := reflect.New(sliceType).Interface()
	reflect.ValueOf(slicePointer).Elem().Set(slice)

	// 返回输入值的指针和包含输入值的切片指针
	return inputPointer, slicePointer
}

func isStrList(v any) bool {
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

func isNumericKind(kind reflect.Kind) bool {
	return kind == reflect.Int || kind == reflect.Int8 || kind == reflect.Int16 || kind == reflect.Int32 || kind == reflect.Int64 ||
		kind == reflect.Uint || kind == reflect.Uint8 || kind == reflect.Uint16 || kind == reflect.Uint32 || kind == reflect.Uint64 ||
		kind == reflect.Float32 || kind == reflect.Float64
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

func indexConjunctions(conj string) int {
	for i, v := range conjunctions {
		if strings.ToUpper(conj) == v {
			return i
		}
	}
	return -1
}

func genStrListValueLikeSQL(p *QuerySetImpl, filterConditions map[string]*cond, fieldName string, valueOf reflect.Value, notFlag int, operator, valueFormat string) {
	op := p.OperatorSQL(operator)

	filterConditions[fieldName] = newCondByValue("", fmt.Sprintf(op, fieldName, not[notFlag]), []any{fmt.Sprintf(valueFormat, valueOf.Index(0).Interface())})
	for i := 1; i < valueOf.Len(); i++ {
		if valueOf.Index(i).IsZero() {
			p.setError("operator [%s] unsupported value empty ", operator)
			return
		}

		filterConditions[fieldName].SQL += fmt.Sprintf(" "+conjunctions[0]+" "+op, fieldName, not[notFlag])
		filterConditions[fieldName].Args = append(filterConditions[fieldName].Args, fmt.Sprintf(valueFormat, valueOf.Index(i).Interface()))
	}
}

func joinSQL(filterSql *string, filterArgs *[]any, index int, condition *cond) {
	if index == 0 {
		*filterSql += condition.SQL
	} else {
		*filterSql += " " + condition.Conj + " " + condition.SQL
	}
	*filterArgs = append(*filterArgs, condition.Args...)
}

// DeepCopy deep copy
func DeepCopy(src any) (any, error) {
	srcValue := reflect.ValueOf(src)

	// If src is not a pointer or interface, simply return the original value
	if srcValue.Kind() != reflect.Ptr && srcValue.Kind() != reflect.Interface {
		return src, nil
	}

	// If src is a nil pointer or interface, return a nil copy
	if srcValue.IsNil() {
		return nil, nil
	}

	// Insert a new value of the same type as src
	dest := reflect.New(srcValue.Elem().Type()).Interface()

	// If src is a slice, perform a deep copy of the slice elements
	if srcValue.Elem().Kind() == reflect.Slice {
		srcSlice := srcValue.Elem()
		destSlice := reflect.ValueOf(dest).Elem()

		for i := 0; i < srcSlice.Len(); i++ {
			elem, err := DeepCopy(srcSlice.Index(i).Interface())
			if err != nil {
				return nil, err
			}
			destSlice = reflect.Append(destSlice, reflect.ValueOf(elem))
		}
		return dest, nil
	}

	// If src is a struct, perform a deep copy of the struct fields
	if srcValue.Elem().Kind() == reflect.Struct {
		for i := 0; i < srcValue.Elem().NumField(); i++ {
			field := srcValue.Elem().Field(i)
			// Deep copy each struct field
			deepCopyField := reflect.New(field.Type()).Interface()
			DeepCopy(field.Interface())
			reflect.ValueOf(dest).Elem().Field(i).Set(reflect.ValueOf(deepCopyField).Elem())
		}
		return dest, nil
	}

	// For other types, return the original value
	return src, nil
}

// StrSlice2Map convert string slice to map
func StrSlice2Map(s []string) (res map[string]struct{}) {
	res = make(map[string]struct{})
	for _, e := range s {
		res[e] = struct{}{}
	}
	return res
}
