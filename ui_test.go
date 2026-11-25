package main

import (
	"context"
	"image/png"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/test"
	"fyne.io/fyne/v2/theme"

	"pad-lab3/internal/bnm"
	"pad-lab3/internal/cache"
	"pad-lab3/internal/rates"
)

// stubFetcher serves fixed rates; `missing` lists dates with no publication.
type stubFetcher struct {
	down    bool
	missing map[string]bool
}

func (s stubFetcher) Fetch(_ context.Context, d time.Time) (bnm.Table, error) {
	if s.down {
		return bnm.Table{}, bnm.ErrUnavailable
	}
	if s.missing[d.Format(bnm.DateLayout)] {
		return bnm.Table{}, bnm.ErrNoData
	}
	return bnm.Table{Date: d, Source: bnm.Source, Rates: map[string]bnm.Rate{
		"MDL": bnm.MDL,
		"EUR": {Code: "EUR", Name: "Euro", Nominal: 1, Value: 20},
		"USD": {Code: "USD", Name: "US Dollar", Nominal: 1, Value: 16},
		"JPY": {Code: "JPY", Name: "Yen", Nominal: 100, Value: 12},
	}}, nil
}

// uiTheme is the theme used by newTestUI.
var uiTheme fyne.Theme = appTheme{}

func newTestUI(t *testing.T, f rates.Fetcher, cachePath string) *ui {
	t.Helper()
	a := test.NewTempApp(t)
	a.Settings().SetTheme(uiTheme)
	w := a.NewWindow("test")
	l := &rates.Loader{Fetcher: f, Cache: cache.Store{Path: cachePath}, Now: time.Now}
	u := newUI(w, l)
	u.spawn = func(f func()) { f() } // synchronous in tests
	w.Show()
	u.load(false)
	return u
}

func online(t *testing.T) *ui {
	return newTestUI(t, stubFetcher{}, filepath.Join(t.TempDir(), "c.json"))
}

func today() string { return time.Now().Format(bnm.DateLayout) }

func TestShowsRateDateAndSource(t *testing.T) {
	u := online(t)
	if u.info.Text != "Курс на "+today() {
		t.Fatalf("info = %q", u.info.Text)
	}
	if u.source.Text != "Источник: "+bnm.Source {
		t.Fatalf("source = %q", u.source.Text)
	}
	if u.banner.Visible() {
		t.Fatalf("no warning expected, got %q", u.notice.Text)
	}
}

func TestCurrencyListsAndDefaults(t *testing.T) {
	u := online(t)
	if u.from.Selected != "EUR" || u.to.Selected != "MDL" {
		t.Fatalf("defaults %s→%s", u.from.Selected, u.to.Selected)
	}
	if got := strings.Join(u.from.Options, ","); got != "MDL,EUR,USD,JPY" {
		t.Fatalf("options = %s (popular first)", got)
	}
}

func TestButtonDisabledUntilAmountEntered(t *testing.T) {
	u := online(t)
	if !u.convert.Disabled() {
		t.Fatal("disabled with empty amount")
	}
	test.Type(u.amount, "1")
	if u.convert.Disabled() {
		t.Fatal("enabled with amount and both currencies")
	}
	u.amount.SetText("   ")
	if !u.convert.Disabled() {
		t.Fatal("blank amount disables the button")
	}
}

func TestConvertForwardSwapAndSameCurrency(t *testing.T) {
	u := online(t)
	test.Type(u.amount, "100")
	test.Tap(u.convert)
	if u.result.Text != "100,00 EUR = 2 000,00 MDL" {
		t.Fatalf("result = %q", u.result.Text)
	}
	if !strings.HasPrefix(u.rateLine.Text, "1 EUR = 20.0000 MDL") || !strings.Contains(u.rateLine.Text, today()) {
		t.Fatalf("rate line = %q", u.rateLine.Text)
	}

	test.Tap(u.swap)
	if u.from.Selected != "MDL" || u.to.Selected != "EUR" || u.result.Text != "100,00 MDL = 5,00 EUR" {
		t.Fatalf("after swap: %s→%s %q", u.from.Selected, u.to.Selected, u.result.Text)
	}

	u.to.SetSelected("MDL")
	if u.result.Text != "100,00 MDL = 100,00 MDL" {
		t.Fatalf("same currency: %q", u.result.Text)
	}
}

func TestCrossRateAndNominal(t *testing.T) {
	u := online(t)
	u.amount.SetText("1000")
	u.from.SetSelected("JPY")
	u.to.SetSelected("USD")
	test.Tap(u.convert)
	if u.result.Text != "1 000,00 JPY = 7,50 USD" {
		t.Fatalf("result = %q", u.result.Text)
	}
}

func TestEnterKeyConverts(t *testing.T) {
	u := online(t)
	test.Type(u.amount, "2")
	u.amount.TypedKey(&fyne.KeyEvent{Name: fyne.KeyReturn})
	if u.result.Text != "2,00 EUR = 40,00 MDL" {
		t.Fatalf("result = %q", u.result.Text)
	}
}

func TestInvalidAmount(t *testing.T) {
	u := online(t)
	for in, msg := range map[string]string{"-5": "больше нуля", "0": "больше нуля", "abc": "числом", "1e5": "числом"} {
		u.amount.SetText(in)
		test.Tap(u.convert)
		if !strings.Contains(u.amountErr.Text, msg) || u.result.Text != "—" {
			t.Fatalf("%q: err=%q result=%q", in, u.amountErr.Text, u.result.Text)
		}
	}
	test.Type(u.amount, "0") // typing clears the error
	if u.amountErr.Text != "" {
		t.Fatalf("error not cleared: %q", u.amountErr.Text)
	}
}

func TestInvalidAndFutureDate(t *testing.T) {
	u := online(t)
	u.date.SetText("31.02.2026")
	test.Tap(u.loadBtn)
	if !u.banner.Visible() || !strings.Contains(u.notice.Text, "не существует") {
		t.Fatalf("notice=%q", u.notice.Text)
	}
	u.date.SetText(time.Now().AddDate(0, 0, 3).Format(bnm.DateLayout))
	test.Tap(u.loadBtn)
	if !strings.Contains(u.notice.Text, "будущую дату") || u.info.Text != "Курс на "+today() {
		t.Fatalf("notice=%q info=%q", u.notice.Text, u.info.Text)
	}
}

func TestHolidayFallbackIsAnnounced(t *testing.T) {
	d := time.Now().AddDate(0, 0, -2)
	key := d.Format(bnm.DateLayout)
	u := newTestUI(t, stubFetcher{missing: map[string]bool{key: true}}, filepath.Join(t.TempDir(), "c.json"))
	u.date.SetText(key)
	test.Tap(u.loadBtn)
	prev := d.AddDate(0, 0, -1).Format(bnm.DateLayout)
	if !u.banner.Visible() || !strings.Contains(u.notice.Text, key) || !strings.Contains(u.notice.Text, prev) {
		t.Fatalf("notice=%q", u.notice.Text)
	}
	if u.info.Text != "Курс на "+prev {
		t.Fatalf("info=%q", u.info.Text)
	}
}

func TestOfflineNoCache(t *testing.T) {
	u := newTestUI(t, stubFetcher{down: true}, filepath.Join(t.TempDir(), "c.json"))
	if len(u.from.Options) != 0 || !u.banner.Visible() || !strings.Contains(u.notice.Text, "нет сохранённых курсов") {
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
	if !u.banner.Visible() || !strings.Contains(u.notice.Text, "Нет соединения") || !strings.Contains(u.notice.Text, today()) {
		t.Fatalf("notice=%q", u.notice.Text)
	}
	u.amount.SetText("32")
	u.from.SetSelected("USD")
	test.Tap(u.convert)
	if u.result.Text != "32,00 USD = 512,00 MDL" {
		t.Fatalf("result=%q", u.result.Text)
	}
}

func TestFormatMoney(t *testing.T) {
	for in, want := range map[float64]string{0.5: "0,50", 1234.567: "1 234,57", 1000000: "1 000 000,00", 12: "12,00"} {
		if got := formatMoney(in); got != want {
			t.Errorf("formatMoney(%v) = %q, want %q", in, got, want)
		}
	}
}

// TestScreenshots renders the window to PNG for a visual check:
//
//	LAB3_SCREENSHOTS=/some/dir go test -run Screenshots .
func TestScreenshots(t *testing.T) {
	out := os.Getenv("LAB3_SCREENSHOTS")
	if out == "" {
		t.Skip("set LAB3_SCREENSHOTS=<dir>")
	}
	for name, v := range map[string]fyne.ThemeVariant{"light": theme.VariantLight, "dark": theme.VariantDark} {
		t.Run(name, func(t *testing.T) {
			uiTheme = fixedVariant{appTheme{}, v}
			defer func() { uiTheme = appTheme{} }()

			u := online(t)
			test.Type(u.amount, "1250")
			test.Tap(u.convert)
			save(t, u.win, filepath.Join(out, "lab3-"+name+".png"))

			p := filepath.Join(t.TempDir(), "c.json")
			newTestUI(t, stubFetcher{}, p)
			off := newTestUI(t, stubFetcher{down: true}, p)
			save(t, off.win, filepath.Join(out, "lab3-offline-"+name+".png"))
		})
	}
}

func save(t *testing.T, w fyne.Window, path string) {
	t.Helper()
	w.Resize(fyne.NewSize(600, 760))
	f, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	if err := png.Encode(f, w.Canvas().Capture()); err != nil {
		t.Fatal(err)
	}
}
