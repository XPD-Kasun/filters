package filters

import "github.com/XPD-Kasun/filters/op"

func Eq(field string, value any) Filter {
	f := CompareFilter{Op: op.Eq, Field: field, Value: value}
	return f
}

func Nq(field string, value any) Filter {
	f := CompareFilter{Op: op.Nq, Field: field, Value: value}
	return f
}

func Gt(field string, value any) Filter {
	f := CompareFilter{Op: op.Gt, Field: field, Value: value}
	return f
}

func Gte(field string, value any) Filter {
	f := CompareFilter{Op: op.Gte, Field: field, Value: value}
	return f
}

func Lt(field string, value any) Filter {
	f := CompareFilter{Op: op.Lt, Field: field, Value: value}
	return f
}

func WithMeta(meta map[string]string, filter Filter) MetaFilter {
	return &MetaFilterImpl{
		meta:   meta,
		filter: filter,
	}
}

func Lte(field string, value any) Filter {
	f := CompareFilter{Op: op.Lte, Field: field, Value: value}
	return f
}

func And(filters ...Filter) Filter {
	f := LogicalFilter{Op: op.And, Children: filters}
	return f
}

func Or(filters ...Filter) Filter {
	f := LogicalFilter{Op: op.Or, Children: filters}
	return f
}

func Not(expr Filter) Filter {
	f := LogicalFilter{Op: op.Not, Children: []Filter{expr}}
	return f
}

func Like(field string, value string, method ...string) Filter {
	var m string = ""
	if len(method) > 0 {
		m = method[0]
	}
	f := LikeFilter{Field: field, Value: value, Method: m}
	return f
}

func In(field string, args ...any) Filter {
	f := InFilter{
		Field:  field,
		Values: args,
	}
	return f
}

func Between(field string, from, to any) Filter {
	f := BetweenFilter{Field: field, From: from, To: to}
	return f
}
