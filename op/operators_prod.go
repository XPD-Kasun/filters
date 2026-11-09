//go:build !dev

package op

type Operator int
type LogicalOperator int

const (
	Eq Operator = iota
	Nq
	Gt
	Gte
	Lt
	Lte
	Like
	In
	Between
)

const (
	And LogicalOperator = LogicalOperator(Between) + iota + 1
	Or
	Not
)
