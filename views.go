package main

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

// card wraps content in a rounded panel.
func card(content fyne.CanvasObject) fyne.CanvasObject {
	bg := canvas.NewRectangle(theme.Color(colorCard))
	bg.CornerRadius = 12
	bg.StrokeColor = theme.Color(theme.ColorNameSeparator)
	bg.StrokeWidth = 1
	return container.NewStack(bg, container.New(layout.NewCustomPaddedLayout(12, 12, 16, 16), content))
}

// banner is a warning-tinted panel with an icon.
func banner(content fyne.CanvasObject) fyne.CanvasObject {
	bg := canvas.NewRectangle(theme.Color(colorBanner))
	bg.CornerRadius = 10
	return container.NewStack(bg, container.New(layout.NewCustomPaddedLayout(6, 6, 12, 12),
		container.NewBorder(nil, nil, widget.NewIcon(theme.WarningIcon()), nil, content)))
}

// header is the tinted title band at the top of the window.
func header(icon fyne.Resource, title, subtitle string) fyne.CanvasObject {
	bg := canvas.NewRectangle(theme.Color(colorHeader))
	t := canvas.NewText(title, theme.Color(theme.ColorNameForeground))
	t.TextSize = 22
	t.TextStyle = fyne.TextStyle{Bold: true}
	sub := widget.NewLabel(subtitle)
	iconBox := container.NewGridWrap(fyne.NewSize(36, 36), widget.NewIcon(icon))
	return container.NewStack(bg, container.New(layout.NewCustomPaddedLayout(14, 8, 16, 16),
		container.NewBorder(nil, nil, container.NewCenter(iconBox), nil, container.NewVBox(t, sub))))
}

func fieldLabel(s string) *widget.Label {
	l := widget.NewLabel(s)
	l.TextStyle = fyne.TextStyle{Bold: true}
	return l
}
