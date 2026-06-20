# folio: Acid2 — feature-support probe (not a layout bug)

**Status:** ⚠️ Feature probe — most discrepancies are *expected non-support* for a
static, script-free HTML→PDF renderer, **not** folio bugs.

The canonical **Acid2** test fetched verbatim from
[acid2.acidtests.org](http://acid2.acidtests.org/). Unlike the other cases this is
**not** filed as a bug: Acid2 was designed to exercise interactive and CSS2
features (`:hover`, `position`, generated content, `data:` images, `<object>`
embedding) that a print renderer either cannot have or does not aim to support.
It is included as a **support matrix** — a record of what folio does with each
feature — so the genuine layout bugs (the `float-*`, `css1-*` and `acid1` cases)
aren't muddied by expected non-support.

## How to reproduce

```sh
go run .          # renders cases/acid2/sample.html -> cases/acid2/output.pdf
```

## Input

See [`sample.html`](./sample.html) — the unmodified Acid2 source. It is
self-contained except for two deliberate parts of the test: a
`<object data="http://www.damowmow.com/404/">` (a 404 a conformant agent must
ignore gracefully) and `#top`/`reference.html` hyperlinks.

## What Acid2 exercises vs. how folio handles it

| Feature (Acid2 usage) | Expectation for folio | Observed |
|---|---|---|
| `:hover` (×3) | Interactive — N/A in a static PDF | No hover state; expected |
| `data:` URI `<img>` | Inline image decoding | **Not decoded** — emitted as a literal `[image: data:...]` text placeholder |
| `<object>` data/embedding | Out of scope for print | Not embedded; expected |
| `:before` generated content (×2) | CSS2 generated content | Not reproduced |
| `position: absolute` (×3) | Absolute positioning / overlays | Not reproduced; the face is not assembled |
| `float` (×11) | Float layout | Broken — see the dedicated `float-*` / `acid1` cases |

## Observed in folio

- The render **succeeds** (no error) but produces **2 pages**; the reference
  "smiley face" is not assembled.
- Inline `data:` URI images are not rendered — folio writes the URI as text
  inside an `[image: ...]` placeholder.
- The only part relevant to the layout-bug investigation is Acid2's float usage,
  which is already isolated by [`css1-float-textflow`](../css1-float-textflow),
  [`css1-clear`](../css1-clear), [`float-left-right`](../float-left-right) and
  [`acid1`](../acid1). Everything else here is feature non-support by design.

## Takeaway

Acid2 is **not** a useful float/layout regression for folio — treat it as a
capability checklist. For actual bugs, see the cases listed in the
[README](../../README.md). For multi-column layout, use `display:flex` or CSS grid
([`flex-columns`](../flex-columns)).

## Environment

- folio `v0.9.1`
- Go 1.26 (module declares `go 1.25.0`)
- macOS arm64 (not OS-specific)
