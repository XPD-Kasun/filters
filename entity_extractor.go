package filters

import (
	"strings"
)

// EntityVisitor traverses a Filter AST to extract unique entity names
// from field paths.
type EntityVisitor struct {
	entities map[string]struct{}
	ordered  []string
}

var _ Visitor = &EntityVisitor{}

// NewEntityVisitor creates a new EntityVisitor
func NewEntityVisitor() *EntityVisitor {
	return &EntityVisitor{
		entities: make(map[string]struct{}),
	}
}

// VisitNot implements [Visitor]
func (v *EntityVisitor) VisitNot(field NotFilter) error {
	if IsEmpty(field.Child) {
		return nil
	}
	return field.Child.Accept(v)
}

// addFromField extracts entity prefixes from a dot-delimited field path
// (e.g., "order.customer.price" -> "order", "customer").
func (v *EntityVisitor) addFromField(field string) {
	parts := strings.Split(field, ".")
	if len(parts) <= 1 {
		return
	}
	// All parts except the last segment (the column name) are entity/table names
	for _, entity := range parts[:len(parts)-1] {
		entity = strings.TrimSpace(entity)
		if entity == "" {
			continue
		}
		if _, exists := v.entities[entity]; !exists {
			v.entities[entity] = struct{}{}
			v.ordered = append(v.ordered, entity)
		}
	}
}

// VisitComparison implements [Visitor].
func (v *EntityVisitor) VisitComparison(op CompareFilter) error {
	v.addFromField(op.Field)
	return nil
}

// VisitLike implements [Visitor].
func (v *EntityVisitor) VisitLike(op LikeFilter) error {
	v.addFromField(op.Field)
	return nil
}

// VisitBetween implements [Visitor].
func (v *EntityVisitor) VisitBetween(op BetweenFilter) error {
	v.addFromField(op.Field)
	return nil
}

// VisitIn implements [Visitor].
func (v *EntityVisitor) VisitIn(op InFilter) error {
	v.addFromField(op.Field)
	return nil
}

// VisitLogical implements [Visitor].
func (v *EntityVisitor) VisitLogical(l LogicalFilter) error {
	for _, child := range l.Children {
		if child != nil {
			if err := child.Accept(v); err != nil {
				return err
			}
		}
	}
	return nil
}

// Entities returns the collected unique entity names in stable discovery order.
func (v *EntityVisitor) Entities() []string {
	result := make([]string, len(v.ordered))
	copy(result, v.ordered)
	return result
}

// ExtractEntities traverses the given Filter AST and returns all unique entity names.
func ExtractEntities(f Filter) ([]string, error) {
	if f == nil {
		return nil, nil
	}
	v := NewEntityVisitor()
	if err := f.Accept(v); err != nil {
		return nil, err
	}
	return v.Entities(), nil
}
