package cache

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	"pad-lab3/internal/bnm"
)

func TestSaveLoadRoundTrip(t *testing.T) {
	s := Store{Path: filepath.Join(t.TempDir(), "sub", "rates.json")}
	date := time.Date(2026, 11, 20, 0, 0, 0, 0, time.Local)
	in := Entry{
		FetchedAt: time.Date(2026, 11, 20, 14, 0, 0, 0, time.UTC),
		Table: bnm.Table{Date: date, Source: bnm.Source, Rates: map[string]bnm.Rate{
			"MDL": bnm.MDL, "EUR": {Code: "EUR", Name: "Euro", Nominal: 1, Value: 20.1267},
		}},
	}
	if err := s.Save(in); err != nil {
		t.Fatal(err)
	}
	// A new Store over the same path simulates an app restart.
	out, err := Store{Path: s.Path}.Load()
	if err != nil {
		t.Fatal(err)
	}
	if !out.Table.Date.Equal(date) || out.Table.Rates["EUR"].Value != 20.1267 || !out.FetchedAt.Equal(in.FetchedAt) {
		t.Fatalf("got %+v", out)
	}
	if _, err := os.Stat(s.Path + ".tmp"); !os.IsNotExist(err) {
		t.Fatal("tmp file left behind")
	}
}

func TestLoadMissing(t *testing.T) {
	if _, err := (Store{Path: filepath.Join(t.TempDir(), "none.json")}).Load(); !errors.Is(err, ErrNoCache) {
		t.Fatalf("err = %v", err)
	}
}

func TestLoadCorrupted(t *testing.T) {
	for name, body := range map[string]string{
		"garbage":     "{not json",
		"empty rates": `{"table":{"rates":{}}}`,
	} {
		t.Run(name, func(t *testing.T) {
			p := filepath.Join(t.TempDir(), "rates.json")
			os.WriteFile(p, []byte(body), 0o644)
			if _, err := (Store{Path: p}).Load(); !errors.Is(err, ErrNoCache) {
				t.Fatalf("err = %v", err)
			}
		})
	}
}
