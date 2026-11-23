# Contract: BNM official exchange rates XML

`GET https://www.bnm.md/en/official_exchange_rates?get_xml=1&date=DD.MM.YYYY`

```xml
<?xml version="1.0" encoding="UTF-8"?>
<ValCurs Date="20.11.2026" name="Official exchange rate">
  <Valute ID="47">
    <NumCode>978</NumCode>
    <CharCode>EUR</CharCode>
    <Nominal>1</Nominal>
    <Name>Euro</Name>
    <Value>20.1267</Value>
  </Valute>
  ...
</ValCurs>
```

Parsing rules:

1. Root element MUST be `ValCurs`; otherwise → `ErrUnavailable` (unexpected format).
2. `Date` attribute is parsed as `02.01.2006`; if missing, the requested date is used.
3. A `Valute` is accepted only if `CharCode` is 3 letters, `Nominal` ≥ 1, `Value` > 0
   (decimal separator `.` or `,`). Invalid items are skipped.
4. Zero accepted items → `ErrNoData`.
5. HTTP 404 or empty body → `ErrNoData`; network error, timeout, HTTP ≥ 500 → `ErrUnavailable`.
