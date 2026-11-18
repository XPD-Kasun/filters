// sqlfilter package contains the components that consumes the given filter and
// converts it into a SQL representation. See [SqlFilter] for more information.
package sqlfilter

import (
	"errors"
	"fmt"
	"strings"

	"github.com/XPD-Kasun/filters"
	"github.com/XPD-Kasun/filters/op"
)

// SqlFilter implements [filters.Visitor] and translates a filter to SQL syntax.
// Currently, supports following SQL dialects:
//   - MSSqlServer (T-SQL)
//   - Oracle (pl/SQL)
//   - Postgres
//   - CocroachDB
//   - MySql
//   - Clickhouse
type SqlFilter struct {
	sqlConfig  SqlConfig
	buf        *strings.Builder
	args       []any
	paramOrder int
	cmdCache   map[filters.Filter]string
	meta       map[string]string
}

var _ filters.Visitor = &SqlFilter{}

// SqlVariant identifies a SQL dialect
type SqlVariant string

const (
	PostgresVariant    SqlVariant = "Postgres"
	CockroachDBVariant SqlVariant = "CockroachDB"
	OracleVariant      SqlVariant = "Oracle"
	MsSqlServerVariant SqlVariant = "MsSqlServer"
	MySqlVariant       SqlVariant = "MySql"
	SQLiteVariant      SqlVariant = "SQLite"
	ClickHouseVariant  SqlVariant = "ClickHouse"
)

var opLogicalMap = map[op.LogicalOperator]string{
	op.And: "and",
	op.Or:  "or",
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
	op.Not:     "NOT",
}

type callOptions struct {
	startAt int // first paramOrder value (default 1)
}

// CallOption is a functional option for New and NewEmpty.
type CallOption func(*callOptions)

// StartAt offsets parameter numbering so it begins at the given value.
// Use when combining multiple SqlFilter outputs into one query.
//
//	sqlfilter.New(cfg, filter, sqlfilter.StartAt(4))
func StartAt(n int) CallOption {
	return func(o *callOptions) { o.startAt = n }
}

// formatField maps segment names using the field alias map, then applies
// the dialect specific identifier quoting style to each segment.
func (v *SqlFilter) formatField(field string) (string, error) {
	segments := strings.Split(field, ".")
	out := make([]string, len(segments))

	open, close := quoteStyleChars(v.sqlConfig.quoteStyle)
	policy := v.sqlConfig.quotePolicy

	for i, seg := range segments {
		// lookup alias
		if mapped, ok := v.sqlConfig.fieldsLookup(seg); ok {
			seg = mapped
		}

		if v.sqlConfig.IdentMaxLen > 0 && len(seg) > v.sqlConfig.IdentMaxLen {
			return "", fmt.Errorf(
				"%w: identifier %q exceeds dialect limit of %d bytes",
				filters.ErrUnsupported, seg, v.sqlConfig.IdentMaxLen,
			)
		}

		if policy == QuoteNever {
			out[i] = seg
			continue
		}

		// QuoteAlways — quote this segment.
		// Reject null bytes for all dialects.
		if strings.ContainsRune(seg, 0) {
			return "", fmt.Errorf(
				"%w: identifier %q contains a null byte, which is not allowed",
				filters.ErrUnsupported, seg,
			)
		}
		// Oracle: reject double-quote inside a double-quoted identifier.
		if v.sqlConfig.Name == OracleVariant && close == '"' && strings.ContainsRune(seg, '"') {
			return "", fmt.Errorf(
				"%w: Oracle does not allow double-quotes inside a quoted identifier: %q",
				filters.ErrUnsupported, seg,
			)
		}

		// Escape the closing character by doubling it.
		escaped := strings.ReplaceAll(seg, string(close), string(close)+string(close))
		out[i] = string(open) + escaped + string(close)
	}

	return strings.Join(out, "."), nil
}

// VisitComparison generates a where clause that satisfies the comparison filter
func (v *SqlFilter) VisitComparison(cmp filters.CompareFilter) error {
	if cmp.Value == nil {
		return v.compareNull(cmp.Field, cmp.Op)
	}
	return v.comparison(cmp.Field, opMap[cmp.Op], cmp.Value)
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
		return fmt.Errorf("%w: operator %v is not compatible with nil", filters.ErrUnsupported, operator)
	}

	formattedField, err := v.formatField(field)
	if err != nil {
		return err
	}
	if v.sqlConfig.Logger != nil {
		v.sqlConfig.Logger.Debug("sqlfilter: visiting null comparison", "field", formattedField, "operator", nullstr)
	}

	w.WriteString(formattedField)
	w.WriteString(nullstr)
	return nil
}

// VisitNot generates a SQL syntax that satisfies the Not filter
func (v *SqlFilter) VisitNot(l filters.NotFilter) error {
	if filters.IsEmpty(l.Child) {
		return nil
	}
	w := v.buf

	w.WriteString("NOT(")
	if err := l.Child.Accept(v); err != nil {
		return err
	}
	w.WriteString(")")

	return nil
}

// VisitNot generates SQL that satisfies the given [filters.LogicalFilter]
// which filters using AND and OR operators.
//
// eg:
//
//	(FilterA AND FilterB) OR FilterC
func (v *SqlFilter) VisitLogical(l filters.LogicalFilter) error {
	w := v.buf

	if v.sqlConfig.Logger != nil {
		v.sqlConfig.Logger.Debug("sqlfilter: visiting logical filter", "operator", opLogicalMap[l.Op], "childrenCount", len(l.Children))
	}

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

	return nil
}

// VisitBetween generates SQL for [filters.BetweenFilter].
//
// eg:
//
//	fieldA between valueA and valueB
func (v *SqlFilter) VisitBetween(a filters.BetweenFilter) error {
	w := v.buf
	formattedField, err := v.formatField(a.Field)
	if err != nil {
		return err
	}

	if v.sqlConfig.Logger != nil {
		v.sqlConfig.Logger.Debug("sqlfilter: visiting between filter", "field", formattedField, "from", a.From, "to", a.To)
	}

	w.WriteString(formattedField)
	w.WriteString(" BETWEEN ")
	v.paramOrder++
	w.WriteString(v.sqlConfig.Placeholder(v.paramOrder))
	w.WriteString(" and ")
	v.paramOrder++
	w.WriteString(v.sqlConfig.Placeholder(v.paramOrder))

	v.args = append(v.args, a.From, a.To)
	return nil
}

func (v *SqlFilter) comparison(field string, operator string, val any) error {
	w := v.buf
	formattedField, err := v.formatField(field)
	if err != nil {
		return err
	}

	if v.sqlConfig.Logger != nil {
		v.sqlConfig.Logger.Debug("sqlfilter: visiting comparison", "field", formattedField, "operator", operator, "value", val)
	}

	w.WriteString(formattedField)
	w.WriteRune(' ')
	w.WriteString(operator)
	w.WriteRune(' ')

	v.paramOrder++
	w.WriteString(v.sqlConfig.Placeholder(v.paramOrder))
	v.args = append(v.args, val)
	return nil
}

// Returns the SQL parameter arguments as an array for the generated placeholders.
// This array contains the provided values as is and can be passed to the driver directly.
//
// Note that the ordering of parameters are specific to the given filter and should not be used
// over multiple filters. In that case, use [StartAt] to specify the parameter numbering start
// value when using multiple filters in the same query.
func (v *SqlFilter) Args() []any {
	return v.args
}

// Returns the [SqlConfig] for this SqlFilter.
func (v *SqlFilter) Config() SqlConfig {
	return v.sqlConfig
}

// VisitIn generates SQL for [filters.InFilter].
//
// eg:
//
//	fieldA IN (valueA, valueB, valueC)
func (v *SqlFilter) VisitIn(inF filters.InFilter) error {
	if len(inF.Values) == 0 {
		return errors.New("sqlfilter: IN clause cannot have empty values")
	}

	w := v.buf
	formattedField, err := v.formatField(inF.Field)
	if err != nil {
		return err
	}

	if v.sqlConfig.Logger != nil {
		v.sqlConfig.Logger.Debug("sqlfilter: visiting in filter", "field", formattedField, "valuesCount", len(inF.Values))
	}

	s := make([]string, len(inF.Values))
	for i, val := range inF.Values {
		v.paramOrder++
		s[i] = v.sqlConfig.Placeholder(v.paramOrder)
		v.args = append(v.args, val)
	}

	w.WriteString(formattedField)
	w.WriteString(" IN (")
	w.WriteString(strings.Join(s, ","))
	w.WriteString(")")

	return nil
}

// VisitLike generates SQL for [filters.LikeFilter].
// If the [filters.LikeFilter.Method] is set to empty string,
// standard SQL Like operator syntax is used. And [filters.LikeFilter.Value]
// expected to contain SQL like syntax. If the Mehthod is "regex", then
// value is considered to be a regex and generates SQL using the syntax
// for the given dialect.
//
// eg:
//
//	fieldA like '%ID_3%'
//	email !~* '\.org$'
func (v *SqlFilter) VisitLike(likeF filters.LikeFilter) error {
	w := v.buf
	formattedField, err := v.formatField(likeF.Field)
	if err != nil {
		return err
	}

	// regular like semantics
	if likeF.Method == "" {
		v.paramOrder++
		placeholder := v.sqlConfig.Placeholder(v.paramOrder)

		if v.sqlConfig.Logger != nil {
			v.sqlConfig.Logger.Debug("sqlfilter: visiting like filter", "field", formattedField, "pattern", likeF.Value, "placeholder", placeholder)
		}

		w.WriteString(formattedField)
		w.WriteString(" LIKE ")
		w.WriteString(placeholder)

		v.args = append(v.args, likeF.Value)
		return nil
	}

	// regex support in sql
	if likeF.Method == "regex" {
		v.paramOrder++
		placeholder := v.sqlConfig.Placeholder(v.paramOrder)

		if v.sqlConfig.Logger != nil {
			v.sqlConfig.Logger.Debug("sqlfilter: visiting regex filter", "field", formattedField, "pattern", likeF.Value, "placeholder", placeholder)
		}

		w.WriteString(FormatRegex(v.sqlConfig.Name, formattedField, placeholder))
		v.args = append(v.args, likeF.Value)
		return nil
	}

	return fmt.Errorf("%w: LIKE method %q", filters.ErrUnsupported, likeF.Method)
}

// Returns the generated SQL for this filter. Implements [fmt.Stringer] interface.
func (v *SqlFilter) String() string {
	return v.Sql()
}

// Returns the generated SQL for this filter.
func (v *SqlFilter) Sql() string {
	return v.buf.String()
}

// NewEmpty creates an empty SqlFilter visitor with the given config.
// Optional CallOption values (eg. StartAt) can be provided.
func NewEmpty(sqlConfig SqlConfig, opts ...CallOption) *SqlFilter {
	if sqlConfig.Name == "" && sqlConfig.Placeholder == nil {
		sqlConfig = UsePostgres()
	}
	o := callOptions{startAt: 0}
	for _, opt := range opts {
		opt(&o)
	}
	return &SqlFilter{
		cmdCache:   map[filters.Filter]string{},
		buf:        &strings.Builder{},
		sqlConfig:  sqlConfig,
		paramOrder: o.startAt,
	}
}

// New compiles the given filter into SQL using the provided config.
// Optional CallOption values (eg. StartAt) can be provided.
func New(sqlConfig SqlConfig, f filters.Filter, opts ...CallOption) (*SqlFilter, error) {
	sf := NewEmpty(sqlConfig, opts...)

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
