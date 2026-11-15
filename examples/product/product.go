package product

//go:generate go run github.com/XPD-Kasun/filters/cmd/gen-filters a.go

import "time"

//genQuery
type Product struct {
	id      int
	Name    string
	Price   float32
	Mfg     time.Time
	sku     string
	SomeVal *int
}

func (p *Product) Id() int     { return p.id }
func (p *Product) Sku() string { return p.sku }
