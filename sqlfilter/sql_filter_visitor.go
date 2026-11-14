package sqlfilter

import (
	"errors"
	"fmt"
	"strings"

	"github.com/XPD-Kasun/filters"
	"github.com/XPD-Kasun/filters/op"
)

type SqlFilter struct {
	sqlConfig    SqlConfig
	sqlParamChar rune
	buf          *strings.Builder
	args         []string
	paramOrder   int
	usesCaching  bool
	cmdCache     map[filters.Filter]string
	meta         map[string]string
}

var _ filters.Visitor = &SqlFilter{}

type SqlVariant string

const (
	Postgres    SqlVariant = "Postgres"
	CockroachDB SqlVariant = "CockroachDB"
	Oracle      SqlVariant = "Oracle"
	MsSqlServer SqlVariant = "MsSqlServer"
	MySql       SqlVariant = "MySql"
	SQLite      SqlVariant = "SQLite"
	ClickHouse  SqlVariant = "ClickHouse"
)

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
func (v *SqlFilter) VisitComparison(op filters.CompareFilter) error {
	return comparison(v.buf, v.sqlConfig, op.Field, opMap[op.Op], op.Value)
}

// VisitLogical implements [filters.Visitor].
func (v *SqlFilter) VisitLogical(l filters.LogicalFilter) error {
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

// VisitBetween implements filters.Visitor.
func (v *SqlFilter) VisitBetween(a filters.BetweenFilter) error {
	w := v.buf
	w.WriteString(QuoteIdentifier(v.sqlConfig, a.Field))
	w.WriteString(" BETWEEN ")
	v.paramOrder++
	w.WriteString(FormatPlaceholder(v.sqlConfig, v.paramOrder))
	w.WriteString(" and ")
	v.paramOrder++
	w.WriteString(FormatPlaceholder(v.sqlConfig, v.paramOrder))

	if arg1, err := ParseWithConfig(v.sqlConfig, a.From); err == nil {
		v.args = append(v.args, arg1)
	} else {
		return err
	}

	if arg2, err := ParseWithConfig(v.sqlConfig, a.To); err == nil {
		v.args = append(v.args, arg2)
	} else {
		return err
	}

	return nil
}

func comparison(w *strings.Builder, cfg SqlConfig, field string, op string, val any) error {
	w.WriteString(QuoteIdentifier(cfg, field))
	w.WriteRune(' ')
	w.WriteString(op)
	w.WriteRune(' ')

	if valx, err := ParseWithConfig(cfg, val); err == nil {
		w.WriteString(valx)
	} else {
		return err
	}
	return nil
}

func (v *SqlFilter) Args() []string {
	return v.args
}

func (v *SqlFilter) Config() SqlConfig {
	return v.sqlConfig
}

// VisitIn implements filters.Visitor.
func (v *SqlFilter) VisitIn(op filters.InFilter) error {
	w := v.buf

	var s []string = make([]string, 0, len(op.Values))

	for _, val := range op.Values {
		if valParsed, err := ParseWithConfig(v.sqlConfig, val); err == nil {
			s = append(s, valParsed)
		} else {
			return err
		}
	}

	if len(s) == 0 {
		return errors.New("sqlfilter: IN clause cannot have empty values")
	}

	w.WriteString(QuoteIdentifier(v.sqlConfig, op.Field))
	w.WriteString(" IN (")
	w.WriteString(strings.Join(s, ","))
	w.WriteString(")")

	return nil
}

// VisitLike implements filters.Visitor.
func (v *SqlFilter) VisitLike(op filters.LikeFilter) error {
	w := v.buf
	quotedField := QuoteIdentifier(v.sqlConfig, op.Field)

	// regular like semantics
	if op.Method == "" {
		v.paramOrder++
		placeholder := FormatPlaceholder(v.sqlConfig, v.paramOrder)
		w.WriteString(quotedField)
		w.WriteString(" LIKE ")
		w.WriteString(placeholder)

		v.args = append(v.args, op.Value)
		return nil
	}

	// regex support in sql
	if op.Method == "regex" {
		v.paramOrder++
		placeholder := FormatPlaceholder(v.sqlConfig, v.paramOrder)
		w.WriteString(FormatRegex(v.sqlConfig.Name, quotedField, placeholder))
		v.args = append(v.args, op.Value)
		return nil
	}

	return fmt.Errorf("sqlfilter: unsupported LIKE method %q", op.Method)
}

func (v *SqlFilter) String() string {
	return v.buf.String()
}

func NewEmpty(sqlConfig SqlConfig) *SqlFilter {
	if sqlConfig.Name == "" && sqlConfig.SqlParamChar == 0 {
		sqlConfig = PostgresConfig
	}
	return &SqlFilter{
		cmdCache:     map[filters.Filter]string{},
		buf:          &strings.Builder{},
		sqlConfig:    sqlConfig,
		sqlParamChar: sqlConfig.SqlParamChar,
	}
}

func New(sqlConfig SqlConfig, f filters.Filter) (*SqlFilter, error) {

	sf := NewEmpty(sqlConfig)

	if mf, ok := f.(filters.MetaFilter); ok {
		sf.meta = mf.GetAll()
	}

	if err := f.Accept(sf); err != nil {
		return nil, err
	}
	return sf, nil
}
