# Feature Specification: Desktop Currency Converter (BNM rates)

**Feature Branch**: `001-currency-converter`
**Created**: 2026-11-24
**Status**: Ready for planning
**Input**: Lab 3 task — GUI currency converter with official National Bank of Moldova rates,
offline cache, weekend/holiday handling, input validation.

## User Scenarios & Testing *(mandatory)*

### User Story 1 — Convert an amount (Priority: P1)

A user enters an amount, selects a source and a target currency and presses "Конвертировать".
The app shows the converted amount, the rate date and the data source.

**Why this priority**: this is the whole purpose of the app; everything else supports it.

**Independent Test**: start the app with network access, enter `100`, choose `EUR` → `MDL`,
press the button; the result equals `100 × EUR rate` of the shown date.

**Acceptance Scenarios**:

1. **Given** rates are loaded, **When** the user enters `100`, selects `USD` → `MDL` and
   presses Convert, **Then** the result `100 × USD/MDL` is shown with 2 decimals.
2. **Given** rates are loaded, **When** the user selects `MDL` → `USD`, **Then** the result is
   `amount ÷ USD/MDL` (reverse direction).
3. **Given** rates are loaded, **When** the user selects `EUR` → `USD`, **Then** the cross rate
   via MDL is used: `amount × EUR/MDL ÷ USD/MDL`.
4. **Given** a result is shown, **When** the user presses "⇄" (swap), **Then** source and target
   currencies are exchanged and the result is recalculated.
5. **Given** the same currency is chosen in both lists, **When** the user converts, **Then** the
   result equals the amount and no error is shown.

### User Story 2 — Input validation (Priority: P1)

**Acceptance Scenarios**:

1. **Given** the amount field is empty or a currency is not selected, **Then** the Convert button
   is disabled.
2. **Given** the amount is `0`, `-5`, `abc`, `1e`, `NaN` or `12,3,4`, **When** the user presses
   Convert, **Then** a clear error is shown under the field and no conversion happens.
3. **Given** the amount uses a comma (`12,50`) or surrounding spaces, **Then** it is accepted as
   `12.50`.

### User Story 3 — Choose the rate date (Priority: P2)

The user may enter a date `ДД.ММ.ГГГГ` (default: today) and press "Загрузить курс".

**Acceptance Scenarios**:

1. **Given** an invalid date (`31.02.2026`, `2026-01-01`, empty), **Then** a validation message
   is shown and nothing is requested.
2. **Given** a date in the future, **Then** the app reports that rates for the future are not
   available and keeps the current rates.
3. **Given** the source publishes no rate for the chosen date (holiday, weekend, before
   publication), **Then** the app walks back day by day (up to 10 days) to the last published
   rate, uses it and tells the user: "Курс на DD.MM.YYYY не опубликован, используется курс от
   DD.MM.YYYY".

### User Story 4 — Work offline (Priority: P1)

**Acceptance Scenarios**:

1. **Given** no network and a cached rate exists, **Then** the app does not crash, shows
   "Нет соединения с bnm.md" and offers to work with the cached rate, showing its date.
2. **Given** no network and no cache, **Then** the app shows a clear message, the currency lists
   are empty and the Convert button stays disabled; the user can retry later.
3. **Given** a successful download, **When** the app is restarted without network, **Then** the
   last successful rates are available (the cache survives restart).

### Edge Cases

- The source answers HTTP 200 with an empty `<ValCurs>` or with HTML instead of XML.
- The XML contains a malformed number for one currency → that currency is skipped, the rest
  are used.
- A currency has `Nominal` > 1 (e.g. 10 or 100 units) → the per-unit rate is `Value / Nominal`.
- The cache file is corrupted → it is ignored (treated as absent) and overwritten on the next
  success.
- A very large amount (> 1e12) → rejected as unrealistic.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: The app MUST have a graphical user interface (desktop window). Console UI is not
  acceptable.
- **FR-002**: The UI MUST contain: amount input, source currency select, target currency select,
  a Convert button, a result area, the rate date and the data source.
- **FR-003**: The Convert button MUST be disabled until the amount field is non-empty and both
  currencies are selected.
- **FR-004**: The app MUST use official BNM rates from
  `https://www.bnm.md/en/official_exchange_rates?get_xml=1&date=DD.MM.YYYY`.
- **FR-005**: MDL MUST be available as a currency with rate 1.
- **FR-006**: Conversion MUST work in both directions and between two foreign currencies via MDL.
- **FR-007**: Same source and target currency MUST return the amount unchanged.
- **FR-008**: Invalid input (empty, zero, negative, letters, non-finite, > 1e12) MUST be rejected
  with a message; no conversion is performed.
- **FR-009**: When the source has no rate for the requested date, the app MUST use the last
  available earlier rate (max 10 days back) and tell the user which date is used.
- **FR-010**: Network failure MUST NOT crash the app; the app MUST show a message and, if
  available, use the cached rate with its date.
- **FR-011**: The last successful response MUST be stored locally (user cache directory) and
  survive restarts.
- **FR-012**: Results MUST display amounts with 2 decimals and rates with 4 decimals.

### Key Entities

- **RateTable**: date of the rate, source name, map currency code → `Rate`.
- **Rate**: code (ISO 4217), name, nominal, value in MDL for `nominal` units.
- **CacheEntry**: `RateTable` + time it was fetched.
- **LoadResult**: `RateTable` + flags `FromCache`, `Fallback` + the requested date + message.

## Success Criteria *(mandatory)*

- **SC-001**: A user converts an amount in ≤ 3 interactions after start (type, choose, click).
- **SC-002**: With the network cable unplugged the app starts, shows cached data and never
  terminates unexpectedly.
- **SC-003**: All unit tests pass with `go test ./...` without network access.
- **SC-004**: For any date between 01.01.2020 and today the app shows a rate (exact or the last
  earlier one) and its true date.

## Assumptions

- BNM is the only source; it is chosen because it is the official MDL rate, free, needs no API
  key and the task explicitly references it.
- A 404 or an empty `ValCurs` from BNM means "no data for that date", while a transport error
  or HTTP 5xx means "source unavailable".
