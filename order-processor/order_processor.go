package orderprocessor

import (
	"errors"
)

type Orderprocessor struct {
	inventory map[int]int
	relations map[int][]int
}

type Order struct {
	id int
	products []Product
	fulfilled bool
}

type Product struct {
	id int
	quantity int
}

func NewOrderProcessor(inventory map[int]int) *Orderprocessor {
	return &Orderprocessor{
		inventory: inventory,
		relations: make(map[int][]int),
	}
}

func (op *Orderprocessor) processOrders(orders []*Order) []*Order {
	for _, order := range orders {
		err := op.canProcess(order) 
		if err != nil {
			continue
		}

		for _, p := range order.products {
			op.inventory[p.id] -= p.quantity
		}

		order.fulfilled = true
	}

	return orders
}

func (op *Orderprocessor) canProcess(order *Order) error {
	if len(order.products) == 0 {
		return errors.New("invalid req")
	}

	for _, p := range order.products {
		inventoryQty := op.inventory[p.id]
		if p.quantity > inventoryQty {
			return errors.New("inventory insufficient")
		}
	}

	return nil
}

func (op *Orderprocessor) addInventory(productID, quantity int) bool {
	op.inventory[productID] += quantity
	return true
}

//Extension
func (op *Orderprocessor) addRelatedProducts(sourceProductID int, productIDs []int) bool {
	visited := map[int]bool{}

	for _, productID := range productIDs {
		_, exists := op.inventory[productID]
		if !exists {
			continue
		} 

		if sourceProductID == productID {
			continue
		}

		op.addRealtion(sourceProductID, productID, visited)
	}

	return true
}

func (op *Orderprocessor) addRealtion(p1, p2 int, visited map[int]bool) {
	toVisit := []int{p2, p1}

	for i:=0; i<len(toVisit); i++ {
		p := toVisit[i]

		if len(visited) == len(op.inventory) {
			return
		}

		if op.areRelated(p1, p) {
			return
		}

		toVisit = append(toVisit, op.relations[p]...)
		op.relations[p] = append(op.relations[p], []int{p1}...)
		op.relations[p1] = append(op.relations[p1], []int{p}...)
		visited[p] = true
	}
}

func (op *Orderprocessor) areRelated(p1 , p2 int) bool {
	for productID := range op.relations[p1] {
		if productID == p2 {
			return true
		}
	}

	return false
}

func (op *Orderprocessor) getRelatedProducts(productID int) []int {
	return op.relations[productID]
}