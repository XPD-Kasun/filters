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
