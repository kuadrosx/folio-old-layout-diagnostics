# folio: inline text does not flow past a `float` (no horizontal space reserved)

**Status:** ❌ Bug

Adapted from the W3C CSS1 Test Suite, [section 5.5.25 "float"](https://www.w3.org/Style/CSS/Test/CSS1/current/test5525.htm).
The original floats an `<img>`; here it is a self-contained CSS block so the case
needs no external assets.

## Summary

A `float:left` (and separately `float:right`) block followed by a paragraph of
normal-flow text should cause that text to **wrap beside** the float — the float
reserves horizontal space in the line box and the text is indented past it. In
folio the float reserves **no** horizontal space: the paragraph starts at the
normal left text margin, exactly as if the float were not there, so the text
overlaps the floated rectangle. `float:right` likewise neither moves the box to
the right edge nor shortens the line.

## How to reproduce

```sh
go run .          # renders cases/css1-float-textflow/sample.html -> output.pdf
```

Then open `output.pdf` in this folder.

## Input

See [`sample.html`](./sample.html): a `float:left` orange block + paragraph, then
(after a `clear:both` block and `<hr>`) a `float:right` orange block + paragraph.

## Expected (per CSS)

- The first paragraph's text begins to the **right** of the left-floated block
  (indented by the float's width), flowing past it.
- The second paragraph's text flows to the **left** of the right-floated block,
  which sits at the container's right edge; the line is shortened accordingly.

## Observed in folio

Measured from the rendered PDF (A4, text content origin x ≈ 125pt):

- Left-float paragraph: first line starts at **x ≈ 125pt** — the normal left
  margin, with no indent for the float. Text overlaps where the floated block is.
- Right-float paragraph: text also starts at **x ≈ 125pt** and wraps at the full
  content width, so `float:right` reserved no space on the right either.

(The orange rectangles are empty blocks and carry no text, so their exact position
isn't recoverable from `pdftotext`; open the PDF to see the overlap visually.)

## Likely cause (source pointers)

- `layout/float.go:97` — a float reports `Consumed: 0`; it adds no space to the
  flow. The same gap that drops *block* content onto floats also means *inline*
  line boxes are not shortened around the float.
- `html/converter.go:658` — `ConvertFull` returns `<body>` children as a flat
  element list with no float context, so the paragraph after the float never sees
  it and lays out at the full content width from the left margin.

## Environment

- folio `v0.9.1`
- Go 1.26 (module declares `go 1.25.0`)
- macOS arm64 (not OS-specific)

## Workaround

Use `display:flex` (or CSS grid) for side-by-side layout; see the sibling
[`flex-columns`](../flex-columns) case.
