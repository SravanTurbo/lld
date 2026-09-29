package orderprocessor

import (
	"testing"
)

func TestProcessOrder(t *testing.T) {
	inventory := map[int]int{}

	op := NewOrderProcessor(inventory)

	op.addInventory(33,10)
	op.addInventory(44,12)

	orders := []*Order{
		{id:2, products: []Product{{44, 1},{33, 2}}}, //happy path
		{id:1, products: []Product{{22, 10},{33, 2}}}, //invalid product
		{id:3, products: []Product{{44, 20},{33, 2}}}, //invalid qty
		}

	op.processOrders(orders)

	for _, order := range orders {
		t.Logf("order: %+v", order)
	}


	op.addInventory(22, 19)
	op.addRelatedProducts(33, []int{44})
	op.addRelatedProducts(44, []int{22})

	t.Logf("state: %+v", op)
}