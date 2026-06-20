# folio: single `float:left` with fixed px width (control — works)

**Status:** ✅ Control / passing

## Purpose

This is a **control** case included to isolate the float bugs. A single
`float:left; width:200px` box followed by normal-flow content renders correctly:
the following content wraps to the right of (and below) the float and does not
overlap it.

It demonstrates that folio's float handling works for the simplest single-float
case, which narrows the failing cases down to: multiple floats, `float:right`
alignment, float clearing, and float containment. See:

- [`float-pct`](../float-pct) — single float with `%` width (also works)
- [`float-left-right`](../float-left-right) — two floats (❌ bug)
- [`float-columns`](../float-columns) — two floats + `clear:both` (❌ bug)
- [`float-in-bfc`](../float-in-bfc) — floats in an `overflow:hidden` BFC (❌ bug)

## How to reproduce

```sh
go run .          # renders cases/<name>/sample.html -> cases/<name>/output.pdf
```

## Observed in folio

Correct: following content does not overlap the float.

## Environment

- folio `v0.9.1` / `main` @ `72b6a6a`, Go 1.26, macOS arm64
