package main

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
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
	result    *widget.Label
	rateLine  *widget.Label

	date    *widget.Entry
	loadBtn *widget.Button
	info    *widget.Label
	notice  *widget.Label

	table  *bnm.Table
	loaded bool
}

func newUI(win fyne.Window, loader *rates.Loader) *ui {
	u := &ui{win: win, loader: loader, spawn: func(f func()) { go f() }}

	u.amount = widget.NewEntry()
	u.amount.SetPlaceHolder("Сумма, например 100 или 12,50")
	u.amount.OnChanged = func(string) { u.amountErr.SetText(""); u.refreshButton() }
	u.amount.OnSubmitted = func(string) {
		if !u.convert.Disabled() {
			u.onConvert()
		}
	}
	u.amountErr = widget.NewLabel("")
	u.amountErr.Importance = widget.DangerImportance

	u.from = widget.NewSelect(nil, func(string) { u.onCurrencyChanged() })
	u.from.PlaceHolder = "Из валюты"
	u.to = widget.NewSelect(nil, func(string) { u.onCurrencyChanged() })
	u.to.PlaceHolder = "В валюту"
	u.swap = widget.NewButtonWithIcon("", theme.ViewRefreshIcon(), u.onSwap)

	u.convert = widget.NewButton("Конвертировать", u.onConvert)
	u.convert.Importance = widget.HighImportance

	u.result = widget.NewLabelWithStyle("", fyne.TextAlignCenter, fyne.TextStyle{Bold: true})
	u.rateLine = widget.NewLabelWithStyle("", fyne.TextAlignCenter, fyne.TextStyle{})

	u.date = widget.NewEntry()
	u.date.SetText(time.Now().Format(bnm.DateLayout))
	u.loadBtn = widget.NewButton("Загрузить курс", func() { u.load(true) })

	u.info = widget.NewLabel("Загрузка курсов…")
	u.info.Wrapping = fyne.TextWrapWord
	u.notice = widget.NewLabel("")
	u.notice.Wrapping = fyne.TextWrapWord
	u.notice.Importance = widget.WarningImportance

	currencies := container.NewBorder(nil, nil, nil, u.swap, container.NewGridWithColumns(2, u.from, u.to))
	form := widget.NewForm(
		widget.NewFormItem("Сумма", container.NewVBox(u.amount, u.amountErr)),
		widget.NewFormItem("Валюты", currencies),
		widget.NewFormItem("Дата курса", container.NewBorder(nil, nil, nil, u.loadBtn, u.date)),
	)
	win.SetContent(container.NewPadded(container.NewVBox(
		widget.NewLabelWithStyle("Конвертер валют · официальный курс НБМ", fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
		form,
		u.convert,
		widget.NewSeparator(),
		u.result,
		u.rateLine,
		widget.NewSeparator(),
		u.info,
		u.notice,
	)))
	win.Resize(fyne.NewSize(560, 480))
	u.refreshButton()
	return u
}

// load requests rates for the date in the date field (today if userAction is false).
func (u *ui) load(userAction bool) {
	d, err := validate.ParseDate(u.date.Text)
	if err != nil {
		u.notice.SetText("Дата: " + err.Error())
		return
	}
	u.loadBtn.Disable()
	u.notice.SetText("")
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
			u.notice.SetText(err.Error() + " Оставлены ранее загруженные курсы.")
		} else {
			u.info.SetText("Курсы не загружены.")
			u.notice.SetText(err.Error())
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
	u.notice.SetText(res.Notice)
	u.result.SetText("")
	u.rateLine.SetText("")
	u.refreshButton()
}

func (u *ui) infoText() string {
	if u.table == nil {
		return ""
	}
	return fmt.Sprintf("Курс на %s · Источник: %s", u.table.Date.Format(bnm.DateLayout), u.table.Source)
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
	if u.result.Text != "" && !u.convert.Disabled() {
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
		u.amountErr.SetText(err.Error())
		u.result.SetText("")
		u.rateLine.SetText("")
		return
	}
	if u.table == nil {
		u.amountErr.SetText("Курсы ещё не загружены")
		return
	}
	r, err := convert.Convert(amount, u.from.Selected, u.to.Selected, *u.table)
	if err != nil {
		u.amountErr.SetText(err.Error())
		return
	}
	u.amountErr.SetText("")
	u.result.SetText(fmt.Sprintf("%s %s = %s %s",
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
