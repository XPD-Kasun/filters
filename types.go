package filters

import (
	"maps"

	"github.com/XPD-Kasun/filters/op"
)

// Defines core filter implementations.

type StringMatchMethod string

const (
	Contains      StringMatchMethod = "Contains"
	StartsWith    StringMatchMethod = "StartsWith"
	EndsWith      StringMatchMethod = "EndsWith"
	WildcardMatch StringMatchMethod = "WildcardMatch"
	Regex         StringMatchMethod = "Regex"
)

type StringMatchFilter struct {
	Field         string
	Value         string
	Method        StringMatchMethod
	CaseSensitive bool
}

func (sm StringMatchFilter) Accept(v Visitor) error {
	return v.VisitStringMatch(sm)
}

type InFilter struct {
	Field  string
	Values []any
}

func (op InFilter) Accept(v Visitor) error {
	return v.VisitIn(op)
}

type BetweenFilter struct {
	Field string
	From  any
	To    any
}

func (op BetweenFilter) Accept(v Visitor) error {
	return v.VisitBetween(op)
}

type LogicalFilter struct {
	Op       op.LogicalOperator
	Children []Filter
}

func (l LogicalFilter) Accept(v Visitor) error {
	return v.VisitLogical(l)
}

type NotFilter struct {
	Child Filter
}

func (n NotFilter) Accept(v Visitor) error {
	return v.VisitNot(n)
}

type CompareFilter struct {
	Op    op.Operator
	Field string
	Value any
}

func (c CompareFilter) Accept(v Visitor) error {
	return v.VisitComparison(c)
}

type MetaFilterImpl struct {
	meta   map[string]string
	filter Filter
}

var _ MetaFilter = &MetaFilterImpl{}

func (m *MetaFilterImpl) Accept(v Visitor) error {
	return m.filter.Accept(v)
}

func (m *MetaFilterImpl) Get(key string) string {
	return m.meta[key]
}

func (m *MetaFilterImpl) GetAll() map[string]string {
	return maps.Clone(m.meta)
}

func (m *MetaFilterImpl) Set(key, value string) {
	m.meta[key] = value
}

type EmptyFilter struct {
}

func (em EmptyFilter) Accept(v Visitor) error {
	return nil
}
