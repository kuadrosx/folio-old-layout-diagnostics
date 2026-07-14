Gap: FOLIO-GAP-11

# folio: background not clipped to `border-radius` on overflow

**Status:** ❌ Bug — a square background band spills below the rounded corner.

## Summary

A fixed-height (34px) flex "rectangle" pill with rounded corners
(`border-radius:10px`) whose children are taller than 34px (a 26px number plus a
16px label). Chrome clips the amber background to the border-radius — the pill
stays rounded and the extra content overflows visibly. folio paints the overflow
background as a **square band** that spills below the rounded bottom, so the
bottom of the fill is a flat, square-cornered rectangle instead of a rounded arc.

## Minimal repro

`sample.html` — `.rectangle { display:flex; align-items:center; height:34px;
padding:10px 13px; border-radius:10px; background:#ffbc00 }` containing a
`font-size:26px` number and a `font-size:16px; line-height:1.58` label (taller
than 34px).

```sh
go run .   # renders sample.html -> output.pdf
```

## Expected vs observed

- **Expected (Chrome):** the amber background is clipped to the rounded pill;
  the bottom edge curves inward at the corners.
- **Observed (folio):** the amber background extends below the pill as a flat
  square band. The parity harness measures the "bottom straight run" — how many
  rows up from the very bottom keep a constant (straight) vertical edge: folio's
  band gives a long run (~12 rows, a flat vertical edge reaching the bottom),
  Chrome's rounded bottom gives ~1 (the arc curves every row).

## Source pointers

folio paints the box's background for its overflow height without re-applying
the `border-radius` clip to the overflowed region, so the spilled background is
an un-rounded rectangle.

## Environment

- folio `v0.10.0-1-g1b17d01` (one commit past the v0.10.0 tag), Go 1.26, macOS arm64.

## Workaround

Use `min-height` + `box-sizing: border-box` so the box grows to its content and
never overflows (the rounded background then covers the whole box).
