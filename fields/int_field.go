package fields

import "github.com/XPD-Kasun/filters"

type IntField[U any] struct {
	ComparebleField[U, int]
}

func (iF IntField[U]) BetweenFloats(val1, val2 float64) filters.Filter {
	return filters.Between(iF.getAliasName(), val1, val2)
}

func NewIntField[U any](name, alias string) IntField[U] {
	return IntField[U]{
		ComparebleField: ComparebleField[U, int]{
			Name:  name,
			Alias: alias,
		},
	}
}
