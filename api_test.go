package filters_test

import (
	"log/slog"
	"reflect"
	"strings"
	"testing"

	f "github.com/XPD-Kasun/filters"
	q "github.com/XPD-Kasun/filters/fields"
	"github.com/XPD-Kasun/filters/sqlfilter"
)

func TestBasic(t *testing.T) {
	s := f.Or(
		f.And(f.Eq("se", 12), f.Gt("fsd", 2432)),
		f.Eq("thx", "success"),
		f.Between("fdsi", 10, 200),
		f.Between("fdsi23", 10, 200),
	)
	filter, err := sqlfilter.New(sqlfilter.PostgresConfig, s)
	if err != nil {
		t.Fatal(err.Error())
	}
	expected := "((se = $1 and fsd > $2) or thx = $3 or fdsi BETWEEN $4 and $5 or fdsi23 BETWEEN $6 and $7)"
	if filter.String() != expected {
		t.Errorf("got %q, want %q", filter.String(), expected)
	}
	expectedArgs := []any{12, 2432, "success", 10, 200, 10, 200}
	if !reflect.DeepEqual(filter.Args(), expectedArgs) {
		t.Errorf("args mismatch: got %v, want %v", filter.Args(), expectedArgs)
	}
}

type B struct{}

type A struct {
	Price q.IntField[B]
	Name  q.StringField[B]
}

func TestFn(t *testing.T) {
	var a A = A{
		Price: q.NewIntField[B]("price", ""),
		Name:  q.NewStringField[B]("Stesll", "df"),
	}
	s := f.Or(a.Price.Gte(200), a.Name.Eq("fnsdf"))
	w := sqlfilter.NewEmpty(sqlfilter.PostgresConfig)
	if err := s.Accept(w); err != nil {
		t.Fatal(err)
	}
	expected := "(price >= $1 or df.Stesll = $2)"
	if w.String() != expected {
		t.Errorf("got %q, want %q", w.String(), expected)
	}
	expectedArgs := []any{200, "fnsdf"}
	if !reflect.DeepEqual(w.Args(), expectedArgs) {
		t.Errorf("args mismatch: got %v, want %v", w.Args(), expectedArgs)
	}
}

func TestFn2(t *testing.T) {
	var a A = A{
		Price: q.NewIntField[B]("price", ""),
		Name:  q.NewStringField[B]("Stesll", "df"),
	}
	s := a.Price.Gte(200)
	w := sqlfilter.NewEmpty(sqlfilter.PostgresConfig)
	if err := s.Accept(w); err != nil {
		t.Fatal(err)
	}
	expected := "price >= $1"
	if w.String() != expected {
		t.Errorf("got %q, want %q", w.String(), expected)
	}
	expectedArgs := []any{200}
	if !reflect.DeepEqual(w.Args(), expectedArgs) {
		t.Errorf("args mismatch: got %v, want %v", w.Args(), expectedArgs)
	}
}

func TestNotFilter(t *testing.T) {
	s := f.Not(f.Eq("status", "deleted"))
	sf, err := sqlfilter.New(sqlfilter.PostgresConfig, s)
	if err != nil {
		t.Fatal(err)
	}
	expected := "NOT(status = $1)"
	if sf.String() != expected {
		t.Errorf("got %q, want %q", sf.String(), expected)
	}
	if len(sf.Args()) != 1 || sf.Args()[0] != "deleted" {
		t.Errorf("args mismatch: got %v", sf.Args())
	}
}

func TestNilComparison(t *testing.T) {
	// Eq nil -> IS NULL
	filterEq := f.Eq("deleted_at", nil)
	sfEq, err := sqlfilter.New(sqlfilter.PostgresConfig, filterEq)
	if err != nil {
		t.Fatal(err)
	}
	if sfEq.String() != "deleted_at IS NULL" {
		t.Errorf("got %q, want deleted_at IS NULL", sfEq.String())
	}
	if len(sfEq.Args()) != 0 {
		t.Errorf("expected 0 args for IS NULL, got %d", len(sfEq.Args()))
	}

	// Nq nil -> IS NOT NULL
	filterNq := f.Nq("deleted_at", nil)
	sfNq, err := sqlfilter.New(sqlfilter.PostgresConfig, filterNq)
	if err != nil {
		t.Fatal(err)
	}
	if sfNq.String() != "deleted_at IS NOT NULL" {
		t.Errorf("got %q, want deleted_at IS NOT NULL", sfNq.String())
	}
	if len(sfNq.Args()) != 0 {
		t.Errorf("expected 0 args for IS NOT NULL, got %d", len(sfNq.Args()))
	}
}

func TestRegexDialects(t *testing.T) {
	tests := []struct {
		name     string
		cfg      sqlfilter.SqlConfig
		expected string
	}{
		{"Postgres", sqlfilter.PostgresConfig, "username ~ $1"},
		{"CockroachDB", sqlfilter.CockroachDBConfig, "username ~ $1"},
		{"MySQL", sqlfilter.MySqlConfig, "username REGEXP ?"},
		{"ClickHouse", sqlfilter.ClickHouseConfig, "match(username, ?)"},
		{"Oracle", sqlfilter.OracleConfig, "REGEXP_LIKE(username, :1)"},
		{"MsSqlServer", sqlfilter.MsSqlServerConfig, "username LIKE @p1"},
		{"SQLite", sqlfilter.SQLiteConfig, "username REGEXP ?"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filter := f.Like("username", "^admin.*", "regex")
			sf, err := sqlfilter.New(tt.cfg, filter)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if sf.String() != tt.expected {
				t.Errorf("got %q, want %q", sf.String(), tt.expected)
			}
			if len(sf.Args()) != 1 || sf.Args()[0] != "^admin.*" {
				t.Errorf("args mismatch: got %v", sf.Args())
			}
		})
	}
}

func TestIdentifierQuoting(t *testing.T) {
	filter := f.And(
		f.Eq("users.order", 100),
		f.In("group.status", "active", "pending"),
	)

	// Postgres quotes with " and uses $1, $2, $3
	pgCfg := sqlfilter.PostgresConfig.WithQuoting(true)
	sfPg, err := sqlfilter.New(pgCfg, filter)
	if err != nil {
		t.Fatal(err)
	}
	expectedPg := `("users"."order" = $1 and "group"."status" IN ($2,$3))`
	if sfPg.String() != expectedPg {
		t.Errorf("Postgres quoting: got %q, want %q", sfPg.String(), expectedPg)
	}
	expectedPgArgs := []any{100, "active", "pending"}
	if !reflect.DeepEqual(sfPg.Args(), expectedPgArgs) {
		t.Errorf("Postgres args mismatch: got %v, want %v", sfPg.Args(), expectedPgArgs)
	}

	// MySQL quotes with ` and uses ?
	myCfg := sqlfilter.MySqlConfig.WithQuoting(true)
	sfMy, err := sqlfilter.New(myCfg, filter)
	if err != nil {
		t.Fatal(err)
	}
	expectedMy := "(`users`.`order` = ? and `group`.`status` IN (?,?))"
	if sfMy.String() != expectedMy {
		t.Errorf("MySQL quoting: got %q, want %q", sfMy.String(), expectedMy)
	}

	// SQL Server quotes with [ ] and uses @p1, @p2, @p3
	msCfg := sqlfilter.MsSqlServerConfig.WithQuoting(true)
	sfMs, err := sqlfilter.New(msCfg, filter)
	if err != nil {
		t.Fatal(err)
	}
	expectedMs := "([users].[order] = @p1 and [group].[status] IN (@p2,@p3))"
	if sfMs.String() != expectedMs {
		t.Errorf("SQL Server quoting: got %q, want %q", sfMs.String(), expectedMs)
	}
}

func TestNumericBooleans(t *testing.T) {
	valPg, err := sqlfilter.ParseWithConfig(sqlfilter.PostgresConfig, true)
	if err != nil || valPg != "TRUE" {
		t.Errorf("Postgres bool: got %q, want TRUE", valPg)
	}

	valMs, err := sqlfilter.ParseWithConfig(sqlfilter.MsSqlServerConfig, true)
	if err != nil || valMs != "1" {
		t.Errorf("SQL Server bool: got %q, want 1", valMs)
	}

	valOra, err := sqlfilter.ParseWithConfig(sqlfilter.OracleConfig, true)
	if err != nil || valOra != "1" {
		t.Errorf("Oracle bool: got %q, want 1", valOra)
	}
}

func TestParseTypes(t *testing.T) {
	// bool
	val, err := sqlfilter.Parse(true)
	if err != nil || val != "TRUE" {
		t.Errorf("Parse(true) = %q, %v; want TRUE", val, err)
	}
	val, err = sqlfilter.Parse(false)
	if err != nil || val != "FALSE" {
		t.Errorf("Parse(false) = %q, %v; want FALSE", val, err)
	}

	// nil
	val, err = sqlfilter.Parse(nil)
	if err != nil || val != "NULL" {
		t.Errorf("Parse(nil) = %q, %v; want NULL", val, err)
	}

	// float32 precision
	val, err = sqlfilter.Parse(float32(3.14159))
	if err != nil || !strings.HasPrefix(val, "3.14159") {
		t.Errorf("Parse(float32) = %q, %v", val, err)
	}

	// string escaping
	val, err = sqlfilter.Parse("O'Reilly")
	if err != nil || val != "'O''Reilly'" {
		t.Errorf("Parse string escaping = %q, %v; want 'O''Reilly'", val, err)
	}
}

func TestEmptyInFilter(t *testing.T) {
	filter := f.In("id")
	_, err := sqlfilter.New(sqlfilter.PostgresConfig, filter)
	if err == nil {
		t.Fatal("expected error for empty IN filter, got nil")
	}
}

func TestSqlFilterLogging(t *testing.T) {
	var buf strings.Builder
	logger := slog.New(slog.NewTextHandler(&buf, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))

	cfg := sqlfilter.PostgresConfig.WithLogger(logger)
	filter := f.And(
		f.Eq("status", "active"),
		f.Between("age", 18, 65),
	)

	_, err := sqlfilter.New(cfg, filter)
	if err != nil {
		t.Fatal(err)
	}

	logOutput := buf.String()
	if !strings.Contains(logOutput, "level=INFO") || !strings.Contains(logOutput, "compiling filter to SQL") {
		t.Errorf("expected INFO log in output, got: %s", logOutput)
	}
	if !strings.Contains(logOutput, "level=DEBUG") || !strings.Contains(logOutput, "visiting comparison") {
		t.Errorf("expected DEBUG log in output, got: %s", logOutput)
	}
}
