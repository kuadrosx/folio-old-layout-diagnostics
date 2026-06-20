# folio: plain block stacking (control — works)

**Status:** ✅ Control / passing

## Purpose

Baseline control: three normal-flow block `<div>`s. They stack vertically as
expected. This confirms that basic block flow is correct and that the failing
cases are specifically about `float` and `display:table` layout, not general
rendering.

## How to reproduce

```sh
go run .          # renders cases/<name>/sample.html -> cases/<name>/output.pdf
```

## Observed in folio

Correct: three stacked blocks.

## Environment

- folio `v0.9.1` / `main` @ `72b6a6a`, Go 1.26, macOS arm64
