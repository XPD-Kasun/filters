package filters

type Filter interface {
	Accept(Visitor) error
}

type MetaFilter interface {
	Filter
	Get(key string) string
	GetAll() map[string]string
}

type Visitor interface {
	VisitLogical(LogicalFilter) error
	VisitComparison(CompareFilter) error
	VisitLike(LikeFilter) error
	VisitBetween(BetweenFilter) error
	VisitIn(InFilter) error
}
