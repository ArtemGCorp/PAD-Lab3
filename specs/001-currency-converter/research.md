# Research: Desktop Currency Converter

## Decision 1 — Data source: BNM XML

- **Decision**: official XML feed of the National Bank of Moldova.
- **Rationale**: official MDL rate (legally used in Moldova), free, no API key, no quota, one
  request returns all ~45 currencies, historical dates supported via `date=DD.MM.YYYY`.
- **Alternatives**: free JSON APIs (exchangerate.host, frankfurter.app) — need a key or use ECB
  rates where MDL is missing or not official; HTML scraping of bnm.md — fragile, the XML feed
  already exists.

Observed behaviour (checked 2026-11-24):

| Request | Response |
|---|---|
| business day | `200`, `<ValCurs Date="…">` with `<Valute>` items |
| weekend (22.11.2026) | `200`, the rate valid for that day (same values as Friday) |
| future date (01.01.2030) | `404`, empty body |
| invalid date (31.02.2026, `abc`) | `404`, empty body |

## Decision 2 — GUI toolkit: Fyne

- **Rationale**: pure Go API, cross-platform, has a headless test driver (`fyne/test`) so UI
  behaviour is unit-testable. Same language as Lab 1 and Lab 2.
- **Alternatives**: Wails (needs a web front-end, second language), Gio (lower level).

## Decision 3 — Cache: JSON file in the user cache dir

- **Rationale**: survives restarts, human-readable, atomic write via tmp + rename.
- **Alternatives**: SQLite — overkill for one table.

## Decision 4 — Error classification

- 404 / empty body / `ValCurs` without valid `Valute` → `ErrNoData` (try an earlier date).
- Transport error, timeout, HTTP 5xx, non-XML body → `ErrUnavailable` (go to cache).
