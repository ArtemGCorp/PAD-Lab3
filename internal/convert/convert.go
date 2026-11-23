// Package convert converts amounts between currencies through MDL.
package convert

import (
	"fmt"

	"pad-lab3/internal/bnm"
)

// Result of one conversion.
type Result struct {
	Amount float64 // converted amount
	Rate   float64 // units of `to` for one unit of `from`
}

// Convert converts amount from one currency to another using table.
// The same currency returns the amount unchanged with rate 1.
func Convert(amount float64, from, to string, table bnm.Table) (Result, error) {
	if from == to {
		return Result{Amount: amount, Rate: 1}, nil
	}
	f, ok := table.Rates[from]
	if !ok {
		return Result{}, fmt.Errorf("нет курса для валюты %s", from)
	}
	t, ok := table.Rates[to]
	if !ok {
		return Result{}, fmt.Errorf("нет курса для валюты %s", to)
	}
	rate := f.PerUnit() / t.PerUnit()
	return Result{Amount: amount * rate, Rate: rate}, nil
}
