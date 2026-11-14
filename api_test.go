package filters_test

import (
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
	expected := "((se = 12 and fsd > 2432) or thx = 'success' or fdsi BETWEEN $1 and $2 or fdsi23 BETWEEN $3 and $4)"
	if filter.String() != expected {
		t.Errorf("got %q, want %q", filter.String(), expected)
	}
	if len(filter.Args()) != 4 {
		t.Errorf("expected 4 args, got %d", len(filter.Args()))
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
	expected := "(price >= 200 or df.Stesll = 'fnsdf')"
	if w.String() != expected {
		t.Errorf("got %q, want %q", w.String(), expected)
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
	expected := "price >= 200"
	if w.String() != expected {
		t.Errorf("got %q, want %q", w.String(), expected)
	}
}

func TestNotFilter(t *testing.T) {
	s := f.Not(f.Eq("status", "deleted"))
	sf, err := sqlfilter.New(sqlfilter.PostgresConfig, s)
	if err != nil {
		t.Fatal(err)
	}
	expected := "NOT(status = 'deleted')"
	if sf.String() != expected {
		t.Errorf("got %q, want %q", sf.String(), expected)
	}
}

func TestRegexDialects(t *testing.T) {
	tests := []struct {
		name     string
		cfg      sqlfilter.SqlConfig
		expected string
	}{
		{"Postgres", sqlfilter.PostgresConfig, "username ~ $1"},
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

	// Postgres quotes with "
	pgCfg := sqlfilter.PostgresConfig.WithQuoting(true)
	sfPg, err := sqlfilter.New(pgCfg, filter)
	if err != nil {
		t.Fatal(err)
	}
	expectedPg := `("users"."order" = 100 and "group"."status" IN ('active','pending'))`
	if sfPg.String() != expectedPg {
		t.Errorf("Postgres quoting: got %q, want %q", sfPg.String(), expectedPg)
	}

	// MySQL quotes with `
	myCfg := sqlfilter.MySqlConfig.WithQuoting(true)
	sfMy, err := sqlfilter.New(myCfg, filter)
	if err != nil {
		t.Fatal(err)
	}
	expectedMy := "(`users`.`order` = 100 and `group`.`status` IN ('active','pending'))"
	if sfMy.String() != expectedMy {
		t.Errorf("MySQL quoting: got %q, want %q", sfMy.String(), expectedMy)
	}

	// SQL Server quotes with [ ]
	msCfg := sqlfilter.MsSqlServerConfig.WithQuoting(true)
	sfMs, err := sqlfilter.New(msCfg, filter)
	if err != nil {
		t.Fatal(err)
	}
	expectedMs := "([users].[order] = 100 and [group].[status] IN ('active','pending'))"
	if sfMs.String() != expectedMs {
		t.Errorf("SQL Server quoting: got %q, want %q", sfMs.String(), expectedMs)
	}
}

func TestNumericBooleans(t *testing.T) {
	filter := f.Eq("is_active", true)

	// Postgres uses TRUE
	sfPg, err := sqlfilter.New(sqlfilter.PostgresConfig, filter)
	if err != nil {
		t.Fatal(err)
	}
	if sfPg.String() != "is_active = TRUE" {
		t.Errorf("Postgres bool: got %q, want is_active = TRUE", sfPg.String())
	}

	// SQL Server uses 1/0
	sfMs, err := sqlfilter.New(sqlfilter.MsSqlServerConfig, filter)
	if err != nil {
		t.Fatal(err)
	}
	if sfMs.String() != "is_active = 1" {
		t.Errorf("SQL Server bool: got %q, want is_active = 1", sfMs.String())
	}

	// Oracle uses 1/0
	sfOra, err := sqlfilter.New(sqlfilter.OracleConfig, filter)
	if err != nil {
		t.Fatal(err)
	}
	if sfOra.String() != "is_active = 1" {
		t.Errorf("Oracle bool: got %q, want is_active = 1", sfOra.String())
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
