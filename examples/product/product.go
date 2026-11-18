package product

//go:generate go run github.com/XPD-Kasun/filters/cmd/gen-filters generated.go -debug

import (
	"time"
	"uuid"
)

//gen:query
type Product struct {
	// Private properties
	id          int
	sku         string
	cost        float32
	warehouseID uuid.UUID

	// Public properties
	Name        string
	Price       float32
	Mfg         time.Time
	SomeVal     *int
	Description string
	IsActive    bool
}

// Getters for private fields
func (p *Product) Id() int                { return p.id }
func (p *Product) Sku() string            { return p.sku }
func (p *Product) Cost() float32          { return p.cost }
func (p *Product) WarehouseID() uuid.UUID { return p.warehouseID }

//gen:query
type Order struct {
	// Private properties
	id            int
	customerID    string
	internalNotes string

	// Public properties
	Price     *int
	CreatedAt time.Time
	Status    string
}

// Getters for private fields
func (o *Order) Id() int               { return o.id }
func (o *Order) CustomerID() string    { return o.customerID }
func (o *Order) InternalNotes() string { return o.internalNotes }

// UserProfile represents a random entity with public fields and no methods.
//
//gen:query
type UserProfile struct {
	ID        int64
	Username  string
	Email     string
	Age       int
	IsActive  bool
	Balance   float64
	CreatedAt time.Time
	Tags      []string
	Metadata  map[string]interface{}
	Address   *Address
}

// Address is a nested public struct used within UserProfile.
//
//gen:query
type Address struct {
	Street  string
	City    string
	State   string
	ZipCode string
	Country string
}
