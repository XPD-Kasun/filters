package fields

import "github.com/XPD-Kasun/filters"

type BoolField[U any] struct {
	Name  string
	Alias string
}

func (bf BoolField[U]) getAliasName() string {
	if bf.Alias != "" {
		return bf.Alias + "." + bf.Name
	}
	return bf.Name
}

func (bf BoolField[U]) Eq(val U) filters.Filter {
	return filters.Eq(bf.getAliasName(), val)
}

func (bf BoolField[U]) Nq(val U) filters.Filter {
	return filters.Nq(bf.getAliasName(), val)
}

func NewBoolField[U any](name, alias string) BoolField[U] {
	return BoolField[U]{
		Name:  name,
		Alias: alias,
	}
}
