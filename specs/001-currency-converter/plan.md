# Implementation Plan: Desktop Currency Converter (BNM rates)

**Branch**: `001-currency-converter` | **Date**: 2026-11-24 | **Spec**: [spec.md](spec.md)

## Summary

A Go desktop app (Fyne) that downloads official BNM XML rates, converts amounts between any two
currencies via MDL, caches the last successful table as JSON, falls back to earlier dates when
BNM has no data and to the cache when the network is down.

## Technical Context

**Language/Version**: Go 1.26
**Primary Dependencies**: `fyne.io/fyne/v2` (GUI); standard library for HTTP, XML, JSON
**Storage**: JSON file `<os.UserCacheDir()>/pad-lab3/rates.json`
**Testing**: `go test ./...` (`testing`, `net/http/httptest`, `fyne.io/fyne/v2/test`)
**Target Platform**: desktop (macOS primary)
**Project Type**: single desktop application
**Performance Goals**: rate load ≤ 10 s (HTTP timeout), conversion instantaneous
**Constraints**: must work offline with cache; no API keys
**Scale/Scope**: 1 window, ~45 currencies

## Constitution Check

| Principle | How the plan complies |
|---|---|
| I. Never crash | All I/O returns errors; UI shows them; goroutine results are delivered with `fyne.Do` |
| II. Provenance | `LoadResult` always carries rate date, source and fallback/cache flags; UI prints them |
| III. Test-first core | `bnm`, `convert`, `validate`, `cache`, `rates` are pure Go and tested offline |
| IV. Separation | Only `main.go` / `ui.go` import Fyne |
| V. Simplicity | One dependency; JSON cache; no DB |

Result: **PASS**.

## Project Structure

### Documentation (this feature)

```
specs/001-currency-converter/
├── spec.md
├── plan.md
├── research.md
├── data-model.md
├── contracts/bnm-xml.md
├── quickstart.md
├── checklists/requirements.md
└── tasks.md
```

### Source Code (repository root)

```
main.go                    entry point: builds services and the window
ui.go                      Fyne widgets and event wiring
ui_test.go                 UI behaviour with fyne/test
internal/bnm/              XML parsing + HTTP client (ErrNoData vs ErrUnavailable)
internal/convert/          conversion math
internal/validate/         amount and date validation
internal/cache/            JSON cache with atomic save
internal/rates/            load strategy: exact date → earlier dates → cache
```

**Structure Decision**: single Go module `pad-lab3`, domain logic under `internal/`.

## Load Strategy (`internal/rates`)

```
Load(date):
  if date > today: return ErrFutureDate
  for d := date; d >= date-10 days; d -= 1 day:
      table, err := client.Fetch(d)
      if err == nil:  cache.Save(table); return {table, Fallback: d != date}
      if err is ErrNoData: continue
      if err is ErrUnavailable: break → go to cache
  cached := cache.Load()
  if cached ok: return {cached, FromCache: true, message: "Нет соединения…"}
  return error "нет сети и нет сохранённых курсов"
```

## Complexity Tracking

No constitution violations.
