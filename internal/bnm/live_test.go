package bnm

import (
	"context"
	"errors"
	"os"
	"testing"
	"time"
)

// TestLiveBNM hits the real bnm.md. Skipped unless LAB3_LIVE=1:
//
//	LAB3_LIVE=1 go test ./internal/bnm -run Live -v
func TestLiveBNM(t *testing.T) {
	if os.Getenv("LAB3_LIVE") != "1" {
		t.Skip("set LAB3_LIVE=1 to query bnm.md")
	}
	c := NewClient()
	tb, err := c.Fetch(context.Background(), time.Now())
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("%s: EUR=%.4f USD=%.4f (%d currencies)", tb.Date.Format(DateLayout),
		tb.Rates["EUR"].Value, tb.Rates["USD"].Value, len(tb.Rates))
	if tb.Rates["EUR"].Value <= 0 || tb.Rates["USD"].Value <= 0 {
		t.Fatal("EUR/USD missing")
	}

	_, err = c.Fetch(context.Background(), time.Now().AddDate(1, 0, 0))
	if !errors.Is(err, ErrNoData) {
		t.Fatalf("future date: err = %v, want ErrNoData", err)
	}
}
