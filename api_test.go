package filters_test

import (
	"testing"

	f "github.com/XPD-Kasun/filters"
	"github.com/XPD-Kasun/filters/q"
	"github.com/XPD-Kasun/filters/sqlfilter"
)

func TestBasic(t *testing.T) {
	s := f.OR(
		f.AND(f.EQ("se", 12), f.GT("fsd", 2432)),
		f.EQ("thx", "success"),
		f.BETWEEN("fdsi", 10, 200),
		f.BETWEEN("fdsi23", 10, 200),
	)
	filter, err := sqlfilter.New(s)
	if err != nil {
		t.Error(err.Error())
		println("err")
		return
	}
	t.Log("df", filter.String())

}

type B struct {
}

type A struct {
	Price q.IntField[B]
	Name  q.StringField[B]
}

func TestFn(t *testing.T) {
	var a A
	s := f.OR(a.Price.Gte(200), a.Name.Eq("fnsdf"))
	w := sqlfilter.NewEmpty()
	s.Accept(w)
	t.Log(w.String())
}
