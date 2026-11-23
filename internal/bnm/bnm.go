// Package bnm downloads and parses the official exchange rates of the
// National Bank of Moldova (contracts/bnm-xml.md).
package bnm

import (
	"bytes"
	"context"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// Source is the human-readable data source shown in the UI.
const Source = "Национальный банк Молдовы (bnm.md)"

// DefaultBaseURL is the BNM XML endpoint (without the date parameter).
const DefaultBaseURL = "https://www.bnm.md/en/official_exchange_rates"

// DateLayout is the BNM date format.
const DateLayout = "02.01.2006"

var (
	// ErrNoData — the source has no rate for the requested date.
	ErrNoData = errors.New("источник не опубликовал курс на эту дату")
	// ErrUnavailable — network failure, timeout, server error or unexpected format.
	ErrUnavailable = errors.New("источник данных недоступен")
)

// Rate is one currency against MDL.
type Rate struct {
	Code    string  `json:"code"`
	Name    string  `json:"name"`
	Nominal int     `json:"nominal"`
	Value   float64 `json:"value"` // MDL for Nominal units
}

// PerUnit returns MDL for one unit of the currency.
func (r Rate) PerUnit() float64 { return r.Value / float64(r.Nominal) }

// Table is the full set of rates for one date.
type Table struct {
	Date   time.Time       `json:"date"`
	Source string          `json:"source"`
	Rates  map[string]Rate `json:"rates"`
}

// MDL is the base currency, always present with rate 1.
var MDL = Rate{Code: "MDL", Name: "Moldovan Leu", Nominal: 1, Value: 1}

type xmlValCurs struct {
	XMLName xml.Name    `xml:"ValCurs"`
	Date    string      `xml:"Date,attr"`
	Valutes []xmlValute `xml:"Valute"`
}

type xmlValute struct {
	CharCode string `xml:"CharCode"`
	Nominal  string `xml:"Nominal"`
	Name     string `xml:"Name"`
	Value    string `xml:"Value"`
}

var codeRe = regexp.MustCompile(`^[A-Z]{3}$`)

// Parse decodes a BNM XML document. requested is used when the document
// has no Date attribute. Invalid <Valute> items are skipped; a document with
// no valid items returns ErrNoData; non-XML input returns ErrUnavailable.
func Parse(data []byte, requested time.Time) (Table, error) {
	if len(bytes.TrimSpace(data)) == 0 {
		return Table{}, ErrNoData
	}
	var doc xmlValCurs
	if err := xml.Unmarshal(data, &doc); err != nil {
		return Table{}, fmt.Errorf("%w: неожиданный формат ответа (%v)", ErrUnavailable, err)
	}

	date := dateOnly(requested)
	if d, err := time.ParseInLocation(DateLayout, strings.TrimSpace(doc.Date), time.Local); err == nil {
		date = d
	}

	t := Table{Date: date, Source: Source, Rates: map[string]Rate{}}
	for _, v := range doc.Valutes {
		code := strings.ToUpper(strings.TrimSpace(v.CharCode))
		if !codeRe.MatchString(code) || code == "MDL" {
			continue
		}
		nom, err := strconv.Atoi(strings.TrimSpace(v.Nominal))
		if err != nil || nom < 1 {
			continue
		}
		val, err := strconv.ParseFloat(strings.ReplaceAll(strings.TrimSpace(v.Value), ",", "."), 64)
		if err != nil || !(val > 0) || val > 1e9 {
			continue
		}
		t.Rates[code] = Rate{Code: code, Name: strings.TrimSpace(v.Name), Nominal: nom, Value: val}
	}
	if len(t.Rates) == 0 {
		return Table{}, ErrNoData
	}
	t.Rates["MDL"] = MDL
	return t, nil
}

// Client fetches rates over HTTP.
type Client struct {
	BaseURL string
	HTTP    *http.Client
}

// NewClient returns a client for the real BNM endpoint with a 10 s timeout.
func NewClient() *Client {
	return &Client{BaseURL: DefaultBaseURL, HTTP: &http.Client{Timeout: 10 * time.Second}}
}

// Fetch downloads and parses the rates for date.
func (c *Client) Fetch(ctx context.Context, date time.Time) (Table, error) {
	url := fmt.Sprintf("%s?get_xml=1&date=%s", c.BaseURL, date.Format(DateLayout))
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return Table{}, fmt.Errorf("%w: %v", ErrUnavailable, err)
	}
	req.Header.Set("Accept", "application/xml")
	resp, err := c.HTTP.Do(req)
	if err != nil {
		return Table{}, fmt.Errorf("%w: %v", ErrUnavailable, err)
	}
	defer resp.Body.Close()

	switch {
	case resp.StatusCode == http.StatusNotFound:
		return Table{}, ErrNoData
	case resp.StatusCode != http.StatusOK:
		return Table{}, fmt.Errorf("%w: HTTP %d", ErrUnavailable, resp.StatusCode)
	}
	body, err := io.ReadAll(io.LimitReader(resp.Body, 2<<20))
	if err != nil {
		return Table{}, fmt.Errorf("%w: %v", ErrUnavailable, err)
	}
	return Parse(body, date)
}

func dateOnly(t time.Time) time.Time {
	y, m, d := t.Date()
	return time.Date(y, m, d, 0, 0, 0, 0, time.Local)
}
