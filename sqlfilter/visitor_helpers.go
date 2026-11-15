package sqlfilter

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"
)

// Parse converts an arbitrary Go value into a valid SQL literal string using default Postgres configuration.
func Parse(val any) (string, error) {
	return ParseWithConfig(PostgresConfig, val)
}

// ParseWithConfig converts an arbitrary Go value into a valid SQL literal string based on the given SqlConfig.
func ParseWithConfig(cfg SqlConfig, val any) (string, error) {
	if val == nil {
		return "NULL", nil
	}

	switch v := val.(type) {
	case string:
		escaped := strings.ReplaceAll(v, "'", "''")
		return fmt.Sprintf("'%s'", escaped), nil
	case bool:
		if cfg.NumericBooleans {
			if v {
				return "1", nil
			}
			return "0", nil
		}
		if v {
			return "TRUE", nil
		}
		return "FALSE", nil
	case time.Time:
		return fmt.Sprintf("'%s'", v.Format(time.DateTime)), nil
	case int64:
		return strconv.FormatInt(v, 10), nil
	case int:
		return strconv.Itoa(v), nil
	case int32:
		return strconv.FormatInt(int64(v), 10), nil
	case int16:
		return strconv.FormatInt(int64(v), 10), nil
	case int8:
		return strconv.FormatInt(int64(v), 10), nil
	case uint:
		return strconv.FormatUint(uint64(v), 10), nil
	case uint64:
		return strconv.FormatUint(v, 10), nil
	case uint32:
		return strconv.FormatUint(uint64(v), 10), nil
	case uint16:
		return strconv.FormatUint(uint64(v), 10), nil
	case uint8:
		return strconv.FormatUint(uint64(v), 10), nil
	case float32:
		return strconv.FormatFloat(float64(v), 'g', -1, 32), nil
	case float64:
		return strconv.FormatFloat(v, 'g', -1, 64), nil
	}
	return "", errors.New("cannot parse given value to string")
}

// QuoteIdentifier safely quotes a table or column name based on the dialect's QuoteOpen and QuoteClose characters.
func QuoteIdentifier(cfg SqlConfig, field string) string {
	if !cfg.QuoteIdentifiers || cfg.QuoteOpen == 0 || cfg.QuoteClose == 0 {
		return field
	}
	if field == "" || field == "*" {
		return field
	}

	parts := strings.Split(field, ".")
	for i, p := range parts {
		trimmed := strings.TrimSpace(p)
		if trimmed == "*" {
			parts[i] = trimmed
			continue
		}
		// Avoid double-quoting if already quoted
		if len(trimmed) > 1 && rune(trimmed[0]) == cfg.QuoteOpen && rune(trimmed[len(trimmed)-1]) == cfg.QuoteClose {
			parts[i] = trimmed
			continue
		}
		parts[i] = fmt.Sprintf("%c%s%c", cfg.QuoteOpen, trimmed, cfg.QuoteClose)
	}
	return strings.Join(parts, ".")
}

// FormatRegex constructs a SQL regex expression based on the target SQL variant.
func FormatRegex(variant SqlVariant, field, pattern string) string {
	switch variant {
	case Postgres, CockroachDB:
		// Postgres & CockroachDB standard POSIX regex operator
		return fmt.Sprintf("%s ~ %s", field, pattern)
	case Oracle:
		// Oracle REGEXP_LIKE function
		return fmt.Sprintf("REGEXP_LIKE(%s, %s)", field, pattern)
	case ClickHouse:
		// ClickHouse uses match(field, pattern)
		return fmt.Sprintf("match(%s, %s)", field, pattern)
	case MsSqlServer:
		// SQL Server doesn't have native REGEXP; uses LIKE
		return fmt.Sprintf("%s LIKE %s", field, pattern)
	case MySql, SQLite:
		// MySQL and SQLite use REGEXP operator
		return fmt.Sprintf("%s REGEXP %s", field, pattern)
	default:
		return fmt.Sprintf("%s REGEXP %s", field, pattern)
	}
}

// FormatPlaceholder constructs a SQL parameter placeholder based on the variant and parameter order.
func FormatPlaceholder(cfg SqlConfig, order int) string {
	switch cfg.Name {
	case MySql, SQLite, ClickHouse:
		if cfg.SqlParamChar == '?' {
			return "?"
		}
		return fmt.Sprintf("%c%d", cfg.SqlParamChar, order)
	case MsSqlServer:
		if cfg.SqlParamChar == '@' {
			return fmt.Sprintf("@p%d", order)
		}
		return fmt.Sprintf("%c%d", cfg.SqlParamChar, order)
	case Oracle:
		return fmt.Sprintf(":%d", order)
	case Postgres, CockroachDB:
		return fmt.Sprintf("$%d", order)
	default:
		if cfg.SqlParamChar == '?' {
			return "?"
		}
		if cfg.SqlParamChar != 0 {
			return fmt.Sprintf("%c%d", cfg.SqlParamChar, order)
		}
		return fmt.Sprintf("$%d", order)
	}
}

func pathSegCount(path string) int {
	return len(strings.Split(path, "."))
}

func pathEntityAndField(path string) (string, string) {

	pathSeg := strings.Split(path, ".")

	if len(pathSeg) == 1 {
		return "", path
	}

	return pathSeg[len(pathSeg)-2], pathSeg[len(pathSeg)-1]

}

func pathEntitiesAndField(path string) (string, string) {
	segCount := strings.Count(path, ".")

	if segCount <= 1 {
		return "", path
	}

	lastDotIndex := strings.LastIndex(path, ".")
	return path[:lastDotIndex], path[lastDotIndex+1:]
}
