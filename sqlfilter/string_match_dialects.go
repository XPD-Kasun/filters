package sqlfilter

import (
	"fmt"
	"strings"
)

// StringMatchDialect generates dialect-specific SQL fragments for string
// matching operations. Each method returns:
//
//   - sql: a SQL expression using "?" as a placeholder token for each bind value
//   - args: the bind values in the order they appear in sql
//   - err: non-nil when the operation is unsupported for this dialect
//
// The [SqlFilter] visitor replaces each "?" token with the actual dialect
// placeholder (e.g. $1, :1, @p1) and appends the returned args to the
// argument list.
type StringMatchDialect interface {
	// Generates SQL that checks a given string contains the value.
	// Usual translation: LIKE '%value%'
	Contains(field string, value any, insensitive bool) (string, []any, error)

	// Generates SQL that checks a given string is at the begining of
	// the provided value. Usual translation: LIKE 'value%'
	StartsWith(field string, value any, insensitive bool) (string, []any, error)

	// Generates SQL that checks a given string is at the end of
	// the provided value. Usual translation: LIKE '%value'
	EndsWith(field string, value any, insensitive bool) (string, []any, error)

	// Like generates SQL for a raw LIKE/ILIKE pattern supplied by the caller.
	// This is the one used for translating [filters.WildcardMatch]
	Like(field string, pattern any, insensitive bool) (string, []any, error)

	// Regex generates SQL for a regular-expression match.
	Regex(field string, pattern any, insensitive bool) (string, []any, error)

	// Checks string equality for the value passed using case-sensitive check.
	// This should be used for string equality implementation for [filters.Eq].
	StringEqual(field string, value any, insensitive bool) (string, []any, error)
}

// escClause returns " ESCAPE '<char>'" for a LIKE expression, or empty string
// when no escape character is configured.
func escClause(escape rune) string {
	if escape == 0 {
		return ""
	}
	return fmt.Sprintf(" ESCAPE '%c'", escape)
}

// toStr extracts a string from an any value, falling back to fmt.Sprintf.
func toStr(v any) string {
	if s, ok := v.(string); ok {
		return s
	}
	return fmt.Sprintf("%v", v)
}

// escapeLikePattern escapes and translate filter wildcard semantics
// into SQL semantics. See [filters.WildcardMatch].
//
// eg:
// escapeLikePattern("**sda_sd%fo*fd?df", '\\') -> *sda\_sd\%fo%fd_df
func escapeLikePattern(val string, escape rune) string {
	if escape == 0 || escape > 127 {
		return val
	}
	esc := byte(escape)
	var b strings.Builder
	b.Grow(len(val))

	for i := 0; i < len(val); i++ {
		c := val[i]

		switch c {
		case '*', '?':
			// search rune forward to see if its ** or ?? while
			// not exeeding the iteration beyond len(val).
			if i+1 < len(val) && val[i+1] == c {
				// * or ? escaped version
				b.WriteByte(c)
				i++
			} else if c == '*' {
				b.WriteByte('%')
			} else {
				b.WriteByte('_')
			}

		case '%', '_', esc:
			b.WriteByte(esc)
			b.WriteByte(c)

		default:
			b.WriteByte(c)
		}
	}

	return b.String()
}

// PostgresStringMatch implements [StringMatchDialect] for PostgreSQL and
// CockroachDB. ILIKE is used for case-insensitive LIKE operations; ~ / ~*
// are used for regex matching.
type PostgresStringMatch struct {
	Escape rune
}

func (d PostgresStringMatch) likeOp(insensitive bool) string {
	if insensitive {
		return "ILIKE"
	}
	return "LIKE"
}

func (d PostgresStringMatch) Contains(field string, value any, insensitive bool) (string, []any, error) {
	pat := "%" + escapeLikePattern(toStr(value), d.Escape) + "%"
	return fmt.Sprintf("%s %s ?%s", field, d.likeOp(insensitive), escClause(d.Escape)), []any{pat}, nil
}

func (d PostgresStringMatch) StartsWith(field string, value any, insensitive bool) (string, []any, error) {
	pat := escapeLikePattern(toStr(value), d.Escape) + "%"
	return fmt.Sprintf("%s %s ?%s", field, d.likeOp(insensitive), escClause(d.Escape)), []any{pat}, nil
}

func (d PostgresStringMatch) EndsWith(field string, value any, insensitive bool) (string, []any, error) {
	pat := "%" + escapeLikePattern(toStr(value), d.Escape)
	return fmt.Sprintf("%s %s ?%s", field, d.likeOp(insensitive), escClause(d.Escape)), []any{pat}, nil
}

func (d PostgresStringMatch) Like(field string, pattern any, insensitive bool) (string, []any, error) {
	return fmt.Sprintf("%s %s ?", field, d.likeOp(insensitive)), []any{toStr(pattern)}, nil
}

func (d PostgresStringMatch) Regex(field string, pattern any, insensitive bool) (string, []any, error) {
	if insensitive {
		return fmt.Sprintf("%s ~* ?", field), []any{toStr(pattern)}, nil
	}
	return fmt.Sprintf("%s ~ ?", field), []any{toStr(pattern)}, nil
}

func (d PostgresStringMatch) StringEqual(field string, value any, insensitive bool) (string, []any, error) {
	if insensitive {
		return fmt.Sprintf("LOWER(%s) = LOWER(?)", field), []any{toStr(value)}, nil
	}
	return fmt.Sprintf("%s = ?", field), []any{toStr(value)}, nil
}

// MySQLStringMatch implements [StringMatchDialect] for MySQL.
// MySQL LIKE is case-insensitive by default on non-binary columns, so plain
// LIKE is emitted regardless of the insensitive flag. REGEXP is used for
// regex; case-sensitivity follows the column collation.
type MySQLStringMatch struct {
	Escape rune
}

func (d MySQLStringMatch) Contains(field string, value any, insensitive bool) (string, []any, error) {
	pat := "%" + escapeLikePattern(toStr(value), d.Escape) + "%"
	return fmt.Sprintf("%s LIKE ?%s", field, escClause(d.Escape)), []any{pat}, nil
}

func (d MySQLStringMatch) StartsWith(field string, value any, insensitive bool) (string, []any, error) {
	pat := escapeLikePattern(toStr(value), d.Escape) + "%"
	return fmt.Sprintf("%s LIKE ?%s", field, escClause(d.Escape)), []any{pat}, nil
}

func (d MySQLStringMatch) EndsWith(field string, value any, insensitive bool) (string, []any, error) {
	pat := "%" + escapeLikePattern(toStr(value), d.Escape)
	return fmt.Sprintf("%s LIKE ?%s", field, escClause(d.Escape)), []any{pat}, nil
}

func (d MySQLStringMatch) Like(field string, pattern any, insensitive bool) (string, []any, error) {
	return fmt.Sprintf("%s LIKE ?", field), []any{toStr(pattern)}, nil
}

func (d MySQLStringMatch) Regex(field string, pattern any, insensitive bool) (string, []any, error) {
	// MySQL REGEXP case-sensitivity follows the column collation.
	// Callers may embed (?i) in the pattern for portable insensitivity.
	return fmt.Sprintf("%s REGEXP ?", field), []any{toStr(pattern)}, nil
}

func (d MySQLStringMatch) StringEqual(field string, value any, insensitive bool) (string, []any, error) {
	if insensitive {
		return fmt.Sprintf("LOWER(%s) = LOWER(?)", field), []any{toStr(value)}, nil
	}
	return fmt.Sprintf("%s = ?", field), []any{toStr(value)}, nil
}

// ClickHouseStringMatch implements [StringMatchDialect] for ClickHouse.
// ILIKE is supported since v22.4; match() is used for regex.
type ClickHouseStringMatch struct {
	Escape rune
}

func (d ClickHouseStringMatch) likeOp(insensitive bool) string {
	if insensitive {
		return "ILIKE"
	}
	return "LIKE"
}

func (d ClickHouseStringMatch) Contains(field string, value any, insensitive bool) (string, []any, error) {
	pat := "%" + escapeLikePattern(toStr(value), d.Escape) + "%"
	return fmt.Sprintf("%s %s ?%s", field, d.likeOp(insensitive), escClause(d.Escape)), []any{pat}, nil
}

func (d ClickHouseStringMatch) StartsWith(field string, value any, insensitive bool) (string, []any, error) {
	pat := escapeLikePattern(toStr(value), d.Escape) + "%"
	return fmt.Sprintf("%s %s ?%s", field, d.likeOp(insensitive), escClause(d.Escape)), []any{pat}, nil
}

func (d ClickHouseStringMatch) EndsWith(field string, value any, insensitive bool) (string, []any, error) {
	pat := "%" + escapeLikePattern(toStr(value), d.Escape)
	return fmt.Sprintf("%s %s ?%s", field, d.likeOp(insensitive), escClause(d.Escape)), []any{pat}, nil
}

func (d ClickHouseStringMatch) Like(field string, pattern any, insensitive bool) (string, []any, error) {
	return fmt.Sprintf("%s %s ?", field, d.likeOp(insensitive)), []any{toStr(pattern)}, nil
}

func (d ClickHouseStringMatch) Regex(field string, pattern any, insensitive bool) (string, []any, error) {
	// match() is always case-sensitive at the engine level.
	// Callers may embed (?i) in the pattern for insensitivity.
	return fmt.Sprintf("match(%s, ?)", field), []any{toStr(pattern)}, nil
}

func (d ClickHouseStringMatch) StringEqual(field string, value any, insensitive bool) (string, []any, error) {
	if insensitive {
		return fmt.Sprintf("LOWER(%s) = LOWER(?)", field), []any{toStr(value)}, nil
	}
	return fmt.Sprintf("%s = ?", field), []any{toStr(value)}, nil
}

// ---- Oracle ----------------------------------------------------------------

// OracleStringMatch implements [StringMatchDialect] for Oracle Database.
// Oracle LIKE is case-sensitive; UPPER(field) LIKE UPPER(?) is emitted for
// case-insensitive matching. REGEXP_LIKE with the optional 'i' flag is used
// for regex.
type OracleStringMatch struct {
	Escape rune
}

func (d OracleStringMatch) likeExpr(field string, insensitive bool) string {
	esc := escClause(d.Escape)
	if insensitive {
		return fmt.Sprintf("UPPER(%s) LIKE UPPER(?)%s", field, esc)
	}
	return fmt.Sprintf("%s LIKE ?%s", field, esc)
}

func (d OracleStringMatch) Contains(field string, value any, insensitive bool) (string, []any, error) {
	pat := "%" + escapeLikePattern(toStr(value), d.Escape) + "%"
	return d.likeExpr(field, insensitive), []any{pat}, nil
}

func (d OracleStringMatch) StartsWith(field string, value any, insensitive bool) (string, []any, error) {
	pat := escapeLikePattern(toStr(value), d.Escape) + "%"
	return d.likeExpr(field, insensitive), []any{pat}, nil
}

func (d OracleStringMatch) EndsWith(field string, value any, insensitive bool) (string, []any, error) {
	pat := "%" + escapeLikePattern(toStr(value), d.Escape)
	return d.likeExpr(field, insensitive), []any{pat}, nil
}

func (d OracleStringMatch) Like(field string, pattern any, insensitive bool) (string, []any, error) {
	return d.likeExpr(field, insensitive), []any{toStr(pattern)}, nil
}

func (d OracleStringMatch) Regex(field string, pattern any, insensitive bool) (string, []any, error) {
	if insensitive {
		return fmt.Sprintf("REGEXP_LIKE(%s, ?, 'i')", field), []any{toStr(pattern)}, nil
	}
	return fmt.Sprintf("REGEXP_LIKE(%s, ?)", field), []any{toStr(pattern)}, nil
}

func (d OracleStringMatch) StringEqual(field string, value any, insensitive bool) (string, []any, error) {
	if insensitive {
		return fmt.Sprintf("UPPER(%s) = UPPER(?)", field), []any{toStr(value)}, nil
	}
	return fmt.Sprintf("%s = ?", field), []any{toStr(value)}, nil
}

// MsSqlServerStringMatch implements [StringMatchDialect] for Microsoft SQL
// Server. SQL Server has no ILIKE; LOWER(field) LIKE LOWER(?) is emitted for
// case-insensitive matching. SQL Server has no native regex support; a LIKE
// fallback is emitted for Regex calls.
type MsSqlServerStringMatch struct {
	Escape rune
}

func (d MsSqlServerStringMatch) likeExpr(field string, insensitive bool) string {
	esc := escClause(d.Escape)
	if insensitive {
		return fmt.Sprintf("LOWER(%s) LIKE LOWER(?)%s", field, esc)
	}
	return fmt.Sprintf("%s LIKE ?%s", field, esc)
}

func (d MsSqlServerStringMatch) Contains(field string, value any, insensitive bool) (string, []any, error) {
	pat := "%" + escapeLikePattern(toStr(value), d.Escape) + "%"
	return d.likeExpr(field, insensitive), []any{pat}, nil
}

func (d MsSqlServerStringMatch) StartsWith(field string, value any, insensitive bool) (string, []any, error) {
	pat := escapeLikePattern(toStr(value), d.Escape) + "%"
	return d.likeExpr(field, insensitive), []any{pat}, nil
}

func (d MsSqlServerStringMatch) EndsWith(field string, value any, insensitive bool) (string, []any, error) {
	pat := "%" + escapeLikePattern(toStr(value), d.Escape)
	return d.likeExpr(field, insensitive), []any{pat}, nil
}

func (d MsSqlServerStringMatch) Like(field string, pattern any, insensitive bool) (string, []any, error) {
	return d.likeExpr(field, insensitive), []any{toStr(pattern)}, nil
}

func (d MsSqlServerStringMatch) Regex(field string, pattern any, insensitive bool) (string, []any, error) {
	// SQL Server has no native REGEXP; fall back to LIKE (limited wildcard support).
	return fmt.Sprintf("%s LIKE ?", field), []any{toStr(pattern)}, nil
}

func (d MsSqlServerStringMatch) StringEqual(field string, value any, insensitive bool) (string, []any, error) {
	if insensitive {
		return fmt.Sprintf("LOWER(%s) = LOWER(?)", field), []any{toStr(value)}, nil
	}
	return fmt.Sprintf("%s = ?", field), []any{toStr(value)}, nil
}

// SQLiteStringMatch implements [StringMatchDialect] for SQLite.
// SQLite LIKE is case-insensitive for ASCII by default; LOWER(field) LIKE
// LOWER(?) is emitted for explicit insensitivity. REGEXP requires a registered
// user function.
type SQLiteStringMatch struct {
	Escape rune
}

func (d SQLiteStringMatch) likeExpr(field string, insensitive bool) string {
	esc := escClause(d.Escape)
	if insensitive {
		return fmt.Sprintf("LOWER(%s) LIKE LOWER(?)%s", field, esc)
	}
	return fmt.Sprintf("%s LIKE ?%s", field, esc)
}

func (d SQLiteStringMatch) Contains(field string, value any, insensitive bool) (string, []any, error) {
	pat := "%" + escapeLikePattern(toStr(value), d.Escape) + "%"
	return d.likeExpr(field, insensitive), []any{pat}, nil
}

func (d SQLiteStringMatch) StartsWith(field string, value any, insensitive bool) (string, []any, error) {
	pat := escapeLikePattern(toStr(value), d.Escape) + "%"
	return d.likeExpr(field, insensitive), []any{pat}, nil
}

func (d SQLiteStringMatch) EndsWith(field string, value any, insensitive bool) (string, []any, error) {
	pat := "%" + escapeLikePattern(toStr(value), d.Escape)
	return d.likeExpr(field, insensitive), []any{pat}, nil
}

func (d SQLiteStringMatch) Like(field string, pattern any, insensitive bool) (string, []any, error) {
	return d.likeExpr(field, insensitive), []any{toStr(pattern)}, nil
}

func (d SQLiteStringMatch) Regex(field string, pattern any, insensitive bool) (string, []any, error) {
	// REGEXP requires a user-registered function in SQLite.
	return fmt.Sprintf("%s REGEXP ?", field), []any{toStr(pattern)}, nil
}

func (d SQLiteStringMatch) StringEqual(field string, value any, insensitive bool) (string, []any, error) {
	if insensitive {
		return fmt.Sprintf("LOWER(%s) = LOWER(?)", field), []any{toStr(value)}, nil
	}
	return fmt.Sprintf("%s = ?", field), []any{toStr(value)}, nil
}

// defaultStringMatch returns the default [StringMatchDialect] implementation for the given [SqlVariant].
func defaultStringMatch(variant SqlVariant, escape rune) StringMatchDialect {
	switch variant {
	case PostgresVariant, CockroachDBVariant:
		return PostgresStringMatch{Escape: escape}
	case MySqlVariant:
		return MySQLStringMatch{Escape: escape}
	case ClickHouseVariant:
		return ClickHouseStringMatch{Escape: escape}
	case OracleVariant:
		return OracleStringMatch{Escape: escape}
	case MsSqlServerVariant:
		return MsSqlServerStringMatch{Escape: escape}
	case SQLiteVariant:
		return SQLiteStringMatch{Escape: escape}
	default:
		return PostgresStringMatch{Escape: escape}
	}
}
