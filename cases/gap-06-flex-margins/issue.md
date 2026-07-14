Gap: FOLIO-GAP-06

# folio: margins on flex items are ignored

**Status:** ❌ Bug — flex-child margins dropped on both axes.

## Summary

`margin-left`/`margin-right` (a main-axis gutter) and `margin-top` (a cross-axis
offset) set on a flex child are ignored by folio. Chrome offsets the child by
both margins. A two-column layout that relies on a side margin for its gutter
collapses together, and a `margin-top` offset is lost.

## Minimal repro

`sample.html` — a `display:flex` row with two 180px columns; the second column
has `margin-left:40px` (gutter) and `margin-top:30px` (cross-axis offset).

```sh
go run .   # renders sample.html -> output.pdf
```

## Expected vs observed

- **Expected (Chrome):** Column B starts 40px to the right of Column A and 30px
  lower.
- **Observed (folio):** Column B butts directly against Column A (no 40px
  gutter) and is not offset down (no 30px top margin).

## Source pointers

folio's flex layout (`layout/flex.go`) positions items from basis/flex sizing
without adding the item's box margins to its main-axis start or cross-axis
offset.

## Environment

- folio `v0.10.0-1-g1b17d01` (one commit past the v0.10.0 tag), Go 1.26, macOS arm64.

## Workaround

Use `justify-content: space-between` (or gap) for the gutter and `align-self`
for cross-axis placement — folio honors both.
