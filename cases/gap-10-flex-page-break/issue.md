Gap: FOLIO-GAP-10

# folio: a flex line taller than a page is clipped and dropped (content loss)

**Status:** ❌ Bug — MOST SEVERE: real content silently lost.

## Summary

folio does not fragment a single flex line across a page boundary. A first/only
flex line taller than the page is laid out whole and everything past the page
bottom is **clipped and silently dropped** — it does not continue on page 2. The
browser fragments the flex line and keeps all the content.

## Minimal repro

`sample.html` — a `display:flex` row with one column containing 19 stacked
60px items (total content taller than one A4 page). The last item is marked and
must survive onto page 2.

```sh
go run .   # renders sample.html -> output.pdf
pdfinfo output.pdf | grep Pages
```

## Expected vs observed

- **Expected (Chrome):** 2 pages; items overflow onto page 2 and the LAST item
  is present.
- **Observed (folio):** **1 page**; the items past the page bottom (including
  the LAST item) are clipped and never rendered — permanent content loss.

## Source pointers

- `layout/flex.go` `planRow` — a line only breaks to the next page when a prior
  line already fit (`fittedLineCount > 0`, ~line 426). A first/only flex line
  taller than the page is placed whole; `overflowFrom` (~line 800) carries only
  *subsequent* lines, so the overflow of the first line is discarded.

## Environment

- folio `v0.10.0-1-g1b17d01` (one commit past the v0.10.0 tag), Go 1.26, macOS arm64.

## Workaround

Keep only page-fitting content inside a flex container. Move content that can
exceed one page height out of the flex row into a full-width block after it,
where normal block-flow pagination applies.
