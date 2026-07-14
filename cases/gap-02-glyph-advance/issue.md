Gap: FOLIO-GAP-02

# folio: Poppins-Bold glyph advances ~9% wider than the browser

**Status:** ❌ Bug (metric) — no CSS fix; blocks pixel parity.

## Summary

For the identical embedded `Poppins-Bold.ttf`, folio's glyph advances run ~9%
wider than the font's `hmtx` table (as the browser shapes it). A 35-char token
measures **~546px in folio vs ~502px in Chrome** (150dpi) at the same font-size,
so it overflows a fixed-width box and wraps where the browser fits it. Different
wrap points reflow lines and drift pagination — pixel parity is unreachable
until folio matches the font's advance widths.

## Minimal repro

`sample.html` — a 320px box containing a long unbreakable Poppins-Bold token
(`white-space: nowrap`). The TTF ships in this folder, referenced via a
case-local `@font-face`.

```sh
go run .   # renders sample.html -> output.pdf
```

## Expected vs observed

- **Expected (Chrome):** advance widths matching the font's `hmtx` table; the
  token's measured width is ~502px.
- **Observed (folio):** ~546px (~9% wider); the token overflows / breaks where
  the browser does not.

Also: folio **ignores `white-space: nowrap` for an over-long unbreakable token**
— it breaks the token anyway instead of overflowing.

## Source pointers

The advance width used during line-breaking comes from folio's font shaping of
the embedded TTF; it does not match the font's `hmtx` advances that browsers
use, so measured text runs consistently wider.

## Environment

- folio `v0.10.0-1-g1b17d01` (one commit past the v0.10.0 tag), Go 1.26, macOS arm64.

## Workaround

Negative `letter-spacing` to shrink the worst-case token back under the
container width (a band-aid for one case, not a general fix).
