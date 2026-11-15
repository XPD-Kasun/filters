package main

import (
	"examples/product"
	"fmt"

	f "github.com/XPD-Kasun/filters"
	"github.com/XPD-Kasun/filters/sqlfilter"
)

func main() {
	proQ := product.NewProductFilter()
	q := f.And(proQ.Name.Eq("fn"), proQ.Price.Gt(100), proQ.Someval.IsNull())
	fmt.Println(sqlfilter.New(sqlfilter.PostgresConfig, q))
	fmt.Println(sqlfilter.New(sqlfilter.MsSqlServerConfig, q))
	fmt.Println(sqlfilter.New(sqlfilter.OracleConfig, q))
}
