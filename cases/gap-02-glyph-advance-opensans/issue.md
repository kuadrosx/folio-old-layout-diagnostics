Gap: FOLIO-GAP-02 (font regression case)

# folio: nowrap/glyph-advance for Open Sans Bold

**Status:** regression coverage — verifies the FOLIO-GAP-02 fix generalizes
beyond Poppins to a font with different glyph advances.

## Summary

Same construct as `gap-02-glyph-advance` (a 320px `white-space:nowrap`
box containing a long unbreakable token), rendered with Open Sans Bold instead
of Poppins Bold, to guard against a fix that happens to work for one
font's advances but regresses another's.

## Minimal repro

`sample.html` — a 320px box containing a long unbreakable Open Sans Bold token
(`white-space: nowrap`). The font ships in this folder, referenced via a
case-local `@font-face`.

```sh
go run .   # renders sample.html -> output.pdf
```

## Source

Open Sans — sourced independently via Homebrew.
