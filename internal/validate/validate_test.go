package validate

import (
	"errors"
	"testing"
)

func TestParseAmountValid(t *testing.T) {
	cases := map[string]float64{
		"100":       100,
		"  12.5 ":   12.5,
		"12,50":     12.5,
		"0.01":      0.01,
		"1 000":     1000,
		"1000000,5": 1000000.5,
	}
	for in, want := range cases {
		got, err := ParseAmount(in)
		if err != nil || got != want {
			t.Errorf("ParseAmount(%q) = %v, %v; want %v", in, got, err, want)
		}
	}
}

func TestParseAmountInvalid(t *testing.T) {
	cases := map[string]error{
		"":              ErrEmpty,
		"   ":           ErrEmpty,
		"0":             ErrNotPos,
		"0,00":          ErrNotPos,
		"-5":            ErrNotPos,
		"-0.5":          ErrNotPos,
		"abc":           ErrNotNum,
		"12abc":         ErrNotNum,
		"1e5":           ErrNotNum,
		"NaN":           ErrNotNum,
		"Inf":           ErrNotNum,
		"12,3,4":        ErrNotNum,
		"+5":            ErrNotNum,
		".5":            ErrNotNum,
		"5.":            ErrNotNum,
		"--5":           ErrNotNum,
		"1000000000001": ErrTooLarge,
	}
	for in, want := range cases {
		if _, err := ParseAmount(in); !errors.Is(err, want) {
			t.Errorf("ParseAmount(%q) err = %v, want %v", in, err, want)
		}
	}
}

func TestParseDate(t *testing.T) {
	d, err := ParseDate(" 29.02.2024 ")
	if err != nil || d.Day() != 29 || d.Month() != 2 || d.Year() != 2024 {
		t.Fatalf("%v %v", d, err)
	}
	cases := map[string]error{
		"":           ErrDateEmpty,
		"31.02.2026": ErrDateInvalid,
		"29.02.2025": ErrDateInvalid,
		"32.01.2026": ErrDateInvalid,
		"01.13.2026": ErrDateInvalid,
		"00.01.2026": ErrDateInvalid,
		"2026-01-01": ErrDateFormat,
		"1.1.2026":   ErrDateFormat,
		"abc":        ErrDateFormat,
	}
	for in, want := range cases {
		if _, err := ParseDate(in); !errors.Is(err, want) {
			t.Errorf("ParseDate(%q) err = %v, want %v", in, err, want)
		}
	}
}
