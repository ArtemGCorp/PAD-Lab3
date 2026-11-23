// Package rates implements the load strategy from plan.md:
// exact date → earlier dates (max MaxBack days) → cache.
package rates

import (
	"context"
	"errors"
	"fmt"
	"time"

	"pad-lab3/internal/bnm"
	"pad-lab3/internal/cache"
)

// MaxBack is how many earlier days are tried when the source has no data.
const MaxBack = 10

// ErrFutureDate — rates for future dates do not exist.
var ErrFutureDate = errors.New("курс на будущую дату ещё не опубликован")

// Fetcher downloads a rate table (implemented by *bnm.Client).
type Fetcher interface {
	Fetch(ctx context.Context, date time.Time) (bnm.Table, error)
}

// Result of a load.
type Result struct {
	Table     bnm.Table
	Requested time.Time
	Fallback  bool // the table is for an earlier date than requested
	FromCache bool // the source was unavailable; the table is from the cache
	Notice    string
}

// Loader combines a fetcher and a cache.
type Loader struct {
	Fetcher Fetcher
	Cache   cache.Store
	Now     func() time.Time
}

func day(t time.Time) time.Time {
	y, m, d := t.Date()
	return time.Date(y, m, d, 0, 0, 0, 0, time.Local)
}

// Load returns rates for date following the strategy.
func (l *Loader) Load(ctx context.Context, date time.Time) (Result, error) {
	date = day(date)
	if date.After(day(l.Now())) {
		return Result{}, ErrFutureDate
	}

	var lastErr error
	for i := 0; i <= MaxBack; i++ {
		d := date.AddDate(0, 0, -i)
		t, err := l.Fetcher.Fetch(ctx, d)
		if err == nil {
			res := Result{Table: t, Requested: date}
			if !sameDay(t.Date, date) {
				res.Fallback = true
				res.Notice = fmt.Sprintf("Курс на %s не опубликован, используется курс от %s.",
					date.Format(bnm.DateLayout), t.Date.Format(bnm.DateLayout))
			}
			_ = l.Cache.Save(cache.Entry{FetchedAt: l.Now(), Table: t}) // cache failure is not fatal
			return res, nil
		}
		lastErr = err
		if !errors.Is(err, bnm.ErrNoData) {
			break // source unavailable → cache
		}
	}

	if errors.Is(lastErr, bnm.ErrNoData) {
		return Result{}, fmt.Errorf("источник не опубликовал курс ни за %s, ни за %d предыдущих дней",
			date.Format(bnm.DateLayout), MaxBack)
	}

	e, err := l.Cache.Load()
	if err != nil {
		return Result{}, fmt.Errorf("нет соединения с bnm.md и нет сохранённых курсов. Проверьте сеть и попробуйте ещё раз (%v)", lastErr)
	}
	return Result{
		Table:     e.Table,
		Requested: date,
		FromCache: true,
		Fallback:  !sameDay(e.Table.Date, date),
		Notice: fmt.Sprintf("Нет соединения с bnm.md. Используется сохранённый курс от %s (загружен %s).",
			e.Table.Date.Format(bnm.DateLayout), e.FetchedAt.Local().Format("02.01.2006 15:04")),
	}, nil
}

func sameDay(a, b time.Time) bool {
	ay, am, ad := a.Date()
	by, bm, bd := b.Date()
	return ay == by && am == bm && ad == bd
}
