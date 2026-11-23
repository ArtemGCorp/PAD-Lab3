# Data Model

## Rate

| Field | Type | Rule |
|---|---|---|
| Code | string | ISO 4217, 3 upper-case letters |
| Name | string | from `<Name>` |
| Nominal | int | ≥ 1 |
| Value | float64 | > 0, MDL for `Nominal` units |

Derived: `PerUnit() = Value / Nominal`.

## RateTable

| Field | Type | Rule |
|---|---|---|
| Date | time.Time | date of the rate (from `ValCurs@Date`) |
| Source | string | `"Национальный банк Молдовы (bnm.md)"` |
| Rates | map[string]Rate | always contains `MDL` with Nominal 1, Value 1 |

## CacheEntry (JSON)

```json
{ "fetched_at": "2026-11-24T10:00:00Z", "table": { "date": "2026-11-24", "source": "...", "rates": { "EUR": {...} } } }
```

## LoadResult

| Field | Meaning |
|---|---|
| Table | rate table to use |
| Requested | date asked by the user |
| Fallback | true if `Table.Date` ≠ `Requested` because of missing data |
| FromCache | true if the table came from the cache because the source is unavailable |
| Notice | Russian message to show to the user (empty if exact fresh data) |
