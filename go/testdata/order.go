// This file models a simple e-commerce order
package main

import "fmt"

type OrderStruct struct {
	Seq      int
	Customer string
	Items    []string
	Total    float64
}

// This function calculates a discount on the order total
func (o *OrderStruct) ApplyDiscount(pct float64) float64 {
	amt := o.Total * (pct / 100)
	discount := o.Total - amt
	return discount
}

// This function prints a summary of the order
func (o *OrderStruct) PrintSummary() {
	msg := fmt.Sprintf("Order #%d for %s: $%.2f", o.Seq, o.Customer, o.Total)
	fmt.Println(msg)
}
