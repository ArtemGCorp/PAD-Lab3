// Command pad-lab3 is a desktop currency converter using official
// National Bank of Moldova rates.
package main

import (
	"time"

	"fyne.io/fyne/v2/app"

	"pad-lab3/internal/bnm"
	"pad-lab3/internal/cache"
	"pad-lab3/internal/rates"
)

func main() {
	a := app.NewWithID("md.pad.lab3")
	a.Settings().SetTheme(appTheme{})
	w := a.NewWindow("Конвертер валют (НБМ)")
	loader := &rates.Loader{
		Fetcher: bnm.NewClient(),
		Cache:   cache.Store{Path: cache.DefaultPath()},
		Now:     time.Now,
	}
	u := newUI(w, loader)
	w.Show()
	u.load(false)
	a.Run()
}
