package norm

import (
	"github.com/leisurelicht/norm/internal/config"
	"github.com/leisurelicht/norm/internal/function"
	"github.com/leisurelicht/norm/internal/operator"
	"github.com/leisurelicht/norm/internal/queryset"
)

type (
	Operator operator.Operator
)

var (
	ErrDuplicateKey = operator.ErrDuplicateKey
	ErrNotFound     = operator.ErrNotFound
)

const (
	SortKey = queryset.SortKey
)

type (
	Cond = queryset.Cond
	AND  = queryset.AND
	OR   = queryset.OR
)

var (
	ToOR = function.ToOR
)

// 泛型包装：在本包重新导出 EachOR
func EachOR[T Cond | AND | OR](conditions T) T {
	return function.EachOR[T](conditions)
}

var Struct2Map = modelStruct2Map

const (
	Debug = config.Debug
	Info  = config.Info
	Warn  = config.Warn
	Error = config.Error
)

func SetLevel(level config.Level) {
	config.SetLevel(level)
}
