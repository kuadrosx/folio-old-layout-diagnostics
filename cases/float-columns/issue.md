# folio: `clear:both` after floated columns does not clear — content overlaps

**Status:** ❌ Bug (severe — content is lost / overlapped)

## Summary

This is the canonical "two floated columns inside a row, then a `clear:both`
spacer, then following content" layout used by many document/report headers.
In folio the `clear:both` div establishes no clearance: the `Content after the
row` block paints at the **top of the page, overlapping the float area**, and
most of the floated column content is visually lost.

## How to reproduce

From the repository root:

```sh
go run .          # renders cases/<name>/sample.html -> cases/<name>/output.pdf
```

Then open `output.pdf` in this folder.

## Input

See [`sample.html`](./sample.html): `.row` containing a `float:left;50%` column,
a `float:right;50%` column, and a `clear:both` div, followed by an `.after` block.

## Expected (per CSS)

Two 50% columns side by side; the `clear:both` div drops below the taller float;
`Content after the row` renders **below** both columns.

## Observed in folio

`Content after the row` is painted near the top of the page, overlapping the
floated columns. The `clear:both` div has no region to clear against, so it
introduces no vertical gap. Most of the float content is obscured.

## Likely cause (source pointers)

- `layout/float.go:97` — floats report `Consumed: 0`, contributing no height to
  the normal flow.
- `html/converter.go:658` — `ConvertFull` returns `<body>` children as a flat
  element list; the document/page flow laying it out does not establish a float
  context, so `clear` has no float extent to clear against.
- `html/converter.go:1538-1549` — when an element has `float:left/right`, **each
  child element** is wrapped in its own `layout.NewFloat(side, e)` rather than the
  element being floated as a single block, which compounds the height/clearance
  miscalculation.

## Environment

- folio `v0.9.1`, also reproduces on `main` @ `72b6a6a`
- Go 1.26 (module declares `go 1.25.0`)
- macOS arm64 (not OS-specific)

## Workaround

Use `display:flex` (or CSS grid) — see the sibling [`flex-columns`](../flex-columns)
case, which renders the same two columns correctly with following content below.
