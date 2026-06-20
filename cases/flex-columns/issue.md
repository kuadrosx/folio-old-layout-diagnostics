# folio: two-column layout via `display:flex` (control / workaround — works)

**Status:** ✅ Control / passing — recommended workaround

## Purpose

The same two 50% columns as the failing float/table cases, expressed with
`display:flex` instead. This renders **correctly** — two columns side by side with
following content below — and is the recommended workaround for the float and
`display:table` bugs.

Compare the failing equivalents:

- [`float-left-right`](../float-left-right) — `float:left` + `float:right` (❌)
- [`float-columns`](../float-columns) — floats + `clear:both` (❌)
- [`table-columns`](../table-columns) — `display:table-cell` (❌)
- [`float-in-bfc`](../float-in-bfc) — floats in `overflow:hidden` (❌)

## How to reproduce

```sh
go run .          # renders cases/<name>/sample.html -> cases/<name>/output.pdf
```

## Observed in folio

Correct: two side-by-side 50% columns, `Content after` below both. CSS grid also
works.

## Environment

- folio `v0.9.1` / `main` @ `72b6a6a`, Go 1.26, macOS arm64
