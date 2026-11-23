package convert

import (
	"math"
	"testing"

	"pad-lab3/internal/bnm"
)

var table = bnm.Table{Rates: map[string]bnm.Rate{
	"MDL": bnm.MDL,
	"EUR": {Code: "EUR", Nominal: 1, Value: 20},
	"USD": {Code: "USD", Nominal: 1, Value: 16},
	"JPY": {Code: "JPY", Nominal: 100, Value: 12},
}}

func near(a, b float64) bool { return math.Abs(a-b) < 1e-9 }

func TestConvert(t *testing.T) {
	cases := []struct {
		amount   float64
		from, to string
		want     float64
	}{
		{100, "EUR", "MDL", 2000},  // forward
		{2000, "MDL", "EUR", 100},  // reverse
		{100, "EUR", "USD", 125},   // cross via MDL
		{125, "USD", "EUR", 100},   // cross reverse
		{1000, "JPY", "MDL", 120},  // nominal 100
		{120, "MDL", "JPY", 1000},  // nominal reverse
		{42.5, "USD", "USD", 42.5}, // same currency
	}
	for _, c := range cases {
		r, err := Convert(c.amount, c.from, c.to, table)
		if err != nil || !near(r.Amount, c.want) {
			t.Errorf("%v %s→%s = %v (%v), want %v", c.amount, c.from, c.to, r.Amount, err, c.want)
		}
	}
}

func TestRoundTrip(t *testing.T) {
	for _, from := range []string{"EUR", "USD", "JPY", "MDL"} {
		for _, to := range []string{"EUR", "USD", "JPY", "MDL"} {
			a, _ := Convert(123.45, from, to, table)
			b, _ := Convert(a.Amount, to, from, table)
			if !near(b.Amount, 123.45) {
				t.Errorf("%s→%s→%s = %v", from, to, from, b.Amount)
			}
		}
	}
}

func TestRateValue(t *testing.T) {
	r, _ := Convert(1, "EUR", "USD", table)
	if !near(r.Rate, 1.25) {
		t.Fatalf("rate = %v", r.Rate)
	}
}

func TestUnknownCurrency(t *testing.T) {
	if _, err := Convert(1, "XXX", "MDL", table); err == nil {
		t.Fatal("expected error for unknown source")
	}
	if _, err := Convert(1, "MDL", "XXX", table); err == nil {
		t.Fatal("expected error for unknown target")
	}
}

func TestSameCurrencyWithEmptyTable(t *testing.T) {
	r, err := Convert(5, "GBP", "GBP", bnm.Table{})
	if err != nil || r.Amount != 5 {
		t.Fatalf("%v %v", r, err)
	}
}
