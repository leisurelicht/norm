package norm

import (
	"strings"

	"github.com/leisurelicht/norm/internal/operator"
)

var (
	ErrDuplicateKey = operator.ErrDuplicateKey
	ErrNotFound     = operator.ErrNotFound
)

type (
	Operator operator.Operator
)

const (
	SortKey = sortKey
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

// EachOR converts all keys in a condition to OR keys
// Only accepts Cond, AND, OR as input
func EachOR[T Cond | AND | OR](conditions T) T {
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
}

var Struct2Map = modelStruct2Map
