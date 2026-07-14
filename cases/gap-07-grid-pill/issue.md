Gap: FOLIO-GAP-07

# folio: `display: grid` box stretches to fill and ignores `border-radius`

**Status:** ❌ Bug — grid box does not shrink-to-content and has square corners.

## Summary

A `display: grid` pill that should shrink to its content and round its corners
instead stretches to fill its container width and renders with square corners
(`border-radius` ignored). Chrome draws a compact rounded pill hugging the text.

## Minimal repro

`sample.html` — a `.pill { display:grid; border-radius:16px }` inside a 400px
container.

```sh
go run .   # renders sample.html -> output.pdf
```

## Expected vs observed

- **Expected (Chrome):** a rounded pill (a `display:grid` item stretches to the
  grid track by default, so it spans the container width in both engines — the
  distinguishing feature is the **rounded** corners).
- **Observed (folio):** a **square-cornered** bar — `border-radius` is not
  applied to the grid box. (The parity harness confirms folio's top-left corner
  is square while Chrome's is rounded.)

Related shrink-to-content limits (same family):
- folio can't shrink a flex box to content either — `width: fit-content` /
  `max-content` collapse to min-content (breaking the word).
- folio stacks `inline-flex` / inline-block children instead of flowing them.

`border-radius` *does* work on non-grid flex / inline-block boxes; the grid box
is the one that ignores it.

## Source pointers

folio's grid layout sizes the grid container to the available width rather than
to its content's intrinsic size, and the grid paint path does not apply the
box's `border-radius` to its background.

## Environment

- folio `v0.10.0-1-g1b17d01` (one commit past the v0.10.0 tag), Go 1.26, macOS arm64.

## Workaround

Use a fixed-width flex row (border-radius works there) instead of grid.
