package filters

import "github.com/XPD-Kasun/filters/op"

func EQ(field string, value any) Filter {
	f := Comparison{Op: op.Eq, Field: field, Value: value}
	return f
}

func NQ(field string, value any) Filter {
	f := Comparison{Op: op.Nq, Field: field, Value: value}
	return f
}

func GT(field string, value any) Filter {
	f := Comparison{Op: op.Gt, Field: field, Value: value}
	return f
}

func GTE(field string, value any) Filter {
	f := Comparison{Op: op.Gte, Field: field, Value: value}
	return f
}

func LT(field string, value any) Filter {
	f := Comparison{Op: op.Lt, Field: field, Value: value}
	return f
}

func LTE(field string, value any) Filter {
	f := Comparison{Op: op.Lte, Field: field, Value: value}
	return f
}

func AND(filters ...Filter) Filter {
	f := Logical{Op: op.And, Children: filters}
	return f
}

func OR(filters ...Filter) Filter {
	f := Logical{Op: op.Or, Children: filters}
	return f
}

func NOT(expr Filter) Filter {
	f := Logical{Op: op.Or, Children: []Filter{expr}}
	return f
}

func LIKE(field string, value string, method string) Filter {
	f := Like{Field: field, Value: value, Method: method}
	return f
}

func IN(field string, args ...any) Filter {
	f := In{
		Field:  field,
		Values: args,
	}
	return f
}

func BETWEEN(field string, from, to any) Filter {
	f := Between{Field: field, From: from, To: to}
	return f
}
