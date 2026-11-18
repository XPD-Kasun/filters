package sqlfilter

import (
	"log/slog"
	"maps"
	"strconv"
	"strings"
)

// PlaceholderFunc defines the SQL parameter placeholder for a given dialect.
// order is the parameter index that should be used for this placeholder.
// Implementations that do not use numbered parameters, such as MySQL, can ignore the order.
//
// Eg:
//
//	func(order int) string {
//		return "$"+ strconv.Itoa(i) // returns $1, $2, $3 ...etc
//	}
type PlaceholderFunc func(order int) string

// Optimize numbered placeholders for first 64 params using cache.

var dollarCache [65]string
var colonCache [65]string
var atpCache [65]string

func init() {
	for i := 1; i <= 64; i++ {
		dollarCache[i] = "$" + strconv.Itoa(i)
		colonCache[i] = ":" + strconv.Itoa(i)
		atpCache[i] = "@p" + strconv.Itoa(i)
	}
}

// PlaceholderFunc implementations for common SQL dialects.
var (

	// QuestionPlaceholder returns "?" ignoring the order. Eg: MySQL, SQLite, ClickHouse.
	QuestionPlaceholder PlaceholderFunc = func(int) string { return "?" }

	// DollarPlaceholder returns "$1", "$2" ...etc. Eg: PostgreSQL, CockroachDB.
	DollarPlaceholder PlaceholderFunc = func(order int) string {
		// check the cache if parameter count is less than 64.
		if order >= 1 && order <= 64 {
			return dollarCache[order]
		}
		return "$" + strconv.Itoa(order)
	}

	// ColonPlaceholder returns ":1", ":2", ...etc. Eg: Oracle pl/sql
	ColonPlaceholder PlaceholderFunc = func(order int) string {
		if order >= 1 && order <= 64 {
			return colonCache[order]
		}
		return ":" + strconv.Itoa(order)
	}

	// AtpPlaceholder returns "@p1", "@p2". This is used in MSSQLServer(T-SQL).
	AtpPlaceholder PlaceholderFunc = func(order int) string {
		if order >= 1 && order <= 64 {
			return atpCache[order]
		}
		return "@p" + strconv.Itoa(order)
	}
)

// QuoteStyle controls which character pair is used to delimit identifiers.
// This is required when using SQL keywords as field names. In built dialects,
// are configured correctly with the accepted QuoteStyle.
//
// eg:
//
//	order.id = 10 #is incorrect.
//
// Lets say we target Oracle, then we need to use [QuoteDouble] QuoteStyle:
//
//	"order"."id"
type QuoteStyle int

const (
	// QuoteDouble uses double-quotes: `"order"."id"`. Used by Postgres, Oracle, SQLite, CockroachDB.
	QuoteDouble QuoteStyle = iota
	// QuoteBacktick uses backticks: ```order``.``id```. Used by MySQL and ClickHouse.
	QuoteBacktick
	// QuoteBracket uses square brackets: `[order].[id]`. Used by MSSQLServer.
	QuoteBracket
)

// QuotePolicy controls whether identifiers are quoted.
type QuotePolicy int

const (
	// QuoteNever emits identifiers as is. The caller guarantees they are safe.
	// This is the default.
	QuoteNever QuotePolicy = iota
	// QuoteAlways quotes every identifier segment unconditionally.
	QuoteAlways
)

// quoteStyleChars returns the open and close runes for the given style.
func quoteStyleChars(s QuoteStyle) (open, close rune) {
	switch s {
	case QuoteBacktick:
		return '`', '`'
	case QuoteBracket:
		return '[', ']'
	default: // QuoteDouble
		return '"', '"'
	}
}

// SqlConfig holds all dialect specific settings for a SQL filter visitor.
// A SqlConfig is considered to be immutable. Use the WithXXX methods to build the config.
type SqlConfig struct {
	Name        SqlVariant
	Placeholder PlaceholderFunc

	// Quoting
	quoteStyle  QuoteStyle
	quotePolicy QuotePolicy

	// IdentMaxLen is the maximum byte length of an unquoted identifier. 0 means no limit.
	IdentMaxLen int
	// ParamLimit is the maximum number of bind parameters the dialect accepts. 0 means no limit.
	ParamLimit int

	// LIKE escape character. 0 means no ESCAPE clause is emitted.
	LikeEscape rune

	// When true, booleans are rendered as 1 or 0, instead of TRUE or FALSE.
	NumericBooleans bool

	// Logger is an optional slogger. Can be nil
	Logger *slog.Logger

	fields map[string]string
}

// WithFields returns a copy of the SqlConfig with specified segment mappings applied.
// Keys are the segments provided by the filter. Values are the mapped segments for SQL
// generation.
//
// eg: If used following, generated query will use 'c' and 'o' as alias.
//
//	m := map[string]string {
//	 "customer": "c"
//	 "order": "o"
//	}
func (c SqlConfig) WithFields(m map[string]string) SqlConfig {
	if m == nil {
		c.fields = nil
		return c
	}
	c.fields = maps.Clone(m)
	return c
}

// WithQuoting returns a copy of the SqlConfig with the specified quote policy.
// Use QuoteAlways to quote every identifier, or QuoteNever (default) to
// emit names as is.
func (c SqlConfig) WithQuoting(policy QuotePolicy) SqlConfig {
	c.quotePolicy = policy
	return c
}

// WithPlaceholder returns a copy of the SqlConfig with a custom PlaceholderFunc.
func (c SqlConfig) WithPlaceholder(fn PlaceholderFunc) SqlConfig {
	c.Placeholder = fn
	return c
}

// WithLogger returns a copy of the SqlConfig with the given logger.
func (c SqlConfig) WithLogger(logger *slog.Logger) SqlConfig {
	c.Logger = logger
	return c
}

// GetQuoteStyle returns the QuoteStyle for this config.
func (c SqlConfig) GetQuoteStyle() QuoteStyle {
	return c.quoteStyle
}

// GetQuotePolicy returns the QuotePolicy for this config.
func (c SqlConfig) GetQuotePolicy() QuotePolicy {
	return c.quotePolicy
}

// UsePostgres returns a SqlConfig preset for PostgreSQL.
//
//   - Placeholder: $1, $2, …
//   - Quote style: double-quote
//   - Identifier limit: 63 bytes
//   - Parameter limit: 65535
func UsePostgres() SqlConfig {
	return SqlConfig{
		Name:        PostgresVariant,
		Placeholder: DollarPlaceholder,
		quoteStyle:  QuoteDouble,
		quotePolicy: QuoteNever,
		IdentMaxLen: 63,
		ParamLimit:  65535,
		LikeEscape:  '\\',
	}
}

// UseCockroachDB returns a SqlConfig preset for CockroachDB (Postgres-compatible).
func UseCockroachDB() SqlConfig {
	return SqlConfig{
		Name:        CockroachDBVariant,
		Placeholder: DollarPlaceholder,
		quoteStyle:  QuoteDouble,
		quotePolicy: QuoteNever,
		IdentMaxLen: 63,
		ParamLimit:  65535,
		LikeEscape:  '\\',
	}
}

// UseMySQL returns a SqlConfig preset for MySQL.
//
//   - Placeholder: ?
//   - Quote style: backtick
//   - Identifier limit: 64 bytes
//   - Parameter limit: 65535
func UseMySQL() SqlConfig {
	return SqlConfig{
		Name:        MySqlVariant,
		Placeholder: QuestionPlaceholder,
		quoteStyle:  QuoteBacktick,
		quotePolicy: QuoteNever,
		IdentMaxLen: 64,
		ParamLimit:  65535,
		LikeEscape:  '\\', // MySQL treats \ specially; callers may override.
	}
}

// UseClickHouse returns a SqlConfig preset for ClickHouse.
func UseClickHouse() SqlConfig {
	return SqlConfig{
		Name:        ClickHouseVariant,
		Placeholder: QuestionPlaceholder,
		quoteStyle:  QuoteBacktick,
		quotePolicy: QuoteNever,
		IdentMaxLen: 0,
		ParamLimit:  0,
		LikeEscape:  '\\',
	}
}

// UseSQLite returns a SqlConfig preset for SQLite.
func UseSQLite() SqlConfig {
	return SqlConfig{
		Name:        SQLiteVariant,
		Placeholder: QuestionPlaceholder,
		quoteStyle:  QuoteDouble,
		quotePolicy: QuoteNever,
		IdentMaxLen: 0,
		ParamLimit:  0,
		LikeEscape:  '\\',
	}
}

// UseMsSqlServer returns a SqlConfig preset for Microsoft SQL Server.
//
//   - Placeholder: @p1, @p2, …
//   - Quote style: bracket [name]
//   - Parameter limit: 2100
//   - NumericBooleans: true (SQL Server doesnt have a native BOOLEAN type)
func UseMsSqlServer() SqlConfig {
	return SqlConfig{
		Name:            MsSqlServerVariant,
		Placeholder:     AtpPlaceholder,
		quoteStyle:      QuoteBracket,
		quotePolicy:     QuoteNever,
		IdentMaxLen:     128,
		ParamLimit:      2100,
		NumericBooleans: true,
		LikeEscape:      '\\',
	}
}

// UseOracle returns a SqlConfig preset for Oracle Database.
//
//   - Placeholder: :1, :2, …
//   - Quote style: double-quote (Oracle preserves exact case when quoted)
//   - Identifier limit: 128 bytes (30 for Oracle < 12.2)
//   - NumericBooleans: true (Oracle does not have BOOLEAN type below 12c)
func UseOracle() SqlConfig {
	return SqlConfig{
		Name:            OracleVariant,
		Placeholder:     ColonPlaceholder,
		quoteStyle:      QuoteDouble,
		quotePolicy:     QuoteNever,
		IdentMaxLen:     128,
		ParamLimit:      0,
		NumericBooleans: true,
		LikeEscape:      '\\',
	}
}

// fieldsLookup returns the mapped field name for a given segment.
//
// filedsLookup first checks the exact match satisfying case-sensitive
// lookup. If case-sensitive value not found, case-insensitive lookup
// is used by checking the lowercased values of segment and fields.
func (c SqlConfig) fieldsLookup(segment string) (string, bool) {
	if len(c.fields) == 0 {
		return "", false
	}
	// Exact match first.
	if v, ok := c.fields[segment]; ok {
		return v, true
	}
	// Heuristic, assuming filter is given with lowercase values.
	lower := strings.ToLower(segment)
	if v, ok := c.fields[lower]; ok {
		return v, true
	}
	// Full search
	for k := range c.fields {
		if strings.ToLower(k) == lower {
			return c.fields[k], true
		}
	}

	return "", false
}
