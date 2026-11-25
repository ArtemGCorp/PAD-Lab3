package main

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"pad-lab3/internal/bnm"
	"pad-lab3/internal/convert"
	"pad-lab3/internal/rates"
	"pad-lab3/internal/validate"
)

type ui struct {
	win    fyne.Window
	loader *rates.Loader
	// spawn runs background work; tests replace it with a synchronous call.
	spawn func(func())

	amount    *widget.Entry
	amountErr *widget.Label
	from      *widget.Select
	to        *widget.Select
	swap      *widget.Button
	convert   *widget.Button
	result    *canvas.Text
	rateLine  *widget.Label

	date    *widget.Entry
	loadBtn *widget.Button
	info    *widget.Label
	source  *widget.Label
	notice  *widget.Label
	banner  fyne.CanvasObject

	table  *bnm.Table
	loaded bool
}

func newUI(win fyne.Window, loader *rates.Loader) *ui {
	u := &ui{win: win, loader: loader, spawn: func(f func()) { go f() }}

	u.amount = widget.NewEntry()
	u.amount.SetPlaceHolder("Например, 100 или 12,50")
	u.amount.OnChanged = func(string) { u.setAmountErr(""); u.refreshButton() }
	u.amount.OnSubmitted = func(string) {
		if !u.convert.Disabled() {
			u.onConvert()
		}
	}
	u.amountErr = widget.NewLabel("")
	u.amountErr.Importance = widget.DangerImportance
	u.amountErr.Hide()

	u.from = widget.NewSelect(nil, func(string) { u.onCurrencyChanged() })
	u.from.PlaceHolder = "Из валюты"
	u.to = widget.NewSelect(nil, func(string) { u.onCurrencyChanged() })
	u.to.PlaceHolder = "В валюту"
	u.swap = widget.NewButtonWithIcon("", theme.ViewRefreshIcon(), u.onSwap)

	u.convert = widget.NewButtonWithIcon("Конвертировать", theme.MailForwardIcon(), u.onConvert)
	u.convert.Importance = widget.HighImportance

	u.result = canvas.NewText("—", theme.Color(theme.ColorNamePrimary))
	u.result.TextSize = 28
	u.result.TextStyle = fyne.TextStyle{Bold: true}
	u.result.Alignment = fyne.TextAlignCenter
	u.rateLine = widget.NewLabelWithStyle("Введите сумму и нажмите «Конвертировать»", fyne.TextAlignCenter, fyne.TextStyle{})

	u.date = widget.NewEntry()
	u.date.SetText(time.Now().Format(bnm.DateLayout))
	u.date.SetPlaceHolder("ДД.ММ.ГГГГ")
	u.loadBtn = widget.NewButtonWithIcon("Загрузить курс", theme.DownloadIcon(), func() { u.load(true) })

	u.info = widget.NewLabel("Загрузка курсов…")
	u.info.TextStyle = fyne.TextStyle{Bold: true}
	u.source = widget.NewLabel("Источник: " + bnm.Source)
	u.notice = widget.NewLabel("")
	u.notice.Wrapping = fyne.TextWrapWord
	u.banner = banner(u.notice)
	u.banner.Hide()

	currencies := container.NewBorder(nil, nil, nil, nil,
		container.NewBorder(nil, nil, nil, container.NewVBox(layout.NewSpacer(), u.swap),
			container.NewGridWithColumns(2,
				container.NewVBox(fieldLabel("Из валюты"), u.from),
				container.NewVBox(fieldLabel("В валюту"), u.to))))

	input := card(container.NewVBox(
		fieldLabel("Сумма"), u.amount, u.amountErr,
		currencies,
		u.convert,
	))

	output := card(container.NewVBox(u.result, u.rateLine))

	rateInfo := card(container.NewVBox(
		container.NewBorder(nil, nil, widget.NewIcon(theme.HistoryIcon()), nil, u.info),
		container.NewBorder(nil, nil, widget.NewIcon(theme.InfoIcon()), nil, u.source),
		widget.NewSeparator(),
		container.NewBorder(nil, nil, fieldLabel("Дата курса"), u.loadBtn, u.date),
	))

	win.SetContent(container.NewBorder(
		header(theme.ComputerIcon(), "Конвертер валют", "Официальный курс Национального банка Молдовы"),
		nil, nil, nil,
		container.NewVScroll(container.NewPadded(container.NewVBox(u.banner, input, output, rateInfo))),
	))
	win.Resize(fyne.NewSize(600, 760))
	u.refreshButton()
	return u
}

// setNotice shows a warning banner, or hides it for an empty text.
func (u *ui) setNotice(s string) {
	u.notice.SetText(s)
	showIf(u.banner, s != "")
	u.win.Content().Refresh() // relayout: the banner changes the column height
}

func (u *ui) setAmountErr(s string) {
	u.amountErr.SetText(s)
	showIf(u.amountErr, s != "")
	u.win.Content().Refresh()
}

func showIf(o fyne.CanvasObject, visible bool) {
	if visible {
		o.Show()
	} else {
		o.Hide()
	}
}

func (u *ui) setResult(s string) {
	if s == "" {
		s = "—"
	}
	u.result.Text = s
	u.result.Refresh()
}

// load requests rates for the date in the date field (today if userAction is false).
func (u *ui) load(userAction bool) {
	d, err := validate.ParseDate(u.date.Text)
	if err != nil {
		u.setNotice("Дата: " + err.Error())
		return
	}
	u.loadBtn.Disable()
	u.setNotice("")
	if !u.loaded {
		u.info.SetText("Загрузка курсов…")
	} else {
		u.info.SetText(u.infoText() + "\nЗагрузка курса на " + d.Format(bnm.DateLayout) + "…")
	}
	u.spawn(func() {
		ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
		defer cancel()
		res, err := u.loader.Load(ctx, d)
		fyne.Do(func() { u.applyLoad(res, err) })
	})
}

func (u *ui) applyLoad(res rates.Result, err error) {
	u.loadBtn.Enable()
	if err != nil {
		if u.loaded {
			u.info.SetText(u.infoText())
			u.setNotice(err.Error() + " Оставлены ранее загруженные курсы.")
		} else {
			u.info.SetText("Курсы не загружены.")
			u.setNotice(err.Error())
		}
		u.refreshButton()
		return
	}
	t := res.Table
	u.table = &t
	u.loaded = true

	codes := make([]string, 0, len(t.Rates))
	for c := range t.Rates {
		codes = append(codes, c)
	}
	sort.Strings(codes)
	// Popular currencies first.
	front := []string{"MDL", "EUR", "USD", "RON", "UAH", "RUB", "GBP"}
	ordered := []string{}
	for _, c := range front {
		if _, ok := t.Rates[c]; ok {
			ordered = append(ordered, c)
		}
	}
	for _, c := range codes {
		if !contains(front, c) {
			ordered = append(ordered, c)
		}
	}
	keep := func(s *widget.Select, def string) {
		cur := s.Selected
		s.Options = ordered
		if cur != "" && contains(ordered, cur) {
			s.Selected = cur
		} else if contains(ordered, def) {
			s.Selected = def
		} else {
			s.Selected = ""
		}
		s.Refresh()
	}
	keep(u.from, "EUR")
	keep(u.to, "MDL")

	u.info.SetText(u.infoText())
	u.setNotice(res.Notice)
	u.setResult("")
	u.rateLine.SetText("")
	u.refreshButton()
}

func (u *ui) infoText() string {
	if u.table == nil {
		return ""
	}
	u.source.SetText("Источник: " + u.table.Source)
	return fmt.Sprintf("Курс на %s", u.table.Date.Format(bnm.DateLayout))
}

func (u *ui) refreshButton() {
	if u.table != nil && strings.TrimSpace(u.amount.Text) != "" && u.from.Selected != "" && u.to.Selected != "" {
		u.convert.Enable()
	} else {
		u.convert.Disable()
	}
}

func (u *ui) onCurrencyChanged() {
	u.refreshButton()
	if u.result.Text != "—" && !u.convert.Disabled() {
		u.onConvert()
	}
}

func (u *ui) onSwap() {
	f, t := u.from.Selected, u.to.Selected
	// Set fields directly to avoid two intermediate conversions.
	u.from.Selected, u.to.Selected = t, f
	u.from.Refresh()
	u.to.Refresh()
	u.refreshButton()
	if !u.convert.Disabled() {
		u.onConvert()
	}
}

func (u *ui) onConvert() {
	amount, err := validate.ParseAmount(u.amount.Text)
	if err != nil {
		u.setAmountErr(err.Error())
		u.setResult("")
		u.rateLine.SetText("")
		return
	}
	if u.table == nil {
		u.setAmountErr("Курсы ещё не загружены")
		return
	}
	r, err := convert.Convert(amount, u.from.Selected, u.to.Selected, *u.table)
	if err != nil {
		u.setAmountErr(err.Error())
		return
	}
	u.setAmountErr("")
	u.setResult(fmt.Sprintf("%s %s = %s %s",
		formatMoney(amount), u.from.Selected, formatMoney(r.Amount), u.to.Selected))
	u.rateLine.SetText(fmt.Sprintf("1 %s = %.4f %s  ·  курс на %s",
		u.from.Selected, r.Rate, u.to.Selected, u.table.Date.Format(bnm.DateLayout)))
}

// formatMoney prints 2 decimals with a space as thousands separator.
func formatMoney(v float64) string {
	s := fmt.Sprintf("%.2f", v)
	intPart, frac := s[:len(s)-3], s[len(s)-3:]
	var b strings.Builder
	for i, c := range intPart {
		if i > 0 && (len(intPart)-i)%3 == 0 {
			b.WriteByte(' ')
		}
		b.WriteRune(c)
	}
	return b.String() + strings.Replace(frac, ".", ",", 1)
}

func contains(xs []string, s string) bool {
	for _, x := range xs {
		if x == s {
			return true
		}
	}
	return false
}
