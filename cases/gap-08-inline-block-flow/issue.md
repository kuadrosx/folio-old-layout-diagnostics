Gap: FOLIO-GAP-08

# folio: consecutive `display: inline-block` badges stack instead of flowing

**Status:** ❌ Bug — inline-block badges laid out one per line.

## Summary

Several `display: inline-block` "badges" back-to-back (with whitespace/newlines
between them) inside a plain block container ~48% of the page width should flow
horizontally and wrap to new lines as needed. folio instead lays each badge out
as its own block — **one badge per line** — stacking them vertically. Chrome
flows them, wrapping to a few rows.

## Minimal repro

`sample.html` — six `<span class="badge">` (`display:inline-block; padding:
0.35em 0.65em; white-space:nowrap; margin-bottom:0.2em; background:#4800ff;
color:#fff; border-radius:0.25rem`) inside a `width:48%` plain block.

```sh
go run .   # renders sample.html -> output.pdf
```

## Expected vs observed

- **Expected (Chrome):** badges flow horizontally and wrap — the six badges
  occupy ~3 rows (the parity harness measures 3 badge row-bands).
- **Observed (folio):** each badge on its own line — 6 badge row-bands, one per
  badge, stacked vertically.

## Source pointers

folio's converter treats `display: inline-block` elements as block-level in the
normal flow (each gets its own line box) rather than placing them inline within
a shared line box.

## Environment

- folio `v0.10.0-1-g1b17d01` (one commit past the v0.10.0 tag), Go 1.26, macOS arm64.

## Workaround

Wrap the badges in a `display:flex; flex-wrap:wrap` container.
