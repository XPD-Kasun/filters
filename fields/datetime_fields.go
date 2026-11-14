package fields

import (
	"time"

	"github.com/XPD-Kasun/filters"
)

type DateTimeField[U any] struct {
	ComparebleField[U, time.Time]
}

func (iF DateTimeField[U]) Between(val1, val2 time.Time) filters.Filter {
	return filters.Between(iF.getAliasName(), val1, val2)
}

func NewDateTimeField[T comparable](name, alias string) DateTimeField[T] {
	return DateTimeField[T]{
		ComparebleField[T, time.Time]{
			Name:  name,
			Alias: alias,
		},
	}
}
