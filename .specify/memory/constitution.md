# BNM Currency Converter — Constitution

## Core Principles

### I. Never Crash on External Failure
The application MUST stay alive and responsive when the network, the data source, the cache
file or user input is broken. Every failure is turned into a clear, user-facing message in
Russian. A `panic` reaching the user is a defect.

### II. Honest Data Provenance
Every converted value MUST be shown together with the date of the exchange rate used and the
data source. When the app uses anything other than a fresh rate for the requested date
(cached data, a previous business day), it MUST say so explicitly.

### III. Test-First Core Logic (NON-NEGOTIABLE)
Parsing, conversion, input validation and fallback logic live in GUI-free packages and are
covered by unit tests that run offline with one command: `go test ./...`. The network is
never touched by unit tests (use `httptest`).

### IV. Separation of Concerns
UI code only wires widgets to services. Domain packages (`bnm`, `convert`, `validate`,
`cache`, `rates`) do not import the GUI toolkit.

### V. Simplicity
Standard library first. One external dependency for the GUI (Fyne). No database: the cache is
one JSON file. No speculative features (YAGNI).

## Technology Constraints

- Language: Go 1.26+. GUI: Fyne v2. Desktop (macOS primary, Windows/Linux supported by Fyne).
- Money arithmetic uses `float64` with rounding only at display time (4 decimals for rates,
  2 decimals for amounts); this is acceptable for an informational converter.
- All user-facing text is in Russian; code identifiers and comments are in English.

## Development Workflow

- One git branch per stage: specification → implementation → tests → report.
- Specifications are committed before any implementation code.
- Merges into `main` go through Pull Requests.
- Commit messages describe the change, not the stage ("Add BNM XML parser", not "wip").

## Governance

This constitution supersedes ad-hoc decisions. Any deviation discovered during implementation
is recorded in `REPORT.md` (section "Где агент ошибся").

**Version**: 1.0.0 | **Ratified**: 2026-11-24 | **Last Amended**: 2026-11-24
