package sqlfilter

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"
)

// Parse converts an arbitrary Go value into a valid SQL literal string using the default Postgres configuration.
func Parse(val any) (string, error) {
	return ParseWithConfig(UsePostgres(), val)
}

// ParseWithConfig converts an arbitrary Go value into a valid SQL string based on the given SqlConfig.
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
	case fmt.Stringer:
		str := v.String()
		escaped := strings.ReplaceAll(str, "'", "''")
		return fmt.Sprintf("'%s'", escaped), nil
	}
	return "", errors.New("cannot parse given value to string")
}

// QuoteIdentifier quotes a dot-separated identifier path using the dialect's
// QuoteStyle, only when the config's QuotePolicy is QuoteAlways. Each segment
// is quoted and escaped independently so that "schema"."table" is produced
// rather than the erroneous "schema.table".
//
// Rules:
//   - QuoteNever: returns the path unchanged.
//   - Null bytes are rejected for all dialects.
//   - Oracle rejects double quotes inside a doubly quoted identifier.
//   - The closing character is escaped by doubling (SQL standard).
func QuoteIdentifier(cfg SqlConfig, path string) (string, error) {
	if cfg.quotePolicy == QuoteNever {
		return path, nil
	}
	if path == "" || path == "*" {
		return path, nil
	}

	open, close := quoteStyleChars(cfg.quoteStyle)
	parts := strings.Split(path, ".")
	out := make([]string, len(parts))

	for i, seg := range parts {
		seg = strings.TrimSpace(seg)
		if seg == "*" {
			out[i] = seg
			continue
		}
		if strings.ContainsRune(seg, 0) {
			return "", fmt.Errorf("sqlfilter: identifier %q contains a null byte", seg)
		}
		if cfg.Name == OracleVariant && close == '"' && strings.ContainsRune(seg, '"') {
			return "", fmt.Errorf("sqlfilter: Oracle does not allow double-quotes inside a quoted identifier: %q", seg)
		}
		escaped := strings.ReplaceAll(seg, string(close), string(close)+string(close))
		out[i] = string(open) + escaped + string(close)
	}
	return strings.Join(out, "."), nil
}

// FormatRegex constructs a SQL regex expression based on the target SQL variant.
func FormatRegex(variant SqlVariant, field, pattern string) string {
	switch variant {
	case PostgresVariant, CockroachDBVariant:
		return fmt.Sprintf("%s ~ %s", field, pattern)
	case OracleVariant:
		return fmt.Sprintf("REGEXP_LIKE(%s, %s)", field, pattern)
	case ClickHouseVariant:
		return fmt.Sprintf("match(%s, %s)", field, pattern)
	case MsSqlServerVariant:
		return fmt.Sprintf("%s LIKE %s", field, pattern)
	case MySqlVariant, SQLiteVariant:
		return fmt.Sprintf("%s REGEXP %s", field, pattern)
	default:
		return fmt.Sprintf("%s REGEXP %s", field, pattern)
	}
}

// Counts the segment count in a given path.
func pathSegCount(path string) int {
	return len(strings.Split(path, "."))
}

// Get the last two segments of a given path and returns them.
//
// eg:
//
//	"order.customer.id" => "customer", "id"
func pathEntityAndField(path string) (string, string) {
	pathSeg := strings.Split(path, ".")
	if len(pathSeg) == 1 {
		return "", path
	}
	return pathSeg[len(pathSeg)-2], pathSeg[len(pathSeg)-1]
}

// Split the path into two parts by the last part such that all segments
// except last one is considered a single segment called entities and last segment
// as field.
//
// eg:
//
//	"order.customer.id" => "order.customer", "id"
func pathEntitiesAndField(path string) (string, string) {
	segCount := strings.Count(path, ".")
	if segCount <= 1 {
		return "", path
	}
	lastDotIndex := strings.LastIndex(path, ".")
	return path[:lastDotIndex], path[lastDotIndex+1:]
}
