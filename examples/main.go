package main

import (
	"encoding/json"
	"examples/product"
	"fmt"
	"time"
	"uuid"

	f "github.com/XPD-Kasun/filters"
	"github.com/XPD-Kasun/filters/sqlfilter"
)

func main() {
	proQ := product.NewProductFilter()
	ordQ := product.NewOrderFilter()
	q := f.And(proQ.Name.Eq("fn"), proQ.Price.Gt(100), proQ.SomeVal.IsNull())
	q = f.Or(q, proQ.WarehouseID.Eq(uuid.New()), f.Not(ordQ.CreatedAt.Gte(time.Now())))
	q = f.And(q, f.Not(ordQ.Id.Between(10, 20)))

	sql1, err := sqlfilter.New(sqlfilter.UsePostgres(), q)
	if err != nil {
		fmt.Println("error at sql1 : ", err.Error())
	}
	fmt.Println(sql1.Sql())
	printArgs(sql1.Args())

	sql2, err := sqlfilter.New(sqlfilter.UsePostgres(), q)
	if err != nil {
		fmt.Println("error at sql1 : ", err.Error())
	}
	fmt.Println(sql2.Sql())
	printArgs(sql2.Args())

	sql3, err := sqlfilter.New(sqlfilter.UsePostgres(), q)
	if err != nil {
		fmt.Println("error at sql1 : ", err.Error())
	}
	fmt.Println(sql3.Sql())
	printArgs(sql3.Args())
}

func printArgs(args []any) {
	a, err := json.Marshal(args)
	if err != nil {
		fmt.Println("JSON marshal error at printArgs: ", err.Error())
		return
	}
	fmt.Println(string(a))
}
