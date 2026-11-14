package fields

import "github.com/XPD-Kasun/filters"

type FloatField[U any] struct {
	ComparebleField[U, float64]
}

func (ff FloatField[U]) BetweenInts(val1, val2 int) filters.Filter {
	return filters.Between(ff.getAliasName(), val1, val2)
}

func NewFloatField[U any](name, alias string) FloatField[U] {
	return FloatField[U]{
		ComparebleField: ComparebleField[U, float64]{
			Name:  name,
			Alias: alias,
		},
	}
}
