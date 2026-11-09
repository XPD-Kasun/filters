package q

import "github.com/XPD-Kasun/filters"

type FloatField[U any] struct {
	Name string
}

func (ff FloatField[U]) Eq(val float64) filters.Filter {
	return filters.EQ(ff.Name, val)
}

func (ff FloatField[U]) Gt(val float64) filters.Filter {
	return filters.GT(ff.Name, val)
}

func (ff FloatField[U]) Gte(val float64) filters.Filter {
	return filters.GTE(ff.Name, val)
}

func (ff FloatField[U]) Lt(val float64) filters.Filter {
	return filters.LT(ff.Name, val)
}

func (ff FloatField[U]) Lte(val float64) filters.Filter {
	return filters.LTE(ff.Name, val)
}

func (ff FloatField[U]) Between(val1, val2 float64) filters.Filter {
	return filters.BETWEEN(ff.Name, val1, val2)
}

func (ff FloatField[U]) BetweenInts(val1, val2 int) filters.Filter {
	return filters.BETWEEN(ff.Name, val1, val2)
}
