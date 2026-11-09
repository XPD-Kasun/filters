package q

import "github.com/XPD-Kasun/filters"

type StringField[U any] struct {
	Name string
}

func (sf StringField[U]) Eq(val string) filters.Filter {
	return filters.EQ(sf.Name, val)
}

func (sf StringField[U]) Like(pattern string) filters.Filter {
	return filters.LIKE(sf.Name, pattern, "")
}

func (sf StringField[U]) Regex(pattern string) filters.Filter {
	return filters.LIKE(sf.Name, pattern, "regex")
}
