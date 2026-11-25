package main

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
)

// appTheme: teal accent, softer background, rounder corners, a bit more air.
// Light and dark variants both follow the system setting.
type appTheme struct{}

var _ fyne.Theme = appTheme{}

var (
	accent      = color.NRGBA{R: 0x0F, G: 0x76, B: 0x6E, A: 0xFF} // teal-700
	accentDark  = color.NRGBA{R: 0x2D, G: 0xD4, B: 0xBF, A: 0xFF} // teal-400
	successCol  = color.NRGBA{R: 0x05, G: 0x96, B: 0x69, A: 0xFF}
	warningCol  = color.NRGBA{R: 0xD9, G: 0x77, B: 0x06, A: 0xFF}
	bgLight     = color.NRGBA{R: 0xF3, G: 0xF7, B: 0xF6, A: 0xFF}
	cardLight   = color.NRGBA{R: 0xFF, G: 0xFF, B: 0xFF, A: 0xFF}
	bgDark      = color.NRGBA{R: 0x10, G: 0x1A, B: 0x1A, A: 0xFF}
	cardDark    = color.NRGBA{R: 0x18, G: 0x26, B: 0x26, A: 0xFF}
	headerLight = color.NRGBA{R: 0xCC, G: 0xFB, B: 0xF1, A: 0xFF}
	headerDark  = color.NRGBA{R: 0x13, G: 0x3B, B: 0x38, A: 0xFF}
)

func (appTheme) Color(n fyne.ThemeColorName, v fyne.ThemeVariant) color.Color {
	dark := v == theme.VariantDark
	switch n {
	case theme.ColorNamePrimary, theme.ColorNameFocus, theme.ColorNameHyperlink:
		if dark {
			return accentDark
		}
		return accent
	case theme.ColorNameBackground:
		if dark {
			return bgDark
		}
		return bgLight
	case theme.ColorNameOverlayBackground, theme.ColorNameMenuBackground:
		if dark {
			return cardDark
		}
		return cardLight
	case colorHeader:
		if dark {
			return headerDark
		}
		return headerLight
	case colorCard:
		if dark {
			return cardDark
		}
		return cardLight
	case colorBanner:
		if dark {
			return color.NRGBA{R: 0x42, G: 0x34, B: 0x0C, A: 0xFF}
		}
		return color.NRGBA{R: 0xFE, G: 0xF3, B: 0xC7, A: 0xFF}
	case theme.ColorNameSuccess:
		return successCol
	case theme.ColorNameWarning:
		return warningCol
	}
	return theme.DefaultTheme().Color(n, v)
}

// Custom colour names, resolved through the theme so they follow the variant.
const (
	colorHeader fyne.ThemeColorName = "lab3Header"
	colorCard   fyne.ThemeColorName = "lab3Card"
	colorBanner fyne.ThemeColorName = "lab3Banner"
)

// fixedVariant forces a light or dark variant (used for screenshots in tests).
type fixedVariant struct {
	fyne.Theme
	v fyne.ThemeVariant
}

func (f fixedVariant) Color(n fyne.ThemeColorName, _ fyne.ThemeVariant) color.Color {
	return f.Theme.Color(n, f.v)
}

func (appTheme) Font(s fyne.TextStyle) fyne.Resource     { return theme.DefaultTheme().Font(s) }
func (appTheme) Icon(n fyne.ThemeIconName) fyne.Resource { return theme.DefaultTheme().Icon(n) }

func (appTheme) Size(n fyne.ThemeSizeName) float32 {
	switch n {
	case theme.SizeNameText:
		return 15
	case theme.SizeNameHeadingText:
		return 24
	case theme.SizeNamePadding:
		return 6
	case theme.SizeNameInnerPadding:
		return 10
	case theme.SizeNameInputRadius, theme.SizeNameSelectionRadius, theme.SizeNameButtonRadius:
		return 8
	}
	return theme.DefaultTheme().Size(n)
}
