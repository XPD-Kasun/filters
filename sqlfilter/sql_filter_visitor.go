package sqlfilter

import (
	"filters"
	"fmt"
	"strings"
)

type SqlFilter struct {
	buf         *strings.Builder
	args        []string
	paramOrder  int
	usesCaching bool
	cmdCache    map[filters.Filter]string
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

// VisitEq implements filters.Visitor.
func (v *SqlFilter) VisitEq(op filters.Eq) error {

	return comparison(v.buf, op.Field, " = ", op.Value)

}

func comparison(w *strings.Builder, field string, op string, val any) error {

	w.WriteString(field)
	w.WriteString(op)

	if valx, err := Parse(val); err == nil {
		w.WriteString(valx)
	} else {
		return err
	}
	return nil
}

// VisitGt implements filters.Visitor.
func (v *SqlFilter) VisitGt(op filters.Gt) error {
	return comparison(v.buf, op.Field, " > ", op.Value)
}

// VisitGte implements filters.Visitor.
func (v *SqlFilter) VisitGte(op filters.Gte) error {
	return comparison(v.buf, op.Field, " >= ", op.Value)
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
	panic("unimplemented")
}

// VisitLt implements filters.Visitor.
func (v *SqlFilter) VisitLt(op filters.Lt) error {
	return comparison(v.buf, op.Field, " > ", op.Value)
}

// VisitLte implements filters.Visitor.
func (v *SqlFilter) VisitLte(op filters.Lte) error {
	return comparison(v.buf, op.Field, " > ", op.Value)
}

// VisitNq implements filters.Visitor.
func (v *SqlFilter) VisitNq(op filters.Nq) error {
	return comparison(v.buf, op.Field, " != ", op.Value)
}

func (v *SqlFilter) VisitAnd(a filters.And) error {
	w := v.buf
	for i := 0; i < len(a.Filters); i++ {
		w.WriteString("(")
		f := a.Filters[i]
		err := f.Accept(v)
		if err != nil {
			return err
		}
		if i != (len(a.Filters) - 1) {
			w.WriteString(") AND ")
		} else {
			w.WriteString(")")
		}

	}
	return nil
}

func (v *SqlFilter) VisitOr(or filters.Or) error {
	w := v.buf
	for i := 0; i < len(or.Filters); i++ {
		w.WriteString("(")
		f := or.Filters[i]
		err := f.Accept(v)
		if err != nil {
			return err
		}
		if i != (len(or.Filters) - 1) {
			w.WriteString(") OR ")
		} else {
			w.WriteString(")")
		}

	}
	return nil
}

func (v *SqlFilter) VisitNot(not filters.Not) error {
	w := v.buf
	w.WriteString("NOT (")
	err := not.Child.Accept(v)
	if err != nil {
		return err
	}
	w.WriteString(")")
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
