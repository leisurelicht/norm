package operator

import (
	"errors"
)

var (
	ErrDuplicateKey = errors.New("duplicate key")
	ErrNotFound     = errors.New("not found")
)

const (
	placeholder = "?"
)

type AddOptions struct {
	TableName   string
	Placeholder string
	DBTag       string
}

type AddFunc func(params *AddOptions)

func DefaultAddOptions(dbTag string) AddOptions {
	return AddOptions{
		Placeholder: placeholder,
		DBTag:       dbTag,
	}
}

func WithTableName(tableName string) AddFunc {
	return func(op *AddOptions) {
		op.TableName = tableName
	}
}

func WithPlaceholder(placeholder string) AddFunc {
	return func(op *AddOptions) {
		op.Placeholder = placeholder
	}
}

func WithDBTag(dbTag string) AddFunc {
	return func(op *AddOptions) {
		op.DBTag = dbTag
	}
}
