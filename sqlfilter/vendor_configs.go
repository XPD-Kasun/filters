package sqlfilter

type SqlConfig struct {
	Name             SqlVariant
	SqlParamChar     rune
	QuoteIdentifiers bool
	QuoteOpen        rune
	QuoteClose       rune
	NumericBooleans  bool
}

var PostgresConfig = SqlConfig{
	Name:             Postgres,
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
