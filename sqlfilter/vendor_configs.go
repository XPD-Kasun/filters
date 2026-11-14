package sqlfilter

import "log/slog"

type SqlConfig struct {
	Name             SqlVariant
	SqlParamChar     rune
	QuoteIdentifiers bool
	QuoteOpen        rune
	QuoteClose       rune
	NumericBooleans  bool
	Logger           *slog.Logger
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
