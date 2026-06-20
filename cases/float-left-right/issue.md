# folio: `float:left` + `float:right` columns collapse and don't clear

**Status:** ❌ Bug

## Summary

Two sibling block elements at the `<body>` level — one `float:left; width:50%`,
the other `float:right; width:50%` — are not laid out as two 50%-wide columns.
Their explicit `width` is ignored (they shrink to content width), `float:right`
is **not** aligned to the container's right edge (it packs immediately next to the
left float), and the normal-flow block that follows renders **beside** the floats
instead of dropping below them.

## How to reproduce

From the repository root:

```sh
go run .          # renders cases/<name>/sample.html -> cases/<name>/output.pdf
```

Then open `output.pdf` in this folder. Or render `sample.html` directly with
`document.AddHTMLWithContext` on an A4 page.

## Input

See [`sample.html`](./sample.html): a `float:left;50%` div, a `float:right;50%`
div, and a normal-flow div after them.

## Expected (per CSS)

- Left column occupies the left 50% of the page.
- Right column occupies the right 50%, aligned to the right edge.
- `Content after` drops **below** both columns (its top is below the taller float).

## Observed in folio

- Both floats shrink to their content width; the `width:50%` is not applied.
- `float:right` sits directly next to the left float rather than at the right edge.
- `Content after` is laid out **beside** the floats, not below them.

## Likely cause (source pointers)

- `layout/float.go:67-74` — the float's width is derived from its content
  (`b.X + b.Width`); an explicit CSS `width` on the floated box is not honored on
  this path, which is why the 50% columns collapse to content width.
- `layout/float.go:97` — a float reports `Consumed: 0`, i.e. it adds no vertical
  space to the normal flow. `layout.Div` compensates for this (see
  `TestDivClearFloat` in `layout/div_test.go`), but...
- `html/converter.go:658` — `ConvertFull` returns the `<body>` children as a
  **flat element list**, and the document/page flow that lays that list out does
  not appear to establish a float context. With no float context, following
  content doesn't see the floats' height and renders beside them.

## Environment

- folio `v0.9.1`, also reproduces on `main` @ `72b6a6a` (`v0.9.1-4-g72b6a6a`)
- Go 1.26 (module declares `go 1.25.0`)
- macOS arm64 (not OS-specific)

## Workaround

Use `display:flex` (or CSS grid) for multi-column layouts — both render
correctly. See the sibling [`flex-columns`](../flex-columns) case.
