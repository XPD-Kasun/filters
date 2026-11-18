package fields

import (
	"cmp"
	"time"

	"github.com/XPD-Kasun/filters"
	f "github.com/XPD-Kasun/filters"
)

// Ordered type interface described a generic type that can be filtered using >=, >,
// <, <== and between.
type Ordered interface {
	cmp.Ordered | time.Time
}

// Equalable describes a type that can be checked for ==, != and in operators.
// T is any comparable type and U is the entity type.
type Equalable[T comparable, U any] struct {
	Name  string
	Alias string
}

func (ef Equalable[T, U]) getAliasName() string {
	if ef.Alias != "" {
		return ef.Alias + "." + ef.Name
	}
	return ef.Name
}

// Returns a [filters.Filter] that checks whether the given value of type T
// equals this field's value. If T is pointer type, then v can be nil. In this case,
// a visitor can use `is_null` semantics if available.
func (ef Equalable[T, U]) Eq(v T) f.Filter {
	return f.Eq(ef.getAliasName(), v)
}

// Returns a [filters.Filter] that checks whether the given value of type T
// does not equal this field's value. If T is pointer type, then v can be nil. In this case,
// a visitor can use `is_not_null` semantics if available.
func (ef Equalable[T, U]) Neq(v T) f.Filter {
	return f.Neq(ef.getAliasName(), v)
}

// Returns a [filters.Filter] that checks whether this field's value
// equals to one of the given values for args.
func (ef Equalable[T, U]) In(args ...T) f.Filter {
	return f.In(ef.getAliasName(), args)
}

// Creates a new [Equalable].
func NewEqualable[T comparable, U any](name, alias string) Equalable[T, U] {
	return Equalable[T, U]{
		Name:  name,
		Alias: alias,
	}
}

// Orderable describes any type that can be ordered according to Ordered interface.
// T is any comparable type and U is the entity type.
type Orderable[T Ordered, U any] struct {
	Equalable[T, U]
}

// Returns a [filters.Filter] that checks whether this field's value
// greater than the given value v.
func (or Orderable[T, U]) Gt(v T) filters.Filter {
	return f.Gt(or.getAliasName(), v)
}

// Returns a [filters.Filter] that checks whether this field's value
// greater than or equals the given value v.
func (or Orderable[T, U]) Gte(v T) filters.Filter {
	return f.Gte(or.getAliasName(), v)
}

// Returns a [filters.Filter] that checks whether this field's value
// less than the given value v.
func (or Orderable[T, U]) Lt(v T) filters.Filter {
	return f.Lt(or.getAliasName(), v)
}

// Returns a [filters.Filter] that checks whether this field's value
// less than or equals the given value v.
func (or Orderable[T, U]) Lte(v T) filters.Filter {
	return f.Lte(or.getAliasName(), v)
}

// Returns a [filters.Filter] that checks whether this field's value
// in the range provided by a and b (a,b inclusive).
func (or Orderable[T, U]) Between(a, b T) filters.Filter {
	return f.Between(or.getAliasName(), a, b)
}

// Creates a new [Orderable].
func NewOrderable[T Ordered, U any](name, alias string) Orderable[T, U] {
	return Orderable[T, U]{
		Equalable: NewEqualable[T, U](name, alias),
	}
}

// Describes a Equalable type that can contain nil. For example pointer types.
//
//	eg: *int, *string
type NullableEqualable[T comparable, U any] struct {
	Equalable[T, U]
}

// Returns a [filters.Filter] that checks whether this field's value
// is nil. Visitor might use `is_null` semantics if available.
func (ne NullableEqualable[T, U]) IsNull() filters.Filter {
	return filters.Eq(ne.getAliasName(), nil)
}

// Returns a [filters.Filter] that checks whether this field's value
// is nil. Visitor might use `is_not_null` semantics if available.
func (ne NullableEqualable[T, U]) IsNotNull() filters.Filter {
	return filters.Neq(ne.getAliasName(), nil)
}

// Creates a new [NullableEqualable]
func NewNullableEqualable[T comparable, U any](name, alias string) NullableEqualable[T, U] {
	return NullableEqualable[T, U]{
		Equalable: NewEqualable[T, U](name, alias),
	}
}

// Describes a [Ordered] type that can contain nil values.
type NullableOrderable[T Ordered, U any] struct {
	Orderable[T, U]
}

// Returns a [filters.Filter] that checks whether this field's value
// is nil. Visitor might use `is_null` semantics if available.
func (no NullableOrderable[T, U]) IsNull() filters.Filter {
	return filters.Eq(no.getAliasName(), nil)
}

// Returns a [filters.Filter] that checks whether this field's value
// is nil. Visitor might use `is_not_null` semantics if available.
func (no NullableOrderable[T, U]) IsNotNull() filters.Filter {
	return filters.Neq(no.getAliasName(), nil)
}

// Creates a new [NullableOrderable]
func NewNullableOrderable[T Ordered, U any](name, alias string) NullableOrderable[T, U] {
	return NullableOrderable[T, U]{
		Orderable: NewOrderable[T, U](name, alias),
	}
}
