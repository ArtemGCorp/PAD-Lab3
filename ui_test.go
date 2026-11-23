package main

import (
	"context"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"fyne.io/fyne/v2/test"

	"pad-lab3/internal/bnm"
	"pad-lab3/internal/cache"
	"pad-lab3/internal/rates"
)

type stubFetcher struct {
	down bool
}

func (s stubFetcher) Fetch(_ context.Context, d time.Time) (bnm.Table, error) {
	if s.down {
		return bnm.Table{}, bnm.ErrUnavailable
	}
	return bnm.Table{Date: d, Source: bnm.Source, Rates: map[string]bnm.Rate{
		"MDL": bnm.MDL,
		"EUR": {Code: "EUR", Nominal: 1, Value: 20},
		"USD": {Code: "USD", Nominal: 1, Value: 16},
	}}, nil
}

func newTestUI(t *testing.T, f rates.Fetcher, cachePath string) *ui {
	a := test.NewTempApp(t)
	w := a.NewWindow("test")
	l := &rates.Loader{Fetcher: f, Cache: cache.Store{Path: cachePath}, Now: time.Now}
	u := newUI(w, l)
	u.spawn = func(f func()) { f() } // synchronous in tests
	w.Show()
	u.load(false)
	return u
}

func TestConvertFlow(t *testing.T) {
	u := newTestUI(t, stubFetcher{}, filepath.Join(t.TempDir(), "c.json"))

	if !strings.Contains(u.info.Text, "Источник: "+bnm.Source) || !strings.Contains(u.info.Text, time.Now().Format(bnm.DateLayout)) {
		t.Fatalf("info = %q", u.info.Text)
	}
	if u.from.Selected != "EUR" || u.to.Selected != "MDL" {
		t.Fatalf("defaults %s→%s", u.from.Selected, u.to.Selected)
	}
	if !u.convert.Disabled() {
		t.Fatal("button must be disabled with empty amount")
	}

	test.Type(u.amount, "100")
	if u.convert.Disabled() {
		t.Fatal("button must be enabled")
	}
	test.Tap(u.convert)
	if u.result.Text != "100,00 EUR = 2 000,00 MDL" {
		t.Fatalf("result = %q", u.result.Text)
	}

	test.Tap(u.swap) // reverse direction
	if u.from.Selected != "MDL" || u.to.Selected != "EUR" || u.result.Text != "100,00 MDL = 5,00 EUR" {
		t.Fatalf("after swap: %s→%s %q", u.from.Selected, u.to.Selected, u.result.Text)
	}

	u.to.SetSelected("MDL") // same currency
	if u.result.Text != "100,00 MDL = 100,00 MDL" {
		t.Fatalf("same currency: %q", u.result.Text)
	}
}

func TestInvalidAmount(t *testing.T) {
	u := newTestUI(t, stubFetcher{}, filepath.Join(t.TempDir(), "c.json"))
	for _, in := range []string{"-5", "0", "abc"} {
		u.amount.SetText(in)
		test.Tap(u.convert)
		if u.amountErr.Text == "" || u.result.Text != "" {
			t.Fatalf("%q: err=%q result=%q", in, u.amountErr.Text, u.result.Text)
		}
	}
}

func TestOfflineNoCache(t *testing.T) {
	u := newTestUI(t, stubFetcher{down: true}, filepath.Join(t.TempDir(), "c.json"))
	if len(u.from.Options) != 0 || !strings.Contains(u.notice.Text, "нет сохранённых курсов") {
		t.Fatalf("options=%v notice=%q", u.from.Options, u.notice.Text)
	}
	u.amount.SetText("10")
	if !u.convert.Disabled() {
		t.Fatal("cannot convert without rates")
	}
}

func TestOfflineWithCacheAfterRestart(t *testing.T) {
	p := filepath.Join(t.TempDir(), "c.json")
	newTestUI(t, stubFetcher{}, p) // first run online fills the cache
	u := newTestUI(t, stubFetcher{down: true}, p)
	if !strings.Contains(u.notice.Text, "Нет соединения") || len(u.from.Options) == 0 {
		t.Fatalf("notice=%q", u.notice.Text)
	}
	u.amount.SetText("32")
	u.from.SetSelected("USD")
	test.Tap(u.convert)
	if u.result.Text != "32,00 USD = 512,00 MDL" {
		t.Fatalf("result=%q", u.result.Text)
	}
}

func TestInvalidDate(t *testing.T) {
	u := newTestUI(t, stubFetcher{}, filepath.Join(t.TempDir(), "c.json"))
	u.date.SetText("31.02.2026")
	test.Tap(u.loadBtn)
	if !strings.Contains(u.notice.Text, "не существует") {
		t.Fatalf("notice=%q", u.notice.Text)
	}
}

func TestFormatMoney(t *testing.T) {
	for in, want := range map[float64]string{0.5: "0,50", 1234.567: "1 234,57", 1000000: "1 000 000,00", 12: "12,00"} {
		if got := formatMoney(in); got != want {
			t.Errorf("formatMoney(%v) = %q, want %q", in, got, want)
		}
	}
}
