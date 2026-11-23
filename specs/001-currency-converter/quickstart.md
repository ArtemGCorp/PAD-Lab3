# Quickstart / Manual Validation

```bash
go run .          # start the app
go test ./...     # run unit tests (offline)
```

1. Start with network → currency lists are filled, info line shows today's BNM date.
2. Enter `100`, choose `EUR` → `MDL`, Convert → `100 × EUR rate`.
3. Press ⇄ → `MDL` → `EUR`, result is the reverse conversion.
4. Enter `-1`, `0`, `abc` → error text, no result.
5. Choose the same currency twice → result equals amount.
6. Enter date of a public holiday or a future date → message about fallback / future date.
7. Turn Wi-Fi off, restart → message "Нет соединения", cached date is shown, conversion works.
