package filters

type Filter interface {
	Accept(Visitor) error
}

type Visitor interface {
	VisitLogical(Logical) error
	VisitComparison(Comparison) error
	VisitLike(Like) error
	VisitBetween(Between) error
	VisitIn(In) error
}
