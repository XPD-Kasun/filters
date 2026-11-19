package filters_test

import (
	"reflect"
	"testing"

	f "github.com/XPD-Kasun/filters"
)

func TestExtractEntities(t *testing.T) {
	// 1. Nil filter
	entities, err := f.ExtractEntities(nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(entities) != 0 {
		t.Errorf("expected 0 entities for nil filter, got %v", entities)
	}

	// 2. Simple filter with no entity prefix (just column name)
	simple := f.Eq("price", 100)
	entities, err = f.ExtractEntities(simple)
	if err != nil {
		t.Fatal(err)
	}
	if len(entities) != 0 {
		t.Errorf("expected 0 entities for simple column name, got %v", entities)
	}

	// 3. Single entity path
	single := f.Eq("users.id", 1)
	entities, err = f.ExtractEntities(single)
	if err != nil {
		t.Fatal(err)
	}
	expectedSingle := []string{"users"}
	if !reflect.DeepEqual(entities, expectedSingle) {
		t.Errorf("got %v, want %v", entities, expectedSingle)
	}

	// 4. Multi-level path: "order.customer.price" -> ["order", "customer"]
	multi := f.Gte("order.customer.price", 200)
	entities, err = f.ExtractEntities(multi)
	if err != nil {
		t.Fatal(err)
	}
	expectedMulti := []string{"order", "customer"}
	if !reflect.DeepEqual(entities, expectedMulti) {
		t.Errorf("got %v, want %v", entities, expectedMulti)
	}

	// 5. Complex logical tree with duplicates across different filter types
	complexFilter := f.And(
		f.Eq("order.customer.price", 200),
		f.Between("order.item.quantity", 1, 10),
		f.In("customer.status", "active", "vip"),
		f.StringMatch("warehouse.location.code", "US-%", f.WildcardMatch),
		f.Or(
			f.Eq("order.id", 55),
			f.Eq("customer.email", "test@example.com"),
		),
	)
	entities, err = f.ExtractEntities(complexFilter)
	if err != nil {
		t.Fatal(err)
	}

	// Expected order of first discovery: order, customer, item, warehouse, location
	expectedComplex := []string{"order", "customer", "item", "warehouse", "location"}
	if !reflect.DeepEqual(entities, expectedComplex) {
		t.Errorf("got %v, want %v", entities, expectedComplex)
	}

	// 6. Not filter
	notFilter := f.Not(f.Eq("user.profile.age", 30))
	notEntities, err := f.ExtractEntities(notFilter)
	if err != nil {
		t.Fatal(err)
	}
	expectedNot := []string{"user", "profile"}
	if !reflect.DeepEqual(notEntities, expectedNot) {
		t.Errorf("got %v, want %v", notEntities, expectedNot)
	}
}
