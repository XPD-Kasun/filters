package filters

import (
	"maps"

	"github.com/XPD-Kasun/filters/op"
)

// StringMatchMethod tells the filter how to match text fields,
// such as checking for substrings, wildcards, or regex patterns.
type StringMatchMethod string

const (
	// Contains checks if the field has this substring anywhere inside it.
	Contains StringMatchMethod = "Contains"
	// StartsWith checks if the field begins with the given text.
	StartsWith StringMatchMethod = "StartsWith"
	// EndsWith checks if the field finishes with the given text.
	EndsWith StringMatchMethod = "EndsWith"
	// WildcardMatch matches text using * (any characters) and ? (single character).
	WildcardMatch StringMatchMethod = "WildcardMatch"
	// Regex runs a regular expression pattern match against the field.
	Regex StringMatchMethod = "Regex"
	// StringEqual checks for an exact string match, with optional case sensitivity.
	StringEqual StringMatchMethod = "StringEqual"
)

// StringMatchFilter holds everything needed for matching string values,
// including the field name, target value, match method, and case rules.
type StringMatchFilter struct {
	Field         string
	Value         string
	Method        StringMatchMethod
	CaseSensitive bool
}

// Accept lets a visitor process this string match filter.
func (sm StringMatchFilter) Accept(v Visitor) error {
	return v.VisitStringMatch(sm)
}

// InFilter checks whether a field's value matches any item in a list,
// similar to an SQL IN (...) clause.
type InFilter struct {
	Field  string
	Values []any
}

// Accept lets a visitor process this IN filter.
func (op InFilter) Accept(v Visitor) error {
	return v.VisitIn(op)
}

// BetweenFilter checks if a field falls within a range,
// between the From and To values (inclusive).
type BetweenFilter struct {
	Field string
	From  any
	To    any
}

// Accept lets a visitor process this BETWEEN filter.
func (op BetweenFilter) Accept(v Visitor) error {
	return v.VisitBetween(op)
}

// LogicalFilter groups several child filters together using AND or OR.
type LogicalFilter struct {
	Op       op.LogicalOperator
	Children []Filter
}

// Accept lets a visitor process this logical group and its children.
func (l LogicalFilter) Accept(v Visitor) error {
	return v.VisitLogical(l)
}

// NotFilter flips the result of its child filter (like NOT in SQL).
type NotFilter struct {
	Child Filter
}

// Accept lets a visitor process this NOT filter and its inner child.
func (n NotFilter) Accept(v Visitor) error {
	return v.VisitNot(n)
}

// CompareFilter compares a field against a value using an operator
// like equals (=), greater than (>), less than (<), and so on.
type CompareFilter struct {
	Op    op.Operator
	Field string
	Value any
}

// Accept lets a visitor process the comparison.
func (c CompareFilter) Accept(v Visitor) error {
	return v.VisitComparison(c)
}

// MetaFilterImpl wraps another filter and lets you store extra key-value
// metadata alongside it, which visitors can read during traversal.
type MetaFilterImpl struct {
	meta   map[string]string
	filter Filter
}

var _ MetaFilter = &MetaFilterImpl{}

// Accept forwards the visitor to the underlying wrapped filter.
func (m *MetaFilterImpl) Accept(v Visitor) error {
	return m.filter.Accept(v)
}

// Get looks up a metadata value by key.
func (m *MetaFilterImpl) Get(key string) string {
	return m.meta[key]
}

// GetAll returns a copy of all metadata key-value pairs.
func (m *MetaFilterImpl) GetAll() map[string]string {
	return maps.Clone(m.meta)
}

// Set stores a metadata key-value pair.
func (m *MetaFilterImpl) Set(key, value string) {
	m.meta[key] = value
}

// EmptyFilter is a blank filter that does nothing when visited.
// It's used as a safe fallback when filters are nil or empty.
type EmptyFilter struct {
}

// Accept does nothing and returns nil.
func (em EmptyFilter) Accept(v Visitor) error {
	return nil
}
