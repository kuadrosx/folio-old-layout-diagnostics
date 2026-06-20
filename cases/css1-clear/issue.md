# folio: `clear: left | right | both | none` against floats

**Status:** ❌ Bug

Adapted from the W3C CSS1 Test Suite, [section 5.5.26 "clear"](https://www.w3.org/Style/CSS/Test/CSS1/current/test5526.htm).
The original floats `<img>` elements; here each is a self-contained CSS block so
the case needs no external assets. This case isolates the four `clear` values,
which the existing [`float-columns`](../float-columns) case only exercises in
bulk.

## Summary

The test pairs floated orange blocks with paragraphs that carry `clear:left`,
`clear:right`, `clear:both`, and `clear:none`. Each cleared paragraph should drop
**below** the relevant float(s); the `clear:none` paragraph should sit beside
them. folio gives the wrong result on two counts: the leading text never flows
past the floats (it starts at the left margin, overlapping them — same root cause
as [`css1-float-textflow`](../css1-float-textflow)), and the cleared paragraphs
are not reliably positioned relative to the floats. The `clear:right` paragraph
is additionally shifted **left of the normal text margin**.

## How to reproduce

```sh
go run .          # renders cases/css1-clear/sample.html -> output.pdf
```

Then open `output.pdf` in this folder.

## Input

See [`sample.html`](./sample.html): five paragraphs, each preceded by one or two
floated orange blocks, classed `clear:left` / `right` / `both` / `none`.

## Expected (per CSS)

- `clear:left` paragraph drops below the left float.
- `clear:right` paragraph drops below the right float.
- `clear:both` paragraph drops below both floats and does not flow between them.
- `clear:none` paragraph sits between the two floats.

## Observed in folio

Measured from the rendered PDF (A4, normal text content origin x ≈ 125pt):

- The first (uncleared) paragraph starts at **x ≈ 125pt**, i.e. it does not flow
  past the left float — same no-horizontal-space failure as `css1-float-textflow`.
- The `clear:right` paragraph renders at **x ≈ 72–105pt**, *left of* the normal
  125pt text margin — a horizontal positioning glitch unique to this branch.
- Paragraphs do stack vertically with line-break spacing, but because floats
  report `Consumed: 0` there is no float height to clear against, so the vertical
  positions are coincidental rather than driven by the floats. Open the PDF to
  confirm visually (the floats are empty blocks and carry no text).

## Likely cause (source pointers)

- `layout/float.go:97` — a float reports `Consumed: 0`, so there is no float
  height for `clear` to push past.
- `html/converter.go:658` — `ConvertFull` returns `<body>` children as a flat
  element list with no float context; `clear` has nothing to clear against.
- `layout.Div` compensates for `Consumed: 0` (see `TestDivClearFloat` in
  `layout/div_test.go`), but that path is not on the document/page flow used here.

## Environment

- folio `v0.9.1`
- Go 1.26 (module declares `go 1.25.0`)
- macOS arm64 (not OS-specific)

## Workaround

Use `display:flex` (or CSS grid) for multi-column layout; see the sibling
[`flex-columns`](../flex-columns) case.
