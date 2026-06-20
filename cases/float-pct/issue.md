# folio: single `float:left` with percentage width (control — works)

**Status:** ✅ Control / passing

## Purpose

A **control** case: a single `float:left; width:50%` box followed by normal-flow
content. This renders correctly, confirming that a single percentage-width float
works on its own.

Contrast with [`float-left-right`](../float-left-right): adding a second
(`float:right`) column is what triggers the width-collapse and no-clear bug, so the
problem is in handling **multiple** floats / `float:right` / clearing — not in
percentage widths per se.

## How to reproduce

```sh
go run .          # renders cases/<name>/sample.html -> cases/<name>/output.pdf
```

## Observed in folio

Correct: the float is laid out and following content does not overlap it.

## Environment

- folio `v0.9.1` / `main` @ `72b6a6a`, Go 1.26, macOS arm64
