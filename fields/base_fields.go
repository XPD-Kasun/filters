package fields

import (
	"cmp"
	"time"

	"github.com/XPD-Kasun/filters"
	f "github.com/XPD-Kasun/filters"
)

type Ordered interface {
	cmp.Ordered | time.Time
}

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

func (ef Equalable[T, U]) Eq(v T) f.Filter {
	return f.Eq(ef.getAliasName(), v)
}

func (ef Equalable[T, U]) Nq(v T) f.Filter {
	return f.Nq(ef.getAliasName(), v)
}

func (ef Equalable[T, U]) In(args ...T) f.Filter {
	return f.In(ef.getAliasName(), args)
}

func NewEqualable[T comparable, U any](name, alias string) Equalable[T, U] {
	return Equalable[T, U]{
		Name:  name,
		Alias: alias,
	}
}

type Orderable[T Ordered, U any] struct {
	Equalable[T, U]
}

func (or Orderable[T, U]) Gt(v T) filters.Filter {
	return f.Gt(or.getAliasName(), v)
}

func (or Orderable[T, U]) Gte(v T) filters.Filter {
	return f.Gte(or.getAliasName(), v)
}

func (or Orderable[T, U]) Lt(v T) filters.Filter {
	return f.Lt(or.getAliasName(), v)
}

func (or Orderable[T, U]) Lte(v T) filters.Filter {
	return f.Lte(or.getAliasName(), v)
}

func (or Orderable[T, U]) Between(a, b T) filters.Filter {
	return f.Between(or.getAliasName(), a, b)
}

func NewOrderable[T Ordered, U any](name, alias string) Orderable[T, U] {
	return Orderable[T, U]{
		Equalable: NewEqualable[T, U](name, alias),
	}
}

type NullableEqualable[T comparable, U any] struct {
	Equalable[T, U]
}

func (ne NullableEqualable[T, U]) IsNull() filters.Filter {
	return filters.Eq(ne.getAliasName(), nil)
}

func (ne NullableEqualable[T, U]) IsNotNull() filters.Filter {
	return filters.Nq(ne.getAliasName(), nil)
}

func NewNullableEqualable[T comparable, U any](name, alias string) NullableEqualable[T, U] {
	return NullableEqualable[T, U]{
		Equalable: NewEqualable[T, U](name, alias),
	}
}

type NullableOrderable[T Ordered, U any] struct {
	Orderable[T, U]
}

func (no NullableOrderable[T, U]) IsNull() filters.Filter {
	return filters.Eq(no.getAliasName(), nil)
}

func (no NullableOrderable[T, U]) IsNotNull() filters.Filter {
	return filters.Nq(no.getAliasName(), nil)
}

func NewNullableOrderable[T Ordered, U any](name, alias string) NullableOrderable[T, U] {
	return NullableOrderable[T, U]{
		Orderable: NewOrderable[T, U](name, alias),
	}
}
