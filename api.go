// filters package exposes the core API to build filters.
package filters

import "github.com/XPD-Kasun/filters/op"

// Returns a [filters.Filter] that checks the given field is equal to the value.
func Eq(field string, value any) Filter {
	f := CompareFilter{Op: op.Eq, Field: field, Value: value}
	return f
}

// Returns a [filters.Filter] that checks the given field is not equal to the value.
func Neq(field string, value any) Filter {
	f := CompareFilter{Op: op.Nq, Field: field, Value: value}
	return f
}

// Returns a [filters.Filter] that checks the given field is greater than the value.
func Gt(field string, value any) Filter {
	f := CompareFilter{Op: op.Gt, Field: field, Value: value}
	return f
}

// Returns a [filters.Filter] that checks the given field is greater than or eq the value.
func Gte(field string, value any) Filter {
	f := CompareFilter{Op: op.Gte, Field: field, Value: value}
	return f
}

// Returns a [filters.Filter] that checks the given field is less than the value.
func Lt(field string, value any) Filter {
	f := CompareFilter{Op: op.Lt, Field: field, Value: value}
	return f
}

// Returns a [filters.Filter] that is annotated with given metadata map.
func WithMeta(meta map[string]string, filter Filter) MetaFilter {
	return &MetaFilterImpl{
		meta:   meta,
		filter: filter,
	}
}

// Returns a [filters.Filter] that checks the given field is less than or equal the value.
func Lte(field string, value any) Filter {
	f := CompareFilter{Op: op.Lte, Field: field, Value: value}
	return f
}

// Returns a [filters.Filter] that combines the given filters using logical AND operator.
func And(filters ...Filter) Filter {
	realFilters := []Filter{}
	for _, fl := range filters {
		if IsEmpty(fl) {
			continue
		}
		realFilters = append(realFilters, fl)
	}

	if len(realFilters) == 0 {
		return EmptyFilter{}
	}

	f := LogicalFilter{Op: op.And, Children: realFilters}
	return f
}

// Returns a [filters.Filter] that combines the given filters using logical OR operator.
func Or(filters ...Filter) Filter {

	realFilters := []Filter{}
	for _, fl := range filters {
		if IsEmpty(fl) {
			continue
		}
		realFilters = append(realFilters, fl)
	}

	if len(realFilters) == 0 {
		return EmptyFilter{}
	}

	f := LogicalFilter{Op: op.Or, Children: realFilters}
	return f
}

// Returns a [filters.Filter] that evaluates the NOT operation for the given filter.
func Not(expr Filter) Filter {
	if IsEmpty(expr) {
		return EmptyFilter{}
	}
	f := NotFilter{Child: expr}
	return f
}

// Returns a [filters.Filter] that checks the given field satisfies the value using
// given method. Method can have following:
//  - "" - sql like wildcard matching.
//  - "regex" - regular expression matching.
func Like(field string, value string, method ...string) Filter {
	var m string = ""
	if len(method) > 0 {
		m = method[0]
	}
	f := LikeFilter{Field: field, Value: value, Method: m}
	return f
}

// Returns a [filters.Filter] that checks the given field is equal to
// one of the values passed into args parameter.
func In(field string, args ...any) Filter {
	f := InFilter{
		Field:  field,
		Values: args,
	}
	return f
}

// Returns a [filters.Filter] that checks the given field is within
// the range specified by from and to.
func Between(field string, from, to any) Filter {
	f := BetweenFilter{Field: field, From: from, To: to}
	return f
}

// Reports whether the given filter is either nil or EmptyFilter.
func IsEmpty(filter Filter) bool {
	if filter == nil {
		return true
	}
	_, isEmpty := filter.(EmptyFilter)
	return isEmpty
}
