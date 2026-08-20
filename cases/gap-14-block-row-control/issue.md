# folio: block row wrapper fragmenting a tall column (control — works)

**Status:** ✅ Control / passing

## Purpose

Passing control for `gap-14-flex-row-column-drop`. It carries the same row
wrapper and the same tall column of stacked tables, with `display: flex` removed
from the wrapper, and is positioned at the same page boundary: the spacer leaves
a sliver of page 1 that is too small for the column's first table row.

The block wrapper fragments correctly — the column continues onto page 2 with
every `Row01` … `Row12` intact — while the flex version dropped the whole column
and rendered a single, complete-looking page. That contrast is what places the
boundary of the bug in the flex row path rather than in the tables, the column,
or the page-boundary arithmetic.

## How to reproduce

```sh
go run .          # renders cases/<name>/sample.html -> cases/<name>/output.pdf
pdfinfo   cases/gap-14-block-row-control/output.pdf | grep Pages     # 2
pdftotext cases/gap-14-block-row-control/output.pdf - | grep -c Row  # 12
```

## Observed in folio

Correct: 2 pages, all 12 rows present — before and after the flex fix.

## Environment

- folio `v0.10.0-1-g1b17d01` (one commit past the v0.10.0 tag), Go 1.26, macOS arm64.
