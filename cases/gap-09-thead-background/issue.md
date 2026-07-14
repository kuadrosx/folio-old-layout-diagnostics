Gap: FOLIO-GAP-09

# folio: `background-color` on `<thead>` is not painted

**Status:** ❌ Bug — header background missing.

## Summary

`background-color` set on a `<thead>` element is not painted. Chrome fills the
header row with the colour; folio leaves it transparent (it only honors the
background when set on `thead td`/`th`).

## Minimal repro

`sample.html` — a table with `thead { background-color:#cccccc }`.

```sh
go run .   # renders sample.html -> output.pdf
```

## Expected vs observed

- **Expected (Chrome):** a solid grey bar across the header row.
- **Observed (folio):** no header fill; the `<thead>` background is ignored.

## Source pointers

- `html/converter_table.go` (`case atom.Thead`) — the `<thead>` section group's
  own `background-color` is not carried onto a painted background rect; only
  cell-level (`td`/`th`) backgrounds are painted.

## Environment

- folio `v0.10.0-1-g1b17d01` (one commit past the v0.10.0 tag), Go 1.26, macOS arm64.

## Workaround

Apply the background to `thead td` / `thead th` instead of `thead`.
