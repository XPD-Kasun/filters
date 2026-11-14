package main

import (
	"examples/product"
	"fmt"

	f "github.com/XPD-Kasun/filters"
	"github.com/XPD-Kasun/filters/sqlfilter"
)

func main() {
	proQ := product.NewProductFilter()
	q := f.And(proQ.Name.Eq("fn"), proQ.Price.Gt(100))
	fmt.Println(sqlfilter.New(sqlfilter.PostgresConfig, q))
}
