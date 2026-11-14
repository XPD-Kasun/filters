package fields

import "github.com/XPD-Kasun/filters"

type StringField[U any] struct {
	ComparebleField[U, string]
}

func (sf StringField[U]) Like(pattern string) filters.Filter {
	return filters.Like(sf.getAliasName(), pattern, "")
}

func (sf StringField[U]) Regex(pattern string) filters.Filter {
	return filters.Like(sf.getAliasName(), pattern, "regex")
}

func NewStringField[U any](name, alias string) StringField[U] {
	return StringField[U]{
		ComparebleField: ComparebleField[U, string]{
			Name:  name,
			Alias: alias,
		},
	}
}
