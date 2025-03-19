package norm

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/leisurelicht/norm/operator"
)

var (
	ErrDuplicateKey = operator.ErrDuplicateKey
	ErrNotFound     = operator.ErrNotFound
)

type Operator operator.Operator

const (
	SortKey = "~sort~"
)

type (
	Cond map[string]any
	AND  map[string]any
	OR   map[string]any
)

func ToOR(key string) string {
	if key != "" && key != SortKey && !strings.HasPrefix(key, orPrefix) {
		return orPrefix + strings.TrimSpace(key)
	}
	return key
}

func EachOR(conds any) any {
	switch ctype := conds.(type) {
	case Cond:
		conditions := ctype
		for k, v := range conditions {
			if k == SortKey {
				continue
			}
			if !strings.HasPrefix(k, orPrefix) {
				delete(conditions, k)
				conditions[orPrefix+k] = v
			}
		}
		return conditions
	case AND:
		conditions := ctype
		for k, v := range conditions {
			if k == SortKey {
				continue
			}
			if !strings.HasPrefix(k, orPrefix) {
				delete(conditions, k)
				conditions[orPrefix+k] = v
			}
		}
		return conditions
	case OR:
		conditions := ctype
		for k, v := range conditions {
			if k == SortKey {
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
