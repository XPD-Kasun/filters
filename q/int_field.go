package q

import "github.com/XPD-Kasun/filters"

type IntField[U any] struct {
	Name string
}

func (iF IntField[U]) Eq(val int) filters.Filter {
	return filters.EQ(iF.Name, val)
}

func (iF IntField[U]) Gt(val int) filters.Filter {
	return filters.GT(iF.Name, val)
}

func (iF IntField[U]) Gte(val int) filters.Filter {
	return filters.GTE(iF.Name, val)
}

func (iF IntField[U]) Lt(val int) filters.Filter {
	return filters.LT(iF.Name, val)
}

func (iF IntField[U]) Lte(val int) filters.Filter {
	return filters.LTE(iF.Name, val)
}

func (iF IntField[U]) Between(val1, val2 int) filters.Filter {
	return filters.BETWEEN(iF.Name, val1, val2)
}

func (iF IntField[U]) BetweenFloats(val1, val2 float64) filters.Filter {
	return filters.BETWEEN(iF.Name, val1, val2)
}
