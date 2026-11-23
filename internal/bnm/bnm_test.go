package bnm

import (
	"context"
	"errors"
	"math"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

const sample = `<?xml version="1.0" encoding="UTF-8"?>
<ValCurs Date="20.11.2026" name="Official exchange rate">
  <Valute ID="47"><NumCode>978</NumCode><CharCode>EUR</CharCode><Nominal>1</Nominal><Name>Euro</Name><Value>20.1267</Value></Valute>
  <Valute ID="44"><NumCode>840</NumCode><CharCode>USD</CharCode><Nominal>1</Nominal><Name>US Dollar</Name><Value>18,3070</Value></Valute>
  <Valute ID="1"><NumCode>392</NumCode><CharCode>JPY</CharCode><Nominal>100</Nominal><Name>Yen</Name><Value>11.8000</Value></Valute>
  <Valute ID="2"><CharCode>BAD</CharCode><Nominal>1</Nominal><Value>abc</Value></Valute>
  <Valute ID="3"><CharCode>ZER</CharCode><Nominal>0</Nominal><Value>1</Value></Valute>
  <Valute ID="4"><CharCode>toolong</CharCode><Nominal>1</Nominal><Value>1</Value></Valute>
</ValCurs>`

var req = time.Date(2026, 11, 20, 0, 0, 0, 0, time.Local)

func TestParseValid(t *testing.T) {
	tb, err := Parse([]byte(sample), req)
	if err != nil {
		t.Fatal(err)
	}
	if tb.Date.Format(DateLayout) != "20.11.2026" || tb.Source != Source {
		t.Fatalf("meta: %+v", tb)
	}
	if len(tb.Rates) != 4 { // EUR USD JPY + MDL
		t.Fatalf("rates: %+v", tb.Rates)
	}
	if tb.Rates["USD"].Value != 18.307 {
		t.Fatalf("comma decimal: %v", tb.Rates["USD"])
	}
	if got := tb.Rates["JPY"].PerUnit(); math.Abs(got-0.118) > 1e-12 {
		t.Fatalf("JPY per unit = %v", got)
	}
	if tb.Rates["MDL"].PerUnit() != 1 {
		t.Fatal("MDL must be 1")
	}
}

func TestParseEmptyAndBroken(t *testing.T) {
	cases := map[string]struct {
		body string
		want error
	}{
		"empty body":     {"", ErrNoData},
		"whitespace":     {"  \n", ErrNoData},
		"empty ValCurs":  {`<ValCurs Date="21.11.2026"></ValCurs>`, ErrNoData},
		"only bad items": {`<ValCurs><Valute><CharCode>X</CharCode></Valute></ValCurs>`, ErrNoData},
		"html":           {`<!DOCTYPE html><html><body>Error</body></html>`, ErrUnavailable},
		"other root":     {`<root/>`, ErrUnavailable},
		"garbage":        {`not xml at all`, ErrUnavailable},
	}
	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			if _, err := Parse([]byte(c.body), req); !errors.Is(err, c.want) {
				t.Fatalf("err = %v, want %v", err, c.want)
			}
		})
	}
}

func TestParseUsesRequestedDateWhenMissing(t *testing.T) {
	tb, err := Parse([]byte(`<ValCurs><Valute><CharCode>EUR</CharCode><Nominal>1</Nominal><Value>20</Value></Valute></ValCurs>`), req)
	if err != nil || !tb.Date.Equal(req) {
		t.Fatalf("%v %v", tb.Date, err)
	}
}

func server(t *testing.T, h http.HandlerFunc) *Client {
	s := httptest.NewServer(h)
	t.Cleanup(s.Close)
	return &Client{BaseURL: s.URL, HTTP: &http.Client{Timeout: 500 * time.Millisecond}}
}

func TestClientFetch(t *testing.T) {
	var gotQuery string
	c := server(t, func(w http.ResponseWriter, r *http.Request) {
		gotQuery = r.URL.RawQuery
		w.Write([]byte(sample))
	})
	tb, err := c.Fetch(context.Background(), req)
	if err != nil || len(tb.Rates) != 4 {
		t.Fatalf("%v %v", tb, err)
	}
	if gotQuery != "get_xml=1&date=20.11.2026" {
		t.Fatalf("query = %s", gotQuery)
	}
}

func TestClientErrors(t *testing.T) {
	cases := map[string]struct {
		h    http.HandlerFunc
		want error
	}{
		"404":   {func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(404) }, ErrNoData},
		"500":   {func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(500) }, ErrUnavailable},
		"empty": {func(w http.ResponseWriter, _ *http.Request) {}, ErrNoData},
		"slow": {func(w http.ResponseWriter, r *http.Request) {
			select {
			case <-time.After(2 * time.Second):
			case <-r.Context().Done():
			}
		}, ErrUnavailable},
	}
	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			if _, err := server(t, c.h).Fetch(context.Background(), req); !errors.Is(err, c.want) {
				t.Fatalf("err = %v, want %v", err, c.want)
			}
		})
	}
}

func TestClientNoNetwork(t *testing.T) {
	c := &Client{BaseURL: "http://127.0.0.1:1", HTTP: &http.Client{Timeout: time.Second}}
	if _, err := c.Fetch(context.Background(), req); !errors.Is(err, ErrUnavailable) {
		t.Fatalf("err = %v", err)
	}
}
