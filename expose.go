package norm

import (
	"fmt"
	"reflect"
	"strings"
)

const (
	OrderKey = "~order~"
)

type (
	Cond map[string]any
	AND  map[string]any
	OR   map[string]any
)

func ToOR(key string) string {
	if key != "" && key != OrderKey && !strings.HasPrefix(key, orPrefix) {
		return orPrefix + strings.TrimSpace(key)
	}
	return key
}

func EachOR(conds any) any {
	switch conds.(type) {
	case Cond:
		conditions := conds.(Cond)
		for k, v := range conditions {
			if k == OrderKey {
				continue
			}
			if !strings.HasPrefix(k, orPrefix) {
				delete(conditions, k)
				conditions[orPrefix+k] = v
			}
		}
		return conditions
	case AND:
		conditions := conds.(AND)
		for k, v := range conditions {
			if k == OrderKey {
				continue
			}
			if !strings.HasPrefix(k, orPrefix) {
				delete(conditions, k)
				conditions[orPrefix+k] = v
			}
		}
		return conditions
	case OR:
		conditions := conds.(OR)
		for k, v := range conditions {
			if k == OrderKey {
				continue
			}
			if !strings.HasPrefix(k, orPrefix) {
				delete(conditions, k)
				conditions[orPrefix+k] = v
			}
		}
		return conditions
	default:
		fmt.Println(reflect.TypeOf(conds).String())
	}
	return conds
}

type Operator interface {
	OperatorSQL(operator string) string
}
