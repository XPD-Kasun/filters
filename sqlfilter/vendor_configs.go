package sqlfilter

import (
	"log/slog"
	"strings"
)

type SqlConfig struct {
	Name             SqlVariant
	SqlParamChar     rune
	QuoteIdentifiers bool
	QuoteOpen        rune
	QuoteClose       rune
	NumericBooleans  bool
	CaseSensitive    bool
	Logger           *slog.Logger

	fields map[string]string
}

// WithFields returns a copy of the SqlConfig with the specified entity-to-alias mappings.
func (c SqlConfig) WithFields(m map[string]string) SqlConfig {
	if m == nil {
		c.fields = nil
		return c
	}
	c.fields = make(map[string]string, len(m))
	for k, v := range m {
		key := k
		if !c.CaseSensitive {
			key = strings.ToLower(k)
		}
		c.fields[key] = v
	}
	return c
}

// WithCaseSensitive returns a copy of the SqlConfig with CaseSensitive set to enabled.
func (c SqlConfig) WithCaseSensitive(enabled bool) SqlConfig {
	c.CaseSensitive = enabled
	if len(c.fields) > 0 {
		return c.WithFields(c.fields)
	}
	return c
}

var PostgresConfig = SqlConfig{
	Name:             Postgres,
	SqlParamChar:     '$',
	QuoteOpen:        '"',
	QuoteClose:       '"',
	QuoteIdentifiers: false,
}

var CockroachDBConfig = SqlConfig{
	Name:             CockroachDB,
	SqlParamChar:     '$',
	QuoteOpen:        '"',
	QuoteClose:       '"',
	QuoteIdentifiers: false,
}

var MySqlConfig = SqlConfig{
	Name:             MySql,
	SqlParamChar:     '?',
	QuoteOpen:        '`',
	QuoteClose:       '`',
	QuoteIdentifiers: false,
}

var ClickHouseConfig = SqlConfig{
	Name:             ClickHouse,
	SqlParamChar:     '?',
	QuoteOpen:        '`',
	QuoteClose:       '`',
	QuoteIdentifiers: false,
}

var SQLiteConfig = SqlConfig{
	Name:             SQLite,
	SqlParamChar:     '?',
	QuoteOpen:        '"',
	QuoteClose:       '"',
	QuoteIdentifiers: false,
}

var MsSqlServerConfig = SqlConfig{
	Name:             MsSqlServer,
	SqlParamChar:     '@',
	QuoteOpen:        '[',
	QuoteClose:       ']',
	QuoteIdentifiers: false,
	NumericBooleans:  true,
}

var OracleConfig = SqlConfig{
	Name:             Oracle,
	SqlParamChar:     ':',
	QuoteOpen:        '"',
	QuoteClose:       '"',
	QuoteIdentifiers: false,
	NumericBooleans:  true,
}

// WithQuoting returns a copy of the SqlConfig with QuoteIdentifiers set to enabled.
func (c SqlConfig) WithQuoting(enabled bool) SqlConfig {
	c.QuoteIdentifiers = enabled
	return c
}

// WithLogger returns a copy of the SqlConfig with the specified structured logger attached.
func (c SqlConfig) WithLogger(logger *slog.Logger) SqlConfig {
	c.Logger = logger
	return c
}
