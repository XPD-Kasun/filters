package sqlfilter

import (
	"errors"
	"fmt"
	"strings"

	"github.com/XPD-Kasun/filters"
	"github.com/XPD-Kasun/filters/op"
)

type SqlFilter struct {
	buf         *strings.Builder
	args        []string
	paramOrder  int
	usesCaching bool
	cmdCache    map[filters.Filter]string
}

var opLogicalMap = map[op.LogicalOperator]string{
	op.And: "and",
	op.Or:  "or",
	op.Not: "not",
}

var opMap = map[op.Operator]string{
	op.Gt:      ">",
	op.Gte:     ">=",
	op.Between: "BETWEEN",
	op.Eq:      "=",
	op.Lt:      "<",
	op.Lte:     "<=",
	op.In:      "IN",
	op.Like:    "LIKE",
	op.Nq:      "!=",
}

// VisitComparison implements [filters.Visitor].
func (v *SqlFilter) VisitComparison(op filters.Comparison) error {
	return comparison(v.buf, op.Field, opMap[op.Op], op.Value)
}

// VisitLogical implements [filters.Visitor].
func (v *SqlFilter) VisitLogical(l filters.Logical) error {
	w := v.buf

	switch l.Op {
	case op.And, op.Or:
		w.WriteString("(")
		for i, a := range l.Children {
			if err := a.Accept(v); err != nil {
				return err
			}
			if i != len(l.Children)-1 {
				w.WriteString(" ")
				w.WriteString(opLogicalMap[l.Op])
				w.WriteString(" ")
			}
		}
		w.WriteString(")")

	case op.Not:
		if len(l.Children) == 1 {
			w.WriteString("NOT(")
			if err := l.Children[0].Accept(v); err != nil {
				return err
			}
			w.WriteString(")")
		} else {
			return errors.New("NOT operator can only have a single child")
		}
	}

	return nil

}

var _ filters.Visitor = &SqlFilter{}

// VisitBetween implements filters.Visitor.
func (v *SqlFilter) VisitBetween(a filters.Between) error {
	w := v.buf
	w.WriteString(a.Field)
	w.WriteString(fmt.Sprintf(" BETWEEN $%d and $%d", v.paramOrder+1, v.paramOrder+2))
	v.paramOrder += 2

	if arg1, err := Parse(a.From); err == nil {
		v.args = append(v.args, arg1)
	} else {
		return err
	}

	if arg2, err := Parse(a.To); err == nil {
		v.args = append(v.args, arg2)
	} else {
		return err
	}

	return nil
}

func comparison(w *strings.Builder, field string, op string, val any) error {

	w.WriteString(field)
	w.WriteRune(' ')
	w.WriteString(op)
	w.WriteRune(' ')

	if valx, err := Parse(val); err == nil {
		w.WriteString(valx)
	} else {
		return err
	}
	return nil
}

// VisitIn implements filters.Visitor.
func (v *SqlFilter) VisitIn(op filters.In) error {
	w := v.buf

	w.WriteString(op.Field)
	w.WriteString(" IN (")

	var s []string = make([]string, len(op.Values))

	for i, val := range op.Values {
		if v, err := Parse(val); err == nil {
			s[i] = v
		}
	}

	w.WriteString(strings.Join(s, ","))
	w.WriteString(")")

	return nil
}

// VisitLike implements filters.Visitor.
func (v *SqlFilter) VisitLike(op filters.Like) error {
	w := v.buf

	//regular like semantics
	if op.Method == "" {
		w.WriteString(op.Field)
		w.WriteString(" LIKE '")
		w.WriteString(op.Value)
		w.WriteRune('\'')
	}

	return nil

}

func (v *SqlFilter) String() string {
	return v.buf.String()
}

func NewEmpty() *SqlFilter {
	return &SqlFilter{
		cmdCache: map[filters.Filter]string{},
		buf:      &strings.Builder{},
	}
}

func New(f filters.Filter) (*SqlFilter, error) {
	sf := &SqlFilter{
		cmdCache: map[filters.Filter]string{},
		buf:      &strings.Builder{},
	}
	if err := f.Accept(sf); err != nil {
		return nil, err
	}
	return sf, nil
}
