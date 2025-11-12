package function

import (
	"github.com/leisurelicht/norm/internal/queryset"
	"strings"
)

func ToOR(key string) string {
	if key != "" && key != queryset.SortKey && !strings.HasPrefix(key, queryset.OrPrefix) {
		return queryset.OrPrefix + strings.TrimSpace(key)
	}
	return key
}

// EachOR converts all keys in a condition to OR keys
// Only accepts Cond, AND, OR as input
func EachOR[T queryset.Cond | queryset.AND | queryset.OR](conditions T) T {
	for k, v := range conditions {
		if k == queryset.SortKey {
			continue
		}
		if !strings.HasPrefix(k, queryset.OrPrefix) {
			delete(conditions, k)
			conditions[queryset.OrPrefix+k] = v
		}
	}
	return conditions
}
