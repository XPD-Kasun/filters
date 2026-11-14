package fields

import "github.com/XPD-Kasun/filters"

type ComparebleField[U any, T comparable] struct {
	Name  string
	Alias string
}

func (cf ComparebleField[U, T]) getAliasName() string {
	if cf.Alias != "" {
		return cf.Alias + "." + cf.Name
	}
	return cf.Name
}

func (cf ComparebleField[U, T]) Eq(val T) filters.Filter {
	return filters.Eq(cf.getAliasName(), val)
}

func (cf ComparebleField[U, T]) Nq(val T) filters.Filter {
	return filters.Nq(cf.getAliasName(), val)
}

func (cf ComparebleField[U, T]) Gt(val T) filters.Filter {
	return filters.Gt(cf.getAliasName(), val)
}

func (cf ComparebleField[U, T]) Gte(val T) filters.Filter {
	return filters.Gte(cf.getAliasName(), val)
}

func (cf ComparebleField[U, T]) Lt(val T) filters.Filter {
	return filters.Lt(cf.getAliasName(), val)
}

func (cf ComparebleField[U, T]) Lte(val T) filters.Filter {
	return filters.Lte(cf.getAliasName(), val)
}

func NewComparebleField[U any, T comparable](name, alias string) ComparebleField[U, T] {
	return ComparebleField[U, T]{
		Name:  name,
		Alias: alias,
	}
}
