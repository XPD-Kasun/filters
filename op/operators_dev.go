//go:build dev

package op

type Operator string
type LogicalOperator string

const (
	Eq      Operator        = "="
	Nq      Operator        = "!="
	Gt      Operator        = ">"
	Gte     Operator        = ">="
	Lt      Operator        = "<"
	Lte     Operator        = "<="
	And     LogicalOperator = "&"
	Or      LogicalOperator = "|"
	Not     LogicalOperator = "!"
	Like    Operator        = "LIKE "
	In      Operator        = "IN "
	Between Operator        = "BETWEEN "
)
