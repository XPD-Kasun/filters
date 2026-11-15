package fields

import "github.com/XPD-Kasun/filters"

type StringField[U any] struct {
	Orderable[string, U]
}

func (sf StringField[U]) Like(pattern string) filters.Filter {
	return filters.Like(sf.getAliasName(), pattern, "")
}

func (sf StringField[U]) Regex(pattern string) filters.Filter {
	return filters.Like(sf.getAliasName(), pattern, "regex")
}

func NewStringField[U any](name, alias string) StringField[U] {
	return StringField[U]{
		Orderable: Orderable[string, U]{
			Equalable: Equalable[string, U]{
				Name:  name,
				Alias: alias,
			},
		},
	}
}

type NullableStringField[U any] struct {
	StringField[U]
}

func (ns NullableStringField[U]) IsNull() filters.Filter {
	return filters.Eq(ns.getAliasName(), nil)
}

func (ns NullableStringField[U]) IsNotNull() filters.Filter {
	return filters.Nq(ns.getAliasName(), nil)
}

func NewNullableStringField[U any](name, alias string) NullableStringField[U] {
	return NullableStringField[U]{
		StringField: NewStringField[U](name, alias),
	}
}
