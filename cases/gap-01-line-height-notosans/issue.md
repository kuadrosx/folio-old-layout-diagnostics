Gap: FOLIO-GAP-01 (font regression case)

# folio: `line-height: normal` for Noto Sans

**Status:** regression coverage — verifies the FOLIO-GAP-01 fix generalizes
beyond Poppins to a font with different declared vertical metrics.

## Summary

Same construct as `gap-01-line-height` (two lines of text at
`line-height: normal`), rendered with Noto Sans instead of Poppins, to guard
against a fix that happens to work for one font's metrics but regresses
another's.

## Minimal repro

`sample.html` — two lines of Noto Sans text at `font-size:11px;
line-height:normal` (the font ships in this folder and is referenced via a
case-local `@font-face`).

```sh
go run .   # renders sample.html -> output.pdf
```

## Source

Noto Sans — sourced independently via Homebrew.
