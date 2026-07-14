Gap: FOLIO-GAP-01

# folio: `line-height: normal` is too tight for Poppins

**Status:** ❌ Bug (metric) — visible spacing deviation vs the browser.

## Summary

For the embedded `Poppins-Regular.ttf`, folio resolves `line-height: normal`
from a tighter vertical metric than browsers do. Two consecutive 11px lines
render **~23px apart in folio vs ~29px in Chrome** (measured at 150dpi, A4).
Over a text-heavy column folio ends up visibly more compact than every browser,
which also drifts pagination.

## Minimal repro

`sample.html` — two lines of Poppins text at `font-size:11px; line-height:normal`
(the TTF ships in this folder and is referenced via a case-local `@font-face`).

```sh
go run .   # renders sample.html -> output.pdf
```

## Expected vs observed

- **Expected (Chrome):** `normal` leading derived from the font's vertical
  metrics — Poppins declares a large line-gap, so the baselines sit ~29px apart.
- **Observed (folio):** baselines ~23px apart (~20% tighter).

Note the resolution is also *inconsistent across contexts*: pinning an explicit
`line-height` on flowing text fixes the flow case, but folio's default `normal`
for `<td>` rows already matches the browser — so a body-wide pin over-inflates
table rows. folio resolves `normal` differently for flow text vs table cells.

## Source pointers

folio's `normal` leading for a `@font-face` TTF is computed from the font's
`hhea`/`OS/2` metrics; the derived leading is narrower than the browser's
`max(ascent+descent, lineGap+...)` convention for Poppins.

## Environment

- folio `v0.10.0-1-g1b17d01` (one commit past the v0.10.0 tag), Go 1.26, macOS arm64.

## Workaround

Pin an explicit `line-height` (~1.58) scoped to non-table flowing text only.
