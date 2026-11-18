package filters

// Core Filter interface.
type Filter interface {
	Accept(Visitor) error
}

// MetaFilter is a filter that contains metadata.
// Currently used as an optional root filter when additional information
// needed to be provided to the visitor.
type MetaFilter interface {
	Filter
	Get(key string) string
	GetAll() map[string]string
}

// Core Visitor interface. A visitor is an any implementation that implements [Visitor]
// interface. A visitor is accepted by a [Filter] which then invokes visitor methods as filter
// traverses through the filter tree.
type Visitor interface {
	VisitLogical(LogicalFilter) error
	VisitComparison(CompareFilter) error
	VisitLike(LikeFilter) error
	VisitBetween(BetweenFilter) error
	VisitIn(InFilter) error
	VisitNot(NotFilter) error
}
