Gap: FOLIO-GAP-05

# folio: `position: absolute` ignored inside a content-collapsed relative box

**Status:** ❌ Bug — the absolutely-positioned text is not placed at its
position; the relative banner renders empty.

## Summary

A relative "banner" (the report's "banner" pill) whose visible height
comes from padding — so its content box collapses to zero — contains a single
absolutely-positioned `<p style="position:absolute; top:12px; width:80%">`.
Chrome honors `position:absolute` and paints the text near the banner top, so
the banner shows its text. folio ignores the absolute positioning and lays the
`<p>` out in **normal flow below the banner**, leaving the banner's
absolute-position region **empty**. The text ends up in the wrong place (or, in
the report's build, absent), so the banner looks empty.

## Minimal repro

`sample.html` — `<div style="position:relative; padding-top:58px;
background:#d4f1e4">` containing `<p style="position:absolute; top:12px;
width:80%">…</p>`.

```sh
go run .   # renders sample.html -> output.pdf
```

## Expected vs observed

- **Expected (Chrome):** the text sits near the top of the mint banner
  (text block top-y ≈ 22 at 150dpi, i.e. `top:12px`).
- **Observed (folio):** the banner is empty; the text is displaced into normal
  flow below it (text block top-y ≈ 112 — well past the banner). The parity
  harness measures the text's top-y in both engines and flags the ~90px
  displacement.

> Note on the trigger: in this build the misbehavior needs the relative
> parent's *content* height to be zero (here via `padding-top`, matching the
> report's content-sized banner). With an explicit `height:58px` folio instead
> paints the overlay in place. The downstream gap doc recorded the report's
> symptom as the text being **dropped entirely** (empty box); here folio keeps
> the text but puts it in the wrong place — either way the banner is empty and
> the construct differs from the browser.

## Source pointers

folio resolves absolutely-positioned elements via `ConvertResult.Absolutes` /
the pending-overlay path (`html/converter_dispatch.go`, `document/html.go`
`AddConvertResult`). When the positioned ancestor's content box collapses, the
overlay is not painted at the absolute position; the element falls back into
normal flow after the banner.

## Environment

- folio `v0.10.0-1-g1b17d01` (one commit past the v0.10.0 tag), Go 1.26, macOS arm64.

## Workaround

Replace absolute positioning with flexbox (folio honors flex alignment), or give
the relative parent real in-flow height so the overlay is painted in place.
