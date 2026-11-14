
[![Go Reference](https://pkg.go.dev/badge/github.com/XPD-Kasun/filters.svg)](https://pkg.go.dev/github.com/XPD-Kasun/filters)
[![License: MIT](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)

# filters

> TL;DR: **A database agnostic filtering API in GO** for your persistance needs.

Filters is a Go library that consist of a filter query design API and a Go file generator for your entity structs. Describe **what** you want to find as a plain Go value, and let a separate consumer (what we call a *visitor*) decide **how** to run it. Make your application independent of underlying filtering logic and allow your filtering logic to switch between various SQL providers and NoSQL when it's ready.
 
```go
filter := f.And(
    f.Eq("status", "active"),
    f.Gte("age", 18),
)
// Pass `filter` to your repository. It never needs to know about SQL.
```
---
 
## Table of contents
 
- [Installation](#installation)
- [Quick start](#quick-start)
- [Why filters? An Introduction](#why-filters)
- [Core API](#core-api)
- [Rendering filters with a visitor](#rendering-filters-with-a-visitor)
- [Writing your own visitor](#writing-your-own-visitor)
- [Strongly-typed filters for your entities](#strongly-typed-filters-for-your-entities)
- [Attaching metadata to a filter](#attaching-metadata-to-a-filter)
- [Coming Up](#coming-up)
- [Contributing](#contributing)
- [License](#license)
---
 
## Installation

```bash
go get github.com/XPD-Kasun/filters
```
 
Import it with a short alias:
 
```go
import f "github.com/XPD-Kasun/filters"
```

## Quick Start

**1. Build a filter.**
 
```go
import f "github.com/XPD-Kasun/filters"
 
someFilter := f.And(
    f.Eq("orderName", "order-23434"),
    f.Or(
        f.And(
            f.Eq("customerName", "john"),
            f.Gte("age", 34),
            f.Gte("regDate", "2026-02-23"),
        ),
        f.Eq("priority", "urgent"),
    ),
)
```
 
**2. Render it with a visitor.** Here we use the bundled SQL visitor:
 
```go
import (
    f "github.com/XPD-Kasun/filters"
    "github.com/XPD-Kasun/filters/sqlfilter"
)
 
sql := sqlfilter.New(sqlfilter.PostgresConfig, someFilter)
fmt.Println(sql)
```
 
<!-- TODO: show the exact output of the call above, and whether New returns a string, a struct with args, or an error. -->
 
The filter itself contains no SQL. Swap the visitor and the same filter can produce a different query language.

## Why filters? An Introduction

A filter is a predicate that describe a certain criteria. Filters are used everywhere whenever you are writing a SQL query with a where clause. This library helps you to decouple your criteria specification from its implementation. THis is specially useful for hexagonal architecture based applications that incorporate Domain Driven Design (DDD) concepts.

For example, lets consider a CRM system where a complex application service invokes a certain REST endpoint that is given with set of criterias. If this set is directly passed to the persistant layer we introduce a coupling of the REST arguments to the hexagonal driving port like SQL/ORM layer. Instead of directly coupling we can only pass a single `Filter` instance to the port. This allows 

- The port implementation can evolve independently.
- You can swap persistence providers (Oracle to Postgres, for example) without touching the domain.
- You can target non-relational stores such as MongoDB with a different visitor.
- Filters are plain data, so they are easy to build, test, log and compare.
- You can create a custom provider by implementing the `visitor` interface and use the same filter down the stream. You can use this for observability purposes. 
- Filter creator can inject metadata that can be consumed by the visitor. This allows to provide more information such as tracing to the visitor.

## Core API


All constructors live in the root package and return a `Filter`.
 
### Comparison
 
| Function | Meaning | Example |
|---|---|---|
| `Eq(field, value)` | equal to | `f.Eq("name", "john")` |
| `Nq(field, value)` | not equal to | `f.Nq("status", "closed")` |
| `Gt(field, value)` | greater than | `f.Gt("age", 17)` |
| `Gte(field, value)` | greater than or equal to | `f.Gte("age", 18)` |
| `Lt(field, value)` | less than | `f.Lt("price", 100)` |
| `Lte(field, value)` | less than or equal to | `f.Lte("price", 100)` |
 
### Logical
 
| Function | Meaning | Example |
|---|---|---|
| `And(filters...)` | all children must match | `f.And(a, b, c)` |
| `Or(filters...)` | at least one child must match | `f.Or(a, b)` |
| `Not(filter)` | negates a single child | `f.Not(a)` |
 
Logical filters nest to any depth.
 
### Pattern, set and range
 
| Function | Meaning | Example |
|---|---|---|
| `Like(field, value, method)` | pattern match | `f.Like("name", "jo", method)` |
| `In(field, values...)` | value is one of a set | `f.In("country", "LK", "IN", "SG")` |
| `Between(field, from, to)` | value falls within a range | `f.Between("age", 18, 65)` |
 
<!-- TODO: document the accepted values for Like's `method` argument (for example prefix / suffix / contains). -->
 
### Putting it together
 
```go
active := f.And(
    f.Eq("status", "active"),
    f.Not(f.In("country", "XX", "YY")),
    f.Between("age", 18, 65),
    f.Or(
        f.Like("email", "%@example.com"),
        f.Gte("loyaltyPoints", 1000),
    ),
)
```

## Rendering filters with a visitor
 
Every filter implements:
 
```go
type Filter interface {
    Accept(Visitor) error
}
```
 
A **visitor** walks the filter tree and translates each node into something concrete. This is the classic Visitor pattern, and it is what keeps criteria separate from implementation.
 
Visitors available in this repository:
 
| Package | Target |
|---|---|
| `sqlfilter` | SQL (PostgreSQL configuration provided via `sqlfilter.PostgresConfig`) |
 
## Writing your own visitor
 
To support another backend, implement the `Visitor` interface:
 
```go
type Visitor interface {
    VisitLogical(LogicalFilter) error
    VisitComparison(CompareFilter) error
    VisitLike(LikeFilter) error
    VisitBetween(BetweenFilter) error
    VisitIn(InFilter) error
}
```
 
Each method receives a concrete node type:
 
| Node | Fields |
|---|---|
| `LogicalFilter` | `Op op.LogicalOperator`, `Children []Filter` |
| `CompareFilter` | `Op op.Operator`, `Field string`, `Value any` |
| `LikeFilter` | `Field string`, `Value string`, `Method string` |
| `InFilter` | `Field string`, `Values []any` |
| `BetweenFilter` | `Field string`, `From any`, `To any` |
 
A minimal sketch of a visitor that renders a filter as a readable string:
 
```go
type debugVisitor struct {
    sb strings.Builder
}
 
func (d *debugVisitor) VisitComparison(c f.CompareFilter) error {
    fmt.Fprintf(&d.sb, "%s %v %v", c.Field, c.Op, c.Value)
    return nil
}
 
func (d *debugVisitor) VisitBetween(b f.BetweenFilter) error {
    fmt.Fprintf(&d.sb, "%s between %v and %v", b.Field, b.From, b.To)
    return nil
}
 
func (d *debugVisitor) VisitIn(i f.InFilter) error {
    fmt.Fprintf(&d.sb, "%s in %v", i.Field, i.Values)
    return nil
}
 
func (d *debugVisitor) VisitLike(l f.LikeFilter) error {
    fmt.Fprintf(&d.sb, "%s like %q", l.Field, l.Value)
    return nil
}
 
func (d *debugVisitor) VisitLogical(l f.LogicalFilter) error {
    d.sb.WriteString("(")
    for i, child := range l.Children {
        if i > 0 {
            fmt.Fprintf(&d.sb, " %v ", l.Op)
        }
        if err := child.Accept(d); err != nil {
            return err
        }
    }
    d.sb.WriteString(")")
    return nil
}
```
 
Usage:
 
```go
v := &debugVisitor{}
if err := someFilter.Accept(v); err != nil {
    log.Fatal(err)
}
fmt.Println(v.sb.String())
```
 
The same approach works for building a MongoDB `bson.M`, an Elasticsearch query, or an in-memory predicate for tests.
 

## Strongly typed filters for your entities

We often work with data containers like structs representing Entities and string field names are easy to mistype. For entity structs, you can use the bundled generator, `gen-filters`, to produce typed query structs so the compiler catches mistakes.

**1. Annotate your entity and add a `go:generate` directive.**

```go
import (
    f "https://github.com/XPD-Kasun/filters"
)

//go:generate gen-filters customer-query.go
type AppService struct {
    myPort SomePort
}

//genQuery
type Customer struct{
    Name, ssn string
    age int
}

func (c *Customer) Ssn() string {
    return c.ssn
}

func (c *Customer) Age() int {
    return c.age
}

func (a *AppService) GetCustomers(name, ssn string, age int) ([]*Customer, error) {
    custFilter := f.And(
        customerFilter.Name.Eq(name),
        customerFilter.Ssn.Eq(ssn),
        customerFilter.Age.GTE(18)
    )
    return a.myPort.loadCustomers(custFilter)
}
```

**2. Run the generator.**
 
```bash
go install github.com/XPD-Kasun/filters/cmd/gen-filters@latest
go generate ./...
```


**3. Use the generated filter.**
 
```go
func (a *AppService) GetCustomers(name, ssn string) ([]*Customer, error) {
    filter := f.And(
        customerFilter.Name.Eq(name),
        customerFilter.Ssn.Eq(ssn),
        customerFilter.Age.Gte(18),
    )
    return a.myPort.LoadCustomers(filter)
}
```
 
<!-- TODO: confirm the generated variable name (customerFilter), whether methods are Gte or GTE, and how unexported fields exposed through getters are handled. Add a snippet of the generated file. -->


## Attaching metadata to a filter
 
Sometimes a filter needs to carry extra instructions that are not part of the criteria itself, such as a table alias, a tenant hint or a sort hint. `WithMeta` wraps any filter with string key/value pairs:
 
```go
filter := f.WithMeta(
    map[string]string{"table": "customers"},
    f.Eq("name", "john"),
)
 
table := filter.Get("table")   // "customers"
all := filter.GetAll()         // a copy of the metadata map
```
 
A `MetaFilter` is still a normal `Filter`, so it can be passed to any visitor. Visitors may read the metadata to adjust their output.

## Coming up

Currently, the only visitor implementation is SQL provider. Future we are hoping to add more visitors. You are welcome to contribute more visitor implementations.

|Visitor|State|
|---|---|
|Sql Visitor|Completed|
|In Memory Visitor|Soon Announced|
|MongoDB Visitor|To be done|
|GORM Visitor|To be done|

## Contributing
 
Contributions are welcome. To get started:
 
1. Fork the repository and create a feature branch.
2. Make your change and add tests (`go test ./...`).
3. Open a pull request describing what you changed and why.
Ideas that would be especially valuable: new visitors (MongoDB, Elasticsearch, in-memory), more SQL dialects, and documentation improvements.
 
## License
 
Released under the [MIT License](LICENSE).
