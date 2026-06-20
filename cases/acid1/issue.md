# folio: Acid1 (CSS1 box/float/clear torture test) — floats collapse to one column

**Status:** ❌ Bug

The canonical **Acid1** test (the Web Standards Project "box acid test", *display/
box/float/clear test*, last modified 1 Dec 1998), fetched verbatim from
[acid1.acidtests.org](http://acid1.acidtests.org/). It is valid HTML 4.0 + CSS1,
uses **no** JavaScript and **no** external rendering assets (the only `href`s are
hyperlinks to the reference image and parent page), so it drops into this repo
self-contained.

Unlike the other cases, this is **not** a minimal isolation — it is a
comprehensive conformance benchmark that layers floats, percentage widths,
borders, padding, margins and `clear` on top of one another. The minimal repros
that isolate the individual failures are [`css1-float-textflow`](../css1-float-textflow),
[`css1-clear`](../css1-clear), [`float-left-right`](../float-left-right) and
[`float-in-bfc`](../float-in-bfc).

## How to reproduce

```sh
go run .          # renders cases/acid1/sample.html -> cases/acid1/output.pdf
```

Then open `output.pdf` in this folder.

## Input

See [`sample.html`](./sample.html) — the unmodified Acid1 source. Floated `dt`
(`float:left; width:10.638%`), `dd` (`float:right; width:34em`), `li`
(`float:left`) and `blockquote` (`float:left`) boxes are meant to interlock into
a fixed picture.

## Expected (per CSS)

All elements above the explanatory paragraph render **pixel-identical** to the
W3C reference rendering [`sec5526c.gif`](http://acid1.acidtests.org/sec5526c.gif)
(except font rasterization and form widgets): a precise grid of floated boxes,
with the `float:right` block against the right edge and percentage widths applied.

## Observed in folio

Measured from the rendered PDF (A4 = 595.28pt wide):

- Every text line lands at **x ≈ 72–147pt** — i.e. everything piles into a single
  left column. **No** content reaches the right half of the page, so the
  `float:right` `<dd>` is not at the right edge and the floated boxes never sit
  side by side.
- The percentage widths (`10.638%`, `41.17%`) collapse rather than resolving
  against the parent width, so the boxes do not size into the intended grid.
- The result bears no resemblance to the reference rendering; the floated boxes
  stack vertically in document order instead of interlocking.

## Likely cause (source pointers)

Same shared root causes as the other float cases:

- `html/converter.go:1538-1549` — each child of a floated element becomes its own
  `layout.NewFloat`, fragmenting multi-child floated boxes.
- `layout/float.go:67-74` — a float's width is derived from content, so explicit
  and percentage CSS `width` is ignored.
- `layout/float.go:97` — a float reports `Consumed: 0`, so it adds no vertical
  space and `clear` has no float height to clear against.
- `html/converter.go:658` — `ConvertFull` returns `<body>` children as a flat
  element list with no float context established.

## Environment

- folio `v0.9.1`
- Go 1.26 (module declares `go 1.25.0`)
- macOS arm64 (not OS-specific)

## Workaround

None for this test as written — it is a CSS1 conformance benchmark. For real
multi-column layout in folio, use `display:flex` or CSS grid; see
[`flex-columns`](../flex-columns).
