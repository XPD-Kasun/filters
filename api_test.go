package filters_test

import (
	"errors"
	"log/slog"
	"reflect"
	"strconv"
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
	filter, err := sqlfilter.New(sqlfilter.UsePostgres(), s)
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
	Price q.Orderable[int, B]
	Name  q.StringField[B]
}

func TestFn(t *testing.T) {
	var a A = A{
		Price: q.NewOrderable[int, B]("price", ""),
		Name:  q.NewStringField[B]("Stesll", "df"),
	}
	s := f.Or(a.Price.Gte(200), a.Name.Eq("fnsdf"))
	w := sqlfilter.NewEmpty(sqlfilter.UsePostgres())
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
		Price: q.NewOrderable[int, B]("price", ""),
		Name:  q.NewStringField[B]("Stesll", "df"),
	}
	s := a.Price.Gte(200)
	w := sqlfilter.NewEmpty(sqlfilter.UsePostgres())
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
	sf, err := sqlfilter.New(sqlfilter.UsePostgres(), s)
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
	sfEq, err := sqlfilter.New(sqlfilter.UsePostgres(), filterEq)
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
	filterNq := f.Neq("deleted_at", nil)
	sfNq, err := sqlfilter.New(sqlfilter.UsePostgres(), filterNq)
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
		{"Postgres", sqlfilter.UsePostgres(), "username ~ $1"},
		{"CockroachDB", sqlfilter.UseCockroachDB(), "username ~ $1"},
		{"MySQL", sqlfilter.UseMySQL(), "username REGEXP ?"},
		{"ClickHouse", sqlfilter.UseClickHouse(), "match(username, ?)"},
		{"Oracle", sqlfilter.UseOracle(), "REGEXP_LIKE(username, :1)"},
		{"MsSqlServer", sqlfilter.UseMsSqlServer(), "username LIKE @p1"},
		{"SQLite", sqlfilter.UseSQLite(), "username REGEXP ?"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filter := f.StringMatch("username", "^admin.*", f.Regex, true)
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
	pgCfg := sqlfilter.UsePostgres().WithQuoting(sqlfilter.QuoteAlways)
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
	myCfg := sqlfilter.UseMySQL().WithQuoting(sqlfilter.QuoteAlways)
	sfMy, err := sqlfilter.New(myCfg, filter)
	if err != nil {
		t.Fatal(err)
	}
	expectedMy := "(`users`.`order` = ? and `group`.`status` IN (?,?))"
	if sfMy.String() != expectedMy {
		t.Errorf("MySQL quoting: got %q, want %q", sfMy.String(), expectedMy)
	}

	// SQL Server quotes with [ ] and uses @p1, @p2, @p3
	msCfg := sqlfilter.UseMsSqlServer().WithQuoting(sqlfilter.QuoteAlways)
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
	valPg, err := sqlfilter.ParseWithConfig(sqlfilter.UsePostgres(), true)
	if err != nil || valPg != "TRUE" {
		t.Errorf("Postgres bool: got %q, want TRUE", valPg)
	}

	valMs, err := sqlfilter.ParseWithConfig(sqlfilter.UseMsSqlServer(), true)
	if err != nil || valMs != "1" {
		t.Errorf("SQL Server bool: got %q, want 1", valMs)
	}

	valOra, err := sqlfilter.ParseWithConfig(sqlfilter.UseOracle(), true)
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
	_, err := sqlfilter.New(sqlfilter.UsePostgres(), filter)
	if err == nil {
		t.Fatal("expected error for empty IN filter, got nil")
	}
}

func TestSqlFilterLogging(t *testing.T) {
	var buf strings.Builder
	logger := slog.New(slog.NewTextHandler(&buf, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))

	cfg := sqlfilter.UsePostgres().WithLogger(logger)
	filter := f.And(
		f.Eq("status", "active"),
		f.Gt("priority", 1),
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
	if !strings.Contains(logOutput, "visiting string match") {
		t.Errorf("expected string match DEBUG log in output, got: %s", logOutput)
	}
}

func TestEntityFieldMapping(t *testing.T) {
	fieldMappings := map[string]string{
		"order":    "o",
		"customer": "c",
	}

	// 1. Comparison with unquoted fields: order.customer.price -> o.c.price
	cfgUnquoted := sqlfilter.UsePostgres().WithFields(fieldMappings)
	filter1 := f.Gte("order.customer.price", 200)
	sf1, err := sqlfilter.New(cfgUnquoted, filter1)
	if err != nil {
		t.Fatal(err)
	}
	if sf1.String() != "o.c.price >= $1" {
		t.Errorf("got %q, want %q", sf1.String(), "o.c.price >= $1")
	}

	// 2. Comparison with quoted fields: order.customer.price -> "o"."c"."price"
	cfgQuoted := sqlfilter.UsePostgres().WithQuoting(sqlfilter.QuoteAlways).WithFields(fieldMappings)
	sf2, err := sqlfilter.New(cfgQuoted, filter1)
	if err != nil {
		t.Fatal(err)
	}
	expectedQuoted := `"o"."c"."price" >= $1`
	if sf2.String() != expectedQuoted {
		t.Errorf("got %q, want %q", sf2.String(), expectedQuoted)
	}

	// 3. Across Between, In, Like, and IS NULL
	complexFilter := f.And(
		f.Between("order.customer.price", 100, 500),
		f.In("order.customer.status", "active", "vip"),
		f.StringMatch("order.customer.name", "Smith%", f.WildcardMatch, true),
		f.Eq("order.customer.deleted_at", nil),
	)
	sf3, err := sqlfilter.New(cfgQuoted, complexFilter)
	if err != nil {
		t.Fatal(err)
	}
	expectedComplex := `("o"."c"."price" BETWEEN $1 and $2 and "o"."c"."status" IN ($3,$4) and "o"."c"."name" LIKE $5 and "o"."c"."deleted_at" IS NULL)`
	if sf3.String() != expectedComplex {
		t.Errorf("got %q, want %q", sf3.String(), expectedComplex)
	}

	// 4. Case-insensitive segment matching (Order.Customer.Price -> o.c.Price)
	filterCase := f.Eq("Order.Customer.Price", 300)
	sf4, err := sqlfilter.New(cfgUnquoted, filterCase)
	if err != nil {
		t.Fatal(err)
	}
	if sf4.String() != "o.c.Price = $1" {
		t.Errorf("got %q, want o.c.Price = $1", sf4.String())
	}

	// 5. Unmapped fields remain unchanged
	filterUnmapped := f.Eq("simple_field", 42)
	sf5, err := sqlfilter.New(cfgUnquoted, filterUnmapped)
	if err != nil {
		t.Fatal(err)
	}
	if sf5.String() != "simple_field = $1" {
		t.Errorf("got %q, want simple_field = $1", sf5.String())
	}
}

// TestQuotingEscaping verifies that closing-character doubling is applied per
// dialect and that Oracle rejects double-quotes inside identifiers.
func TestQuotingEscaping(t *testing.T) {
	// Postgres: double-quote inside identifier → doubled.
	// Field name: say`me → `say``me` with backtick style (MySQL)
	filter := f.Eq("say`field", 1)
	myCfg := sqlfilter.UseMySQL().WithQuoting(sqlfilter.QuoteAlways)
	sf, err := sqlfilter.New(myCfg, filter)
	if err != nil {
		t.Fatal(err)
	}
	// closing backtick should be doubled
	if sf.String() != "`say``field` = ?" {
		t.Errorf("MySQL escaping: got %q", sf.String())
	}
	// SQL Server: ] inside identifier → ]]
	filter2 := f.Eq("col]umn", 1)
	msCfg := sqlfilter.UseMsSqlServer().WithQuoting(sqlfilter.QuoteAlways)
	sf2, err := sqlfilter.New(msCfg, filter2)
	if err != nil {
		t.Fatal(err)
	}
	if sf2.String() != "[col]]umn] = @p1" {
		t.Errorf("SQL Server ] escaping: got %q", sf2.String())
	}

	// Oracle: double-quote inside identifier → error
	filter3 := f.Eq(`col"umn`, 1)
	oraCfg := sqlfilter.UseOracle().WithQuoting(sqlfilter.QuoteAlways)
	_, err3 := sqlfilter.New(oraCfg, filter3)
	if err3 == nil {
		t.Error("Oracle: expected error for double-quote in identifier, got nil")
	}
}

// TestStartAt verifies that parameter numbering can be offset for multi-fragment queries.
func TestStartAt(t *testing.T) {
	filter := f.And(f.Eq("a", 1), f.Eq("b", 2))
	sf, err := sqlfilter.New(sqlfilter.UsePostgres(), filter, sqlfilter.StartAt(3))
	if err != nil {
		t.Fatal(err)
	}
	want := "(a = $4 and b = $5)"
	if sf.String() != want {
		t.Errorf("StartAt(3): got %q, want %q", sf.String(), want)
	}
}

// TestIdentifierLengthLimit verifies that identifiers exceeding the dialect
// byte limit produce an error rather than a silent truncation.
func TestIdentifierLengthLimit(t *testing.T) {
	longName := strings.Repeat("x", 64) // Postgres limit is 63
	filter := f.Eq(longName, 1)
	_, err := sqlfilter.New(sqlfilter.UsePostgres().WithQuoting(sqlfilter.QuoteAlways), filter)
	if err == nil {
		t.Error("expected error for identifier exceeding Postgres 63-byte limit, got nil")
	}
}

// TestWithPlaceholder verifies that a custom PlaceholderFunc overrides the default.
func TestWithPlaceholder(t *testing.T) {
	custom := sqlfilter.PlaceholderFunc(func(order int) string {
		return "{{" + strconv.Itoa(order) + "}}"
	})
	cfg := sqlfilter.UsePostgres().WithPlaceholder(custom)
	sf, err := sqlfilter.New(cfg, f.Eq("x", 42))
	if err != nil {
		t.Fatal(err)
	}
	if sf.String() != "x = {{1}}" {
		t.Errorf("custom placeholder: got %q", sf.String())
	}
}

// TestNilChildrenAndEmptyFilter verifies that nil / empty filters are discarded
// and do not cause nil panics or malformed SQL.
func TestNilChildrenAndEmptyFilter(t *testing.T) {
	// And with nil elements
	filterAnd := f.And(nil, f.Eq("a", 1), f.EmptyFilter{}, f.Eq("b", 2))
	sfAnd, err := sqlfilter.New(sqlfilter.UsePostgres(), filterAnd)
	if err != nil {
		t.Fatal(err)
	}
	wantAnd := "(a = $1 and b = $2)"
	if sfAnd.String() != wantAnd {
		t.Errorf("got %q, want %q", sfAnd.String(), wantAnd)
	}

	// All nil in And -> EmptyFilter
	allNilAnd := f.And(nil, nil)
	sfEmpty, err := sqlfilter.New(sqlfilter.UsePostgres(), allNilAnd)
	if err != nil {
		t.Fatal(err)
	}
	if sfEmpty.String() != "" {
		t.Errorf("expected empty SQL for all-nil And, got %q", sfEmpty.String())
	}

	// Not with nil
	notNil := f.Not(nil)
	sfNotNil, err := sqlfilter.New(sqlfilter.UsePostgres(), notNil)
	if err != nil {
		t.Fatal(err)
	}
	if sfNotNil.String() != "" {
		t.Errorf("expected empty SQL for Not(nil), got %q", sfNotNil.String())
	}
}

// TestErrUnsupportedSentinel verifies that unsupported operations return
// an error wrapping filters.ErrUnsupported.
func TestErrUnsupportedSentinel(t *testing.T) {
	// 1. Unsupported StringMatch method
	filterUnknown := f.StringMatch("name", "foo", f.StringMatchMethod("custom_unknown"))
	_, err := sqlfilter.New(sqlfilter.UsePostgres(), filterUnknown)
	if !errors.Is(err, f.ErrUnsupported) {
		t.Errorf("expected ErrUnsupported for unknown StringMatch method, got %v", err)
	}

	// 2. Unsupported operator with nil (e.g. Gt with nil)
	filterGtNil := f.Gt("age", nil)
	_, err = sqlfilter.New(sqlfilter.UsePostgres(), filterGtNil)
	if !errors.Is(err, f.ErrUnsupported) {
		t.Errorf("expected ErrUnsupported for Gt with nil, got %v", err)
	}

	// 3. Identifier length exceeded
	longName := strings.Repeat("z", 70)
	filterLong := f.Eq(longName, 1)
	_, err = sqlfilter.New(sqlfilter.UsePostgres().WithQuoting(sqlfilter.QuoteAlways), filterLong)
	if !errors.Is(err, f.ErrUnsupported) {
		t.Errorf("expected ErrUnsupported for identifier length limit, got %v", err)
	}
}
