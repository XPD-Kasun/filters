package sqlfilter

import "testing"

func TestPathHelpers(t *testing.T) {

	path := "Order.customer.payment.id"

	t.Run("pathSegCount", func(tt *testing.T) {

		result := pathSegCount(path)

		if result != 4 {
			tt.Errorf("pathSegCount returns invalid path lentgh. got %d, expected %d", result, 4)
		}
	})

	t.Run("pathEntityAndField", func(tt *testing.T) {

		entity, field := pathEntityAndField(path)

		if entity != "payment" || field != "id" {
			tt.Errorf("returns wrong entity or field. got %s, %s, expected %s, %s", entity, field, "payment", "id")
		}
	})

	t.Run("pathEntitiesAndField", func(tt *testing.T) {

		entities, field := pathEntitiesAndField(path)

		if entities != "Order.customer.payment" || field != "id" {
			tt.Errorf(
				"returns wrong entity or field. got %s, %s, expected %s, %s",
				entities,
				field,
				"Order.customer.payment",
				"id",
			)
		}
	})

}
