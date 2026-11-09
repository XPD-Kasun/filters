package filters

import "github.com/XPD-Kasun/filters/op"

type Like struct {
	Field  string
	Value  string
	Method string
}

func (op Like) Accept(v Visitor) error {
	return v.VisitLike(op)
}

type In struct {
	Field  string
	Values []any
}

func (op In) Accept(v Visitor) error {
	return v.VisitIn(op)
}

type Between struct {
	Field string
	From  any
	To    any
}

func (op Between) Accept(v Visitor) error {
	return v.VisitBetween(op)
}

type Logical struct {
	Op       op.LogicalOperator
	Children []Filter
}

func (l Logical) Accept(v Visitor) error {
	return v.VisitLogical(l)
}

type Comparison struct {
	Op    op.Operator
	Field string
	Value any
}

func (c Comparison) Accept(v Visitor) error {
	return v.VisitComparison(c)
}
