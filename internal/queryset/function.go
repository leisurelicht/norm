package queryset

import (
	"strings"
)

func ToOR(key string) string {
	if key != "" && key != SortKey && !strings.HasPrefix(key, OrPrefix) {
		return OrPrefix + strings.TrimSpace(key)
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
		if !strings.HasPrefix(k, OrPrefix) {
			delete(conditions, k)
			conditions[OrPrefix+k] = v
		}
	}
	return conditions
}
