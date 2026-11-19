package fields

import "github.com/XPD-Kasun/filters"

// Describes a non nullable string field in struct
type StringField[U any] struct {
	Orderable[string, U]
}

// Returns a [filters.Filter] that checks whether this field's value
// is similar to the given pattern. pattern format uses `sql_like` semantics.
// Returns a [filters.Filter] that checks whether this field's value
// is similar to the given pattern. pattern format uses sql_like semantics.
func (sf StringField[U]) Like(pattern string) filters.Filter {
	return filters.StringMatch(sf.getAliasName(), pattern, filters.WildcardMatch)
}

// Returns a [filters.Filter] that checks whether this field's value
// contains the given substring.
func (sf StringField[U]) Contains(value string) filters.Filter {
	return filters.StringMatch(sf.getAliasName(), value, filters.Contains)
}

// Returns a [filters.Filter] that checks whether this field's value
// starts with the given prefix.
func (sf StringField[U]) StartsWith(value string) filters.Filter {
	return filters.StringMatch(sf.getAliasName(), value, filters.StartsWith)
}

// Returns a [filters.Filter] that checks whether this field's value
// ends with the given suffix.
func (sf StringField[U]) EndsWith(value string) filters.Filter {
	return filters.StringMatch(sf.getAliasName(), value, filters.EndsWith)
}

// Returns a [filters.Filter] that checks whether this field's value
// is similar to the given regex pattern.
func (sf StringField[U]) Regex(pattern string) filters.Filter {
	return filters.StringMatch(sf.getAliasName(), pattern, filters.Regex)
}

// Creates a new [StringField].
func NewStringField[U any](name, alias string) StringField[U] {
	return StringField[U]{
		Name:  name,
		Alias: alias,
	}
}

// Describes a string field that can be null. That is a *string.
type NullableStringField[U any] struct {
	StringField[U]
}

// Returns a [filters.Filter] that checks whether this field is nil.
func (ns NullableStringField[U]) IsNull() filters.Filter {
	return filters.Eq(ns.getAliasName(), nil)
}

// Returns a [filters.Filter] that checks whether this field is not nil.
func (ns NullableStringField[U]) IsNotNull() filters.Filter {
	return filters.Neq(ns.getAliasName(), nil)
}

// Creates a new [NullableStringField].
func NewNullableStringField[U any](name, alias string) NullableStringField[U] {
	return NullableStringField[U]{
		StringField: NewStringField[U](name, alias),
	}
}
