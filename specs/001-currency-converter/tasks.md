# Tasks: Desktop Currency Converter (BNM rates)

**Input**: design documents from `specs/001-currency-converter/`
**Prerequisites**: plan.md, spec.md, research.md, data-model.md, contracts/

Format: `[ID] [P?] [Story] Description` — `[P]` = can run in parallel (different files).

## Phase 1: Setup

- [ ] T001 Create Go module `pad-lab3` with `go.mod`, add `fyne.io/fyne/v2`
- [ ] T002 [P] Add `.gitignore` (binary, `.DS_Store`)

## Phase 2: Foundational (blocks all stories)

- [ ] T003 [P] `internal/bnm/model.go`: `Rate`, `RateTable`, `PerUnit()`, MDL entry
- [ ] T004 [P] `internal/bnm/parse.go`: `Parse([]byte, requested time.Time)` per contracts/bnm-xml.md
- [ ] T005 `internal/bnm/client.go`: `Client.Fetch(ctx, date)` with 10 s timeout, base URL injectable, error classes `ErrNoData` / `ErrUnavailable`
- [ ] T006 [P] `internal/cache/cache.go`: `Save` (tmp + rename) / `Load` (corrupted → absent)

## Phase 3: User Story 1 — Convert (P1) 🎯 MVP

- [ ] T007 [P] [US1] `internal/convert/convert.go`: `Convert(amount, from, to, table)` via MDL; same currency → amount; unknown currency → error
- [ ] T008 [US1] `ui.go`: amount entry, two selects, ⇄ swap, Convert button, result label, info label (date + source)
- [ ] T009 [US1] `main.go`: wire client, cache, loader, window

## Phase 4: User Story 2 — Validation (P1)

- [ ] T010 [P] [US2] `internal/validate/amount.go`: `ParseAmount` (comma/dot, trims, rejects ≤ 0, letters, non-finite, > 1e12)
- [ ] T011 [US2] Button enabled only when amount non-empty and both currencies selected; error label under the field

## Phase 5: User Story 3 — Date and fallback (P2)

- [ ] T012 [P] [US3] `internal/validate/date.go`: `ParseDate("DD.MM.YYYY")` with real-calendar check
- [ ] T013 [US3] `internal/rates/loader.go`: exact date → up to 10 earlier days → notice text
- [ ] T014 [US3] Date entry + "Загрузить курс" button in `ui.go`

## Phase 6: User Story 4 — Offline (P1)

- [ ] T015 [US4] Loader falls back to cache on `ErrUnavailable`, notice "Нет соединения… курс от DD.MM.YYYY"
- [ ] T016 [US4] Initial load in a goroutine; UI updates through `fyne.Do`; no cache + no network → message, controls disabled, retry possible

## Phase 7: Tests (Constitution III)

- [ ] T017 [P] `internal/bnm/parse_test.go`: valid XML, nominal > 1, comma decimals, invalid items skipped, empty `ValCurs`, HTML body
- [ ] T018 [P] `internal/bnm/client_test.go`: `httptest` 200 / 404 / 500 / timeout
- [ ] T019 [P] `internal/convert/convert_test.go`: forward, reverse, cross, same currency, round-trip
- [ ] T020 [P] `internal/validate/*_test.go`: table tests for amounts and dates
- [ ] T021 [P] `internal/cache/cache_test.go`: save/load round-trip, corrupted file, missing file
- [ ] T022 `internal/rates/loader_test.go`: exact, weekend fallback, offline → cache, offline no cache, future date
- [ ] T023 `ui_test.go`: button state, conversion, swap, invalid input message

## Phase 8: Polish

- [ ] T024 README.md: run command, test command, behaviour table
- [ ] T025 REPORT.md: source choice, models used, what worked, agent mistakes, spec conformance

## Dependencies

- T003 → T004 → T005; T006 independent
- Phase 2 → Phases 3–6; T007 and T010 → T008 → T011
- T013 → T015 → T016
- Tests (Phase 7) may be written alongside their units
