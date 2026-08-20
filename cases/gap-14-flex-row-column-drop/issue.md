# folio: a flex row drops a column that cannot start on the current page (content loss)

**Status:** ❌ Bug — content silently lost. Fixed on a local folio checkout; see
"Fix" below. The case is kept as a regression guard.

## Summary

A `display: flex` row used as a *row wrapper* — a short fixed-size marker in a
gutter column plus a second column holding a tall stack of tables — loses the
entire second column when the row straddles a page boundary and the space left
on the page is too small for the column's first table row.

The output is a well-formed PDF. It paginates, every page looks complete,
nothing appears cut off: the column's content is simply absent. The page count
moves the wrong way too — content that was dropped never asks for pages of its
own, so the broken render is **shorter** than the correct one. On a large
multi-page document, turning one block wrapper into such a flex row dropped
about 10% of the extractable text while reducing the page count from 44 to 39.
Count the text, not the pages.

## Minimal repro

`sample.html` — a spacer that leaves a sliver of page 1, then a flex row
(`align-items: flex-start`) whose gutter column holds a 6px marker and whose
second column holds three 4-row tables (`Row01` … `Row12`). The sliver is big
enough for the marker but not for the first table row.

```sh
go run .                                                  # renders sample.html -> output.pdf
pdfinfo cases/gap-14-flex-row-column-drop/output.pdf | grep Pages
pdftotext cases/gap-14-flex-row-column-drop/output.pdf - | grep -c Row   # expect 12
```

`cases/gap-14-block-row-control/` is the passing control: the same wrapper and
the same column with `display: flex` removed. It fragments correctly, which
places the boundary of the bug in the flex row path rather than in the tables or
the column.

## Expected vs observed

|                          | pages | `RowNN` tokens in the text |
| ------------------------ | ----- | -------------------------- |
| Expected (Chrome)        | 2     | 12                         |
| Observed (folio, before) | **1** | **0**                      |
| Control, plain block     | 2     | 12                         |

Chrome starts the row on page 1 and continues the column on page 2. folio laid
the row out with the marker alone and dropped the column.

## Cause

`layout/flex.go` `planRow` did not handle a flex item whose plan came back
`LayoutNothing`. That status means the item placed nothing at all, so it reports
neither blocks nor overflow and its content lives on only in the element itself.
`planRow` read the (empty) blocks, laid the line out with the other columns and
reported `LayoutFull`, so the item's content was never drawn and never carried to
the next page.

The status is easy to reach at a page boundary: a column whose first table has
to move its header and first body row together reports it (`layout/table.go`
orphan guard), as does a nested auto-height box that could not place its first
child (`layout/div.go` zero-progress guard).

With `align-items: flex-start` / `center` / `flex-end` the whole column
disappeared. Under the default `align-items: stretch` the cross-axis re-layout
pass hid the fault whenever the marker happened to be *taller* than the column's
first unbreakable unit, which is why the bug looks intermittent.

## Fix

Defer a line with such a column instead of laying it out: to the next page when
earlier lines already fit, or — when nothing has been placed yet — report
`LayoutNothing` so the renderer relocates the container to a fresh page where
the column can make progress. At the top of a page the renderer force-places a
`LayoutNothing` element with an effectively unbounded height, so pagination
cannot loop. Scoped to auto-height flex containers, matching the existing
carve-out that lets a definite-height flex contain/clip its content.

## Environment

- folio `v0.10.0-1-g1b17d01` (one commit past the v0.10.0 tag), Go 1.26, macOS arm64.
- Chrome (headless, `--print-to-pdf`) as the reference; poppler `pdftotext` /
  `pdfinfo` for the measurements above.

## Workaround (before the fix)

Do not use a flex row as a wrapper around content that can exceed the space left
on a page. Use a block wrapper for the row and reserve flex for content that
fits within one page.
