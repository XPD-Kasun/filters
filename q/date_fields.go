package q

import (
	"time"

	"github.com/XPD-Kasun/filters"
)

type DateField[U any] struct {
	Name string
}

func (iF DateField[U]) Eq(val time.Time) filters.Filter {
	return filters.EQ(iF.Name, val)
}

func (iF DateField[U]) Gt(val time.Time) filters.Filter {
	return filters.GT(iF.Name, val)
}

func (iF DateField[U]) Gte(val time.Time) filters.Filter {
	return filters.GTE(iF.Name, val)
}

func (iF DateField[U]) Lt(val time.Time) filters.Filter {
	return filters.LT(iF.Name, val)
}

func (iF DateField[U]) Lte(val time.Time) filters.Filter {
	return filters.LTE(iF.Name, val)
}

func (iF DateField[U]) Between(val1, val2 time.Time) filters.Filter {
	return filters.BETWEEN(iF.Name, val1, val2)
}
