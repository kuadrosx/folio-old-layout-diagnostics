# folio: block establishing a BFC (`overflow:hidden`) does not contain its floats

**Status:** ❌ Bug

## Summary

Wrapping floated columns in a container with `overflow:hidden` is the standard CSS
"contain the floats" idiom — the container establishes a block formatting context
and grows to enclose its floated children. In folio this does not happen: the
container collapses to zero height and the **floated children disappear entirely**
from the output.

This case is the diagnostic counterpart to [`float-left-right`](../float-left-right):
it shows the problem is not specific to the body element — even an explicit BFC
container fails to contain floats.

## How to reproduce

From the repository root:

```sh
go run .          # renders cases/<name>/sample.html -> cases/<name>/output.pdf
```

Then open `output.pdf` in this folder.

## Input

See [`sample.html`](./sample.html): an `overflow:hidden` `.container` wrapping a
`float:left;50%` and a `float:right;50%` column, followed by an `.after` block.

## Expected (per CSS)

`.container` establishes a BFC and grows to contain both floats; the two columns
render side by side inside it; `Content after` renders below the container.

## Observed in folio

The container collapses to zero height and the floated children are not rendered.
Only `Content after` is visible.

## Likely cause (source pointers)

- Block containers do not appear to grow to enclose their floated children, so a
  BFC container collapses (related to floats reporting `Consumed: 0`,
  `layout/float.go:97`).
- `html/converter.go:1538-1549` — each child of a floated element is wrapped in its
  own `layout.NewFloat(side, e)`, so a floated `<div>` with `<h2><p>` becomes
  multiple independent floats rather than one block.

## Environment

- folio `v0.9.1`, also reproduces on `main` @ `72b6a6a`
- Go 1.26 (module declares `go 1.25.0`)
- macOS arm64 (not OS-specific)

## Workaround

Use `display:flex` (or CSS grid) — see [`flex-columns`](../flex-columns).
