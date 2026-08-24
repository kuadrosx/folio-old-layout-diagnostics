# Internal anchors that flow inline are emitted as `/URI (#id)` and navigate nowhere

**Case:** `cases/gap-15-inline-internal-link` ❌ reproduces
**Control:** `cases/gap-15-block-internal-link-control` ✅ works

## Summary

An `<a href="#some-id">` that appears inside a paragraph, heading, list item
or table cell is written into the PDF as a `/URI` action whose value is the
literal string `#some-id`:

```
<< /Type /Annot /Subtype /Link /Rect [...] /A << /Type /Action /S /URI /URI (#target-section) >> >>
```

A URI action pointing at a bare fragment is not a URI at all — there is no
scheme and no resource — so no viewer resolves it. The link looks present
(it is underlined, it has a clickable rect) but clicking it does nothing.

This is the ordinary case, not an exotic one: `display: inline` is the CSS
default for `<a>`, so *every* internal link in running text is affected. The
construct matters for any document with a table of contents or
cross-references.

The same `href` on an anchor forced to `display: block` is handled correctly
and produces a real destination — see the control case.

## Reproduce

```sh
go run .                       # renders cases/*/sample.html -> output.pdf
```

Then inspect the two outputs:

```sh
strings cases/gap-15-inline-internal-link/output.pdf | grep -E '/URI|/Dest'
strings cases/gap-15-block-internal-link-control/output.pdf | grep -E '/URI|/Dest'
```

## Expected vs observed

`cases/gap-15-inline-internal-link/sample.html` places the same
`<a href="#target-section">` in four inline hosts — a heading, a paragraph,
a list item and a table cell — with the target on a later page.

| | expected | observed (v0.10.0, `next` @ 438802b) |
|---|---|---|
| `/Link` annotations | 4 | **3** |
| `/URI (#target-section)` actions | 0 | **3** |
| resolvable destinations | 4 | **0** |

Two distinct failures show up in that table:

1. **Wrong action kind.** The three anchors that do produce an annotation
   emit `/URI (#target-section)` instead of a destination.
2. **Missing annotation.** The anchor in the table cell produces no `/Link`
   annotation at all.

The control renders 1 annotation with `/Dest [<pageref> /XYZ null <y> null]`
and zero `/URI` actions, which is the correct shape.

## Source pointers

- `html/converter_dispatch.go` — `isInlineFlowChild` treats `<a>` as inline
  flow unless CSS overrides `display`, so an `<a>` reaches `convertLink`
  essentially only when it is `display: block`.
- `html/converter_link.go:convertLink` — the correct path. It checks
  `strings.HasPrefix(href, "#")` and builds `layout.NewInternalLink`.
- `html/converter_paragraph.go` (two sites, plus `collectListItemRuns`) —
  the inline path. Both assign the raw href straight to `TextRun.LinkURI`
  with no `#` check:

  ```go
  if child.DataAtom == atom.A {
      href := getAttr(child, "href")
      if href != "" {
          for i := range childRuns { childRuns[i].LinkURI = href }
      }
  }
  ```

  `LinkURI` is written out as a `/URI` action, and there is no field on
  `TextRun` for an internal destination — which is likely why the check was
  omitted rather than mislaid.
- `layout/paragraph_wrap.go:linkSpans` — builds `LinkArea` values keyed on
  `Word.LinkURI` only, so a destination has nowhere to travel.
- `layout/table.go:drawCellElementDirect` — table cells are drawn through the
  direct-draw path, which never copies `block.Links` onto the page. That is a
  separate root cause and explains the missing fourth annotation; a link in a
  table cell is lost regardless of whether it is internal or external.

`layout.LinkArea` already carries a `DestName` field and
`document/document.go` already resolves a named destination to a direct page
reference (the `ann.dest != ""` branch), so the writer end needs no change —
only a companion field on `TextRun`/`Word` to carry the destination from the
converter to `linkSpans`.

## Suggested fix

1. Add `LinkDest` alongside `LinkURI` on `layout.TextRun` and `layout.Word`,
   and carry it through `paragraph_measure.go`, `paragraph_split.go` and
   `linkSpans` into `LinkArea.DestName`.
2. Route the href in the inline `<a>` handlers: a leading `#` becomes
   `LinkDest`, anything else stays `LinkURI`.
3. Record `block.Links` in the table cell direct-draw path so cell links
   reach the page at all.

## Environment

- folio `v0.10.0`, branch `next` @ `438802b`
- Go 1.25, darwin/arm64
- A4 page, zero margins (`main.go`)

## Workaround

Force `display: block` on internal anchors, which routes them down the
working path. That is only viable where an inline link is not required —
an anchor in the middle of a sentence cannot be made block-level without
breaking the line.
