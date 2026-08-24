# Control: a block-level internal anchor emits a working destination

**Case:** `cases/gap-15-block-internal-link-control` ✅ passing control
**Paired with:** `cases/gap-15-inline-internal-link` ❌ reproduces

## Why this case exists

This is the control that makes the boundary of
`cases/gap-15-inline-internal-link` provable. It renders the same
`<a href="#target-section">` pointing at the same target on the same page
geometry, differing in exactly one declaration:

```css
a { display: block; }
```

That single declaration decides which converter path the anchor takes. As a
block-level element it is dispatched on its own and reaches
`html/converter_link.go:convertLink`, which checks for a leading `#` and
builds an internal link. Left at its CSS default (`display: inline`) the same
anchor is folded into the surrounding text and emitted as a `/URI` action
carrying the bare fragment — a link that resolves to nothing.

## Expected vs observed

| | expected | observed (v0.10.0, `next` @ 438802b) |
|---|---|---|
| `/Link` annotations | 1 | 1 ✅ |
| `/URI (#target-section)` actions | 0 | 0 ✅ |
| resolvable destinations | 1 | 1 ✅ |

The emitted annotation is:

```
<< /Type /Annot /Subtype /Link /Rect [...] /Dest [<pageref> /XYZ null <y> null] >>
```

which is the correct shape: a direct destination array pointing at the page
object that holds the target, with `/XYZ` so the viewer keeps the current
zoom.

## Reproduce

```sh
go run .
strings cases/gap-15-block-internal-link-control/output.pdf | grep -E '/URI|/Dest'
```

## Environment

- folio `v0.10.0`, branch `next` @ `438802b`
- Go 1.25, darwin/arm64
- A4 page, zero margins (`main.go`)

## Note

`cases/gap-03-internal-links` covers the same block-level construct from the
destination-registration angle (does the target `id` become a named
destination at all). This case is narrower: it fixes the anchor's *action
kind* as the reference point for the inline case next to it.
