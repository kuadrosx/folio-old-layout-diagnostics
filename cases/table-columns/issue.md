# folio: `display:table` / `display:table-cell` columns stack vertically

**Status:** ❌ Bug

## Summary

A `display:table` row containing two `display:table-cell` columns should render the
cells side by side in a single row. In folio the cells stack **vertically** (each
at 50% width) instead of forming a horizontal row.

## How to reproduce

From the repository root:

```sh
go run .          # renders cases/<name>/sample.html -> cases/<name>/output.pdf
```

Then open `output.pdf` in this folder.

## Input

See [`sample.html`](./sample.html): a `display:table` `.row` with two
`display:table-cell` columns, followed by an `.after` block.

## Expected (per CSS)

The two table cells render side by side as a single table row; `Content after the
row` renders below them.

## Observed in folio

The two cells are stacked vertically (one above the other), each 50% wide, rather
than placed side by side.

## Likely cause

`display:table` / `table-cell` does not appear to be implemented as a horizontal
table-layout formatting context; the cells fall back to block stacking.

## Environment

- folio `v0.9.1`, also reproduces on `main` @ `72b6a6a`
- Go 1.26 (module declares `go 1.25.0`)
- macOS arm64 (not OS-specific)

## Workaround

Use `display:flex` (or CSS grid) for multi-column layouts — see
[`flex-columns`](../flex-columns).
