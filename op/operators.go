// op package defines the supported operators that are represented by the
// main API
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
	Not     Operator        = "NOT"
	Like    Operator        = "LIKE"
	In      Operator        = "IN"
	Between Operator        = "BETWEEN"
)
