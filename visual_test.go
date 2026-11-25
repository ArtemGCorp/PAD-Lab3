package main

// Visual UI test: opens the REAL application window and drives it step by
// step with pauses, so the scenario can be watched. Every step is also checked.
// BNM is replaced by a stub, so the scenario does not depend on the network.
//
//	VISUAL=1 go test -run Visual -v .
//	VISUAL=1 VISUAL_DELAY=2s go test -run Visual -v .   # slower
//
// Without VISUAL the test is skipped and the normal headless tests run.

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/test"

	"pad-lab3/internal/cache"
	"pad-lab3/internal/rates"
)

var visualApp fyne.App

func TestMain(m *testing.M) {
	if os.Getenv("VISUAL") == "" {
		os.Exit(m.Run())
	}
	// Real windows need the GUI event loop on the main thread, so the tests
	// run in a goroutine while a.Run() owns the main thread.
	flag.Parse()
	_ = flag.Set("test.run", "^TestVisual")
	visualApp = app.NewWithID("md.pad.lab3.visual")
	visualApp.Settings().SetTheme(appTheme{})
	code := 1
	go func() {
		defer fyne.Do(visualApp.Quit)
		code = m.Run()
	}()
	visualApp.Run()
	os.Exit(code)
}

type stepper struct {
	t     *testing.T
	win   fyne.Window
	n     int
	delay time.Duration
}

// step shows the step in the window title, runs the action on the UI thread,
// waits, then checks the result.
func (s *stepper) step(name string, action func(), check func() error) {
	s.t.Helper()
	s.n++
	title := fmt.Sprintf("Шаг %d: %s", s.n, name)
	s.t.Log(title)
	fyne.DoAndWait(func() {
		s.win.SetTitle(title)
		if action != nil {
			action()
		}
	})
	time.Sleep(s.delay / 2)
	var err error
	if check != nil {
		fyne.DoAndWait(func() { err = check() })
	}
	if err != nil {
		fyne.DoAndWait(func() { s.win.SetTitle("ОШИБКА: " + title) })
		time.Sleep(2 * s.delay)
		s.t.Fatalf("%s: %v", title, err)
	}
	time.Sleep(s.delay / 2)
}

// typeText types into an entry one character at a time, visibly.
func (s *stepper) typeText(e interface{ SetText(string) }, typed func(r rune), text string) {
	fyne.DoAndWait(func() { e.SetText("") })
	for _, r := range text {
		fyne.DoAndWait(func() { typed(r) })
		time.Sleep(s.delay / 8)
	}
}

func visualDelay() time.Duration {
	if d, err := time.ParseDuration(os.Getenv("VISUAL_DELAY")); err == nil {
		return d
	}
	return 1200 * time.Millisecond
}

func expect(ok bool, format string, args ...any) error {
	if ok {
		return nil
	}
	return fmt.Errorf(format, args...)
}

func openVisual(f rates.Fetcher, cachePath string) *ui {
	var u *ui
	fyne.DoAndWait(func() {
		w := visualApp.NewWindow("Конвертер валют")
		u = newUI(w, &rates.Loader{Fetcher: f, Cache: cache.Store{Path: cachePath}, Now: time.Now})
		w.Show()
	})
	return u
}

func TestVisualScenario(t *testing.T) {
	if visualApp == nil {
		t.Skip("set VISUAL=1 to watch the scenario in a real window")
	}
	cachePath := filepath.Join(t.TempDir(), "rates.json")
	u := openVisual(stubFetcher{}, cachePath)
	s := &stepper{t: t, win: u.win, delay: visualDelay()}

	s.step("загрузка курсов", func() { u.load(false) }, func() error {
		return expect(u.info.Text == "Курс на "+today() && strings.Contains(u.source.Text, "bnm.md"),
			"info=%q source=%q", u.info.Text, u.source.Text)
	})
	s.step("кнопка неактивна без суммы", nil, func() error {
		return expect(u.convert.Disabled() && u.from.Selected == "EUR" && u.to.Selected == "MDL", "initial state")
	})
	s.step("ввод суммы 1250", func() {}, nil)
	s.typeText(u.amount, u.amount.TypedRune, "1250")
	s.step("кнопка стала активной", nil, func() error {
		return expect(!u.convert.Disabled(), "button must be enabled")
	})
	s.step("конвертация EUR → MDL", func() { test.Tap(u.convert) }, func() error {
		return expect(u.result.Text == "1 250,00 EUR = 25 000,00 MDL", "result = %q", u.result.Text)
	})
	s.step("обмен направлений ⇄", func() { test.Tap(u.swap) }, func() error {
		return expect(u.result.Text == "1 250,00 MDL = 62,50 EUR", "result = %q", u.result.Text)
	})
	s.step("кросс-курс USD → JPY", func() {
		u.from.SetSelected("USD")
		u.to.SetSelected("JPY")
		test.Tap(u.convert)
	}, func() error {
		return expect(u.result.Text == "1 250,00 USD = 166 666,67 JPY", "result = %q", u.result.Text)
	})
	s.step("одинаковые валюты", func() { u.to.SetSelected("USD") }, func() error {
		return expect(u.result.Text == "1 250,00 USD = 1 250,00 USD", "result = %q", u.result.Text)
	})
	s.step("некорректная сумма −5", func() {}, nil)
	s.typeText(u.amount, u.amount.TypedRune, "-5")
	s.step("ошибка вместо результата", func() { test.Tap(u.convert) }, func() error {
		return expect(strings.Contains(u.amountErr.Text, "больше нуля") && u.result.Text == "—", "err=%q", u.amountErr.Text)
	})
	s.step("несуществующая дата 31.02.2026", func() {
		u.date.SetText("31.02.2026")
		test.Tap(u.loadBtn)
	}, func() error {
		return expect(u.banner.Visible() && strings.Contains(u.notice.Text, "не существует"), "notice=%q", u.notice.Text)
	})

	// "Restart" without network: a new window over the same cache file.
	fyne.DoAndWait(func() { u.win.Close() })
	off := openVisual(stubFetcher{down: true}, cachePath)
	s.win = off.win
	s.step("перезапуск без сети — курс из кэша", func() { off.load(false) }, func() error {
		return expect(off.banner.Visible() && strings.Contains(off.notice.Text, "Нет соединения"), "notice=%q", off.notice.Text)
	})
	s.typeText(off.amount, off.amount.TypedRune, "100")
	s.step("конвертация работает офлайн", func() { test.Tap(off.convert) }, func() error {
		return expect(off.result.Text == "100,00 EUR = 2 000,00 MDL", "result = %q", off.result.Text)
	})
	s.step("готово — все проверки пройдены", nil, nil)
	fyne.DoAndWait(func() { off.win.Close() })
}
