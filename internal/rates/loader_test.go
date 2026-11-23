package rates

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"pad-lab3/internal/bnm"
	"pad-lab3/internal/cache"
)

// fake answers per date: a table, ErrNoData or ErrUnavailable.
type fake struct {
	data  map[string]float64 // date → EUR value
	down  bool
	calls []string
}

func (f *fake) Fetch(_ context.Context, d time.Time) (bnm.Table, error) {
	key := d.Format(bnm.DateLayout)
	f.calls = append(f.calls, key)
	if f.down {
		return bnm.Table{}, fmt.Errorf("%w: dial tcp: no route", bnm.ErrUnavailable)
	}
	v, ok := f.data[key]
	if !ok {
		return bnm.Table{}, bnm.ErrNoData
	}
	return bnm.Table{Date: d, Source: bnm.Source, Rates: map[string]bnm.Rate{
		"MDL": bnm.MDL, "EUR": {Code: "EUR", Nominal: 1, Value: v},
	}}, nil
}

func date(s string) time.Time {
	t, _ := time.ParseInLocation(bnm.DateLayout, s, time.Local)
	return t
}

func newLoader(t *testing.T, f *fake) *Loader {
	return &Loader{
		Fetcher: f,
		Cache:   cache.Store{Path: filepath.Join(t.TempDir(), "rates.json")},
		Now:     func() time.Time { return date("24.11.2026").Add(15 * time.Hour) },
	}
}

func TestExactDate(t *testing.T) {
	l := newLoader(t, &fake{data: map[string]float64{"20.11.2026": 20}})
	r, err := l.Load(context.Background(), date("20.11.2026"))
	if err != nil || r.Fallback || r.FromCache || r.Notice != "" {
		t.Fatalf("%+v %v", r, err)
	}
	if _, err := l.Cache.Load(); err != nil {
		t.Fatal("successful load must be cached")
	}
}

func TestHolidayFallsBackToLastPublished(t *testing.T) {
	f := &fake{data: map[string]float64{"20.11.2026": 20}}
	l := newLoader(t, f)
	r, err := l.Load(context.Background(), date("23.11.2026"))
	if err != nil {
		t.Fatal(err)
	}
	if !r.Fallback || r.Table.Date.Format(bnm.DateLayout) != "20.11.2026" {
		t.Fatalf("%+v", r)
	}
	if !strings.Contains(r.Notice, "23.11.2026") || !strings.Contains(r.Notice, "20.11.2026") {
		t.Fatalf("notice = %q", r.Notice)
	}
	if strings.Join(f.calls, ",") != "23.11.2026,22.11.2026,21.11.2026,20.11.2026" {
		t.Fatalf("calls = %v", f.calls)
	}
}

func TestNoDataAtAll(t *testing.T) {
	f := &fake{data: map[string]float64{}}
	_, err := newLoader(t, f).Load(context.Background(), date("23.11.2026"))
	if err == nil || len(f.calls) != MaxBack+1 {
		t.Fatalf("err = %v calls = %d", err, len(f.calls))
	}
}

func TestOfflineUsesCacheAcrossRestart(t *testing.T) {
	f := &fake{data: map[string]float64{"20.11.2026": 20.5}}
	l := newLoader(t, f)
	if _, err := l.Load(context.Background(), date("20.11.2026")); err != nil {
		t.Fatal(err)
	}

	// "Restart": new loader, same cache file, network down.
	l2 := &Loader{Fetcher: &fake{down: true}, Cache: cache.Store{Path: l.Cache.Path}, Now: l.Now}
	r, err := l2.Load(context.Background(), date("24.11.2026"))
	if err != nil {
		t.Fatal(err)
	}
	if !r.FromCache || r.Table.Rates["EUR"].Value != 20.5 || !strings.Contains(r.Notice, "Нет соединения") ||
		!strings.Contains(r.Notice, "20.11.2026") {
		t.Fatalf("%+v", r)
	}
}

func TestOfflineWithoutCache(t *testing.T) {
	f := &fake{down: true}
	_, err := newLoader(t, f).Load(context.Background(), date("24.11.2026"))
	if err == nil || !strings.Contains(err.Error(), "нет сохранённых курсов") {
		t.Fatalf("err = %v", err)
	}
	if len(f.calls) != 1 {
		t.Fatalf("must not walk back dates when the source is down: %v", f.calls)
	}
}

func TestFutureDate(t *testing.T) {
	f := &fake{}
	_, err := newLoader(t, f).Load(context.Background(), date("25.11.2026"))
	if !errors.Is(err, ErrFutureDate) || len(f.calls) != 0 {
		t.Fatalf("err = %v calls = %v", err, f.calls)
	}
}
