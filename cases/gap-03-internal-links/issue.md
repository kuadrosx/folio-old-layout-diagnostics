Gap: FOLIO-GAP-03

# folio: internal anchor links (`href="#id"`) — RESOLVED in this build

**Status:** ✅ Fixed in folio `v0.10.0-1-g1b17d01` — the internal link now
resolves to a registered named destination.

## Summary

An `<a href="#sec">` pointing at an element with `id="sec"` on a later page now
navigates correctly. folio commit **1b17d01 "html: register element ids as PDF
named destinations"** auto-registers any element carrying an `id` as a named
destination, and the block-level link resolves to a real `/Dest` on the target's
page. The parity harness confirms the folio PDF has a `/Dests` dictionary and no
dangling `/GoTo` string action, matching Chrome.

## Minimal repro

`sample.html` — a block-level `<a href="#sec">` at the top, a tall spacer, then
`<div id="sec">` (which lands on a later page).

```sh
go run .   # renders sample.html -> output.pdf
```

## Expected vs observed

- **Expected (Chrome):** a working internal link resolving to `#sec` on its page.
- **Observed (folio `v0.10.0-1-g1b17d01`):** the decompressed PDF contains a
  `/Dests` dictionary registering `sec` on the target's page, and the link
  annotation resolves to it (a direct `/Dest`, no dangling `/GoTo (sec)`). The
  link works — matches Chrome.

## History (the gap this case guards against)

Before 1b17d01 folio emitted a dangling `/GoTo /D (sec)` string action (or a
bare `/URI (#sec)`) with **zero** registered destinations, so the link jumped
nowhere. This was the PDF-rendering report's "Ver desglose" dead-link symptom
(FOLIO-GAP-03). commit 1b17d01 restores the `layout.Anchor` auto-registration
documented under folio 0.8.0 (#223), which had been absent from the tree.

## Structural check (used by the parity test)

The check decompresses the folio PDF and flags the link **dead** only if it has
a `/GoTo` with no registered destination (0 `/Dests` and 0 `/Names`) or a bare
`/URI (#…)`. Here a `/Dests` dict is present and no dangling `/GoTo` remains, so
the check reads **PASS** (matches Chrome). It will alert (FAIL) if the
registration ever regresses.

## Environment

- folio `v0.10.0-1-g1b17d01` (one commit past the v0.10.0 tag), Go 1.26, macOS arm64.

## Workaround

None needed in this build. (Production, on an older folio without 1b17d01, still
needs the fix — this case verifies the fix and guards against regression.)
