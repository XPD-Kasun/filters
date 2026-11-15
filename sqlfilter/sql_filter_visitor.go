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
	args         []any
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

// formatField applies entity/field segment mapping and dialect-specific identifier quoting.
func (v *SqlFilter) formatField(field string) string {
	mapped := v.mapEntityFieldNames(field)
	return QuoteIdentifier(v.sqlConfig, mapped)
}

func (v *SqlFilter) mapEntityFieldNames(pathField string) string {
	if len(v.sqlConfig.fields) == 0 {
		return pathField
	}
	entities := strings.Split(pathField, ".")
	mappedPath := make([]string, len(entities))

	for i, entity := range entities {
		normalisedEntity := entity
		if !v.sqlConfig.CaseSensitive {
			normalisedEntity = strings.ToLower(entity)
		}

		if val, hasEntity := v.sqlConfig.fields[normalisedEntity]; hasEntity {
			mappedPath[i] = val
		} else {
			mappedPath[i] = entity
		}
	}

	return strings.Join(mappedPath, ".")
}

// VisitComparison implements [filters.Visitor].
func (v *SqlFilter) VisitComparison(op filters.CompareFilter) error {
	if op.Value == nil {
		return v.compareNull(op.Field, op.Op)
	}
	return v.comparison(op.Field, opMap[op.Op], op.Value)
}

func (v *SqlFilter) compareNull(field string, operator op.Operator) error {
	w := v.buf
	var nullstr string

	switch operator {
	case op.Eq:
		nullstr = " IS NULL"
	case op.Nq:
		nullstr = " IS NOT NULL"
	default:
		return fmt.Errorf("SqlFilter: operator %v is not compatible with nil", operator)
	}

	formattedField := v.formatField(field)
	if v.sqlConfig.Logger != nil {
		v.sqlConfig.Logger.Debug("sqlfilter: visiting null comparison", "field", formattedField, "operator", nullstr)
	}

	w.WriteString(formattedField)
	w.WriteString(nullstr)
	return nil
}

// VisitLogical implements [filters.Visitor].
func (v *SqlFilter) VisitLogical(l filters.LogicalFilter) error {
	w := v.buf

	if v.sqlConfig.Logger != nil {
		v.sqlConfig.Logger.Debug("sqlfilter: visiting logical filter", "operator", opLogicalMap[l.Op], "childrenCount", len(l.Children))
	}

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
	formattedField := v.formatField(a.Field)

	if v.sqlConfig.Logger != nil {
		v.sqlConfig.Logger.Debug("sqlfilter: visiting between filter", "field", formattedField, "from", a.From, "to", a.To)
	}

	w.WriteString(formattedField)
	w.WriteString(" BETWEEN ")
	v.paramOrder++
	w.WriteString(FormatPlaceholder(v.sqlConfig, v.paramOrder))
	w.WriteString(" and ")
	v.paramOrder++
	w.WriteString(FormatPlaceholder(v.sqlConfig, v.paramOrder))

	v.args = append(v.args, a.From, a.To)
	return nil
}

func (v *SqlFilter) comparison(field string, operator string, val any) error {
	w := v.buf
	formattedField := v.formatField(field)

	if v.sqlConfig.Logger != nil {
		v.sqlConfig.Logger.Debug("sqlfilter: visiting comparison", "field", formattedField, "operator", operator, "value", val)
	}

	w.WriteString(formattedField)
	w.WriteRune(' ')
	w.WriteString(operator)
	w.WriteRune(' ')

	v.paramOrder++
	w.WriteString(FormatPlaceholder(v.sqlConfig, v.paramOrder))
	v.args = append(v.args, val)
	return nil
}

func (v *SqlFilter) Args() []any {
	return v.args
}

func (v *SqlFilter) Config() SqlConfig {
	return v.sqlConfig
}

// VisitIn implements filters.Visitor.
func (v *SqlFilter) VisitIn(op filters.InFilter) error {
	if len(op.Values) == 0 {
		return errors.New("sqlfilter: IN clause cannot have empty values")
	}

	w := v.buf
	formattedField := v.formatField(op.Field)

	if v.sqlConfig.Logger != nil {
		v.sqlConfig.Logger.Debug("sqlfilter: visiting in filter", "field", formattedField, "valuesCount", len(op.Values))
	}

	s := make([]string, len(op.Values))
	for i, val := range op.Values {
		v.paramOrder++
		s[i] = FormatPlaceholder(v.sqlConfig, v.paramOrder)
		v.args = append(v.args, val)
	}

	w.WriteString(formattedField)
	w.WriteString(" IN (")
	w.WriteString(strings.Join(s, ","))
	w.WriteString(")")

	return nil
}

// VisitLike implements filters.Visitor.
func (v *SqlFilter) VisitLike(op filters.LikeFilter) error {
	w := v.buf
	formattedField := v.formatField(op.Field)

	// regular like semantics
	if op.Method == "" {
		v.paramOrder++
		placeholder := FormatPlaceholder(v.sqlConfig, v.paramOrder)

		if v.sqlConfig.Logger != nil {
			v.sqlConfig.Logger.Debug("sqlfilter: visiting like filter", "field", formattedField, "pattern", op.Value, "placeholder", placeholder)
		}

		w.WriteString(formattedField)
		w.WriteString(" LIKE ")
		w.WriteString(placeholder)

		v.args = append(v.args, op.Value)
		return nil
	}

	// regex support in sql
	if op.Method == "regex" {
		v.paramOrder++
		placeholder := FormatPlaceholder(v.sqlConfig, v.paramOrder)

		if v.sqlConfig.Logger != nil {
			v.sqlConfig.Logger.Debug("sqlfilter: visiting regex filter", "field", formattedField, "pattern", op.Value, "placeholder", placeholder)
		}

		w.WriteString(FormatRegex(v.sqlConfig.Name, formattedField, placeholder))
		v.args = append(v.args, op.Value)
		return nil
	}

	return fmt.Errorf("sqlfilter: unsupported LIKE method %q", op.Method)
}

func (v *SqlFilter) String() string {
	return v.Sql()
}

func (v *SqlFilter) Sql() string {
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

	if sf.sqlConfig.Logger != nil {
		sf.sqlConfig.Logger.Info("sqlfilter: compiling filter to SQL", "variant", sf.sqlConfig.Name)
	}

	if mf, ok := f.(filters.MetaFilter); ok {
		sf.meta = mf.GetAll()
	}

	if err := f.Accept(sf); err != nil {
		if sf.sqlConfig.Logger != nil {
			sf.sqlConfig.Logger.Error("sqlfilter: compilation failed", "error", err)
		}
		return nil, err
	}

	if sf.sqlConfig.Logger != nil {
		sf.sqlConfig.Logger.Info("sqlfilter: compilation succeeded", "sql", sf.String(), "argsCount", len(sf.args))
	}

	return sf, nil
}
