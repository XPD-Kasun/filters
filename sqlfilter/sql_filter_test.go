package sqlfilter

import (
	"testing"

	"github.com/XPD-Kasun/filters"
)

func TestPathHelpers(t *testing.T) {

	path := "Order.customer.payment.id"

	t.Run("pathSegCount", func(tt *testing.T) {

		result := pathSegCount(path)

		if result != 4 {
			tt.Errorf("pathSegCount returns invalid path lentgh. got %d, expected %d", result, 4)
		}
	})

	t.Run("pathEntityAndField", func(tt *testing.T) {

		entity, field := pathEntityAndField(path)

		if entity != "payment" || field != "id" {
			tt.Errorf("returns wrong entity or field. got %s, %s, expected %s, %s", entity, field, "payment", "id")
		}
	})

	t.Run("pathEntitiesAndField", func(tt *testing.T) {

		entities, field := pathEntitiesAndField(path)

		if entities != "Order.customer.payment" || field != "id" {
			tt.Errorf(
				"returns wrong entity or field. got %s, %s, expected %s, %s",
				entities,
				field,
				"Order.customer.payment",
				"id",
			)
		}
	})

}

func TestStringMatchDialects(t *testing.T) {
	tests := []struct {
		name     string
		cfg      SqlConfig
		filter   filters.StringMatchFilter
		wantSql  string
		wantArgs []any
	}{
		// Postgres
		{
			name:     "Postgres Contains Case-Insensitive",
			cfg:      UsePostgres(),
			filter:   filters.StringMatchFilter{Field: "title", Value: "Go", Method: filters.Contains, CaseSensitive: false},
			wantSql:  "title ILIKE $1 ESCAPE '\\'",
			wantArgs: []any{"%Go%"},
		},
		{
			name:     "Postgres Contains Case-Sensitive",
			cfg:      UsePostgres(),
			filter:   filters.StringMatchFilter{Field: "title", Value: "Go", Method: filters.Contains, CaseSensitive: true},
			wantSql:  "title LIKE $1 ESCAPE '\\'",
			wantArgs: []any{"%Go%"},
		},
		{
			name:     "Postgres StartsWith",
			cfg:      UsePostgres(),
			filter:   filters.StringMatchFilter{Field: "title", Value: "Hello", Method: filters.StartsWith, CaseSensitive: false},
			wantSql:  "title ILIKE $1 ESCAPE '\\'",
			wantArgs: []any{"Hello%"},
		},
		{
			name:     "Postgres EndsWith",
			cfg:      UsePostgres(),
			filter:   filters.StringMatchFilter{Field: "title", Value: "World", Method: filters.EndsWith, CaseSensitive: false},
			wantSql:  "title ILIKE $1 ESCAPE '\\'",
			wantArgs: []any{"%World"},
		},
		{
			name:     "Postgres WildcardMatch",
			cfg:      UsePostgres(),
			filter:   filters.StringMatchFilter{Field: "code", Value: "A_%", Method: filters.WildcardMatch, CaseSensitive: false},
			wantSql:  "code ILIKE $1",
			wantArgs: []any{"A_%"},
		},
		{
			name:     "Postgres Regex Insensitive",
			cfg:      UsePostgres(),
			filter:   filters.StringMatchFilter{Field: "email", Value: "@gmail\\.com$", Method: filters.Regex, CaseSensitive: false},
			wantSql:  "email ~* $1",
			wantArgs: []any{"@gmail\\.com$"},
		},
		{
			name:     "Postgres Regex Sensitive",
			cfg:      UsePostgres(),
			filter:   filters.StringMatchFilter{Field: "email", Value: "@gmail\\.com$", Method: filters.Regex, CaseSensitive: true},
			wantSql:  "email ~ $1",
			wantArgs: []any{"@gmail\\.com$"},
		},
		// Oracle
		{
			name:     "Oracle Contains Insensitive",
			cfg:      UseOracle(),
			filter:   filters.StringMatchFilter{Field: "name", Value: "test", Method: filters.Contains, CaseSensitive: false},
			wantSql:  "UPPER(name) LIKE UPPER(:1) ESCAPE '\\'",
			wantArgs: []any{"%test%"},
		},
		{
			name:     "Oracle Contains Sensitive",
			cfg:      UseOracle(),
			filter:   filters.StringMatchFilter{Field: "name", Value: "test", Method: filters.Contains, CaseSensitive: true},
			wantSql:  "name LIKE :1 ESCAPE '\\'",
			wantArgs: []any{"%test%"},
		},
		{
			name:     "Oracle Regex Insensitive",
			cfg:      UseOracle(),
			filter:   filters.StringMatchFilter{Field: "name", Value: "^A.*", Method: filters.Regex, CaseSensitive: false},
			wantSql:  "REGEXP_LIKE(name, :1, 'i')",
			wantArgs: []any{"^A.*"},
		},
		// SQL Server
		{
			name:     "MsSqlServer Contains Insensitive",
			cfg:      UseMsSqlServer(),
			filter:   filters.StringMatchFilter{Field: "desc", Value: "item", Method: filters.Contains, CaseSensitive: false},
			wantSql:  "LOWER(desc) LIKE LOWER(@p1) ESCAPE '\\'",
			wantArgs: []any{"%item%"},
		},
		{
			name:     "MsSqlServer Contains Sensitive",
			cfg:      UseMsSqlServer(),
			filter:   filters.StringMatchFilter{Field: "desc", Value: "item", Method: filters.Contains, CaseSensitive: true},
			wantSql:  "desc LIKE @p1 ESCAPE '\\'",
			wantArgs: []any{"%item%"},
		},
		// MySQL
		{
			name:     "MySQL Contains",
			cfg:      UseMySQL(),
			filter:   filters.StringMatchFilter{Field: "tag", Value: "tech", Method: filters.Contains, CaseSensitive: false},
			wantSql:  "tag LIKE ? ESCAPE '\\'",
			wantArgs: []any{"%tech%"},
		},
		{
			name:     "MySQL Regex",
			cfg:      UseMySQL(),
			filter:   filters.StringMatchFilter{Field: "tag", Value: "[0-9]+", Method: filters.Regex, CaseSensitive: false},
			wantSql:  "tag REGEXP ?",
			wantArgs: []any{"[0-9]+"},
		},
		// ClickHouse
		{
			name:     "ClickHouse Contains Insensitive",
			cfg:      UseClickHouse(),
			filter:   filters.StringMatchFilter{Field: "host", Value: "srv", Method: filters.Contains, CaseSensitive: false},
			wantSql:  "host ILIKE ? ESCAPE '\\'",
			wantArgs: []any{"%srv%"},
		},
		{
			name:     "ClickHouse Regex",
			cfg:      UseClickHouse(),
			filter:   filters.StringMatchFilter{Field: "host", Value: ".*\\.local", Method: filters.Regex, CaseSensitive: false},
			wantSql:  "match(host, ?)",
			wantArgs: []any{".*\\.local"},
		},
		// SQLite
		{
			name:     "SQLite Contains Insensitive",
			cfg:      UseSQLite(),
			filter:   filters.StringMatchFilter{Field: "city", Value: "York", Method: filters.Contains, CaseSensitive: false},
			wantSql:  "LOWER(city) LIKE LOWER(?) ESCAPE '\\'",
			wantArgs: []any{"%York%"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			sf, err := New(tc.cfg, tc.filter)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if sf.String() != tc.wantSql {
				t.Errorf("got sql %q, want %q", sf.String(), tc.wantSql)
			}
			if len(sf.Args()) != len(tc.wantArgs) {
				t.Fatalf("got args count %d, want %d", len(sf.Args()), len(tc.wantArgs))
			}
			for i, arg := range tc.wantArgs {
				if sf.Args()[i] != arg {
					t.Errorf("arg[%d] = %v, want %v", i, sf.Args()[i], arg)
				}
			}
		})
	}
}
