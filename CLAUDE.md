# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## What this repo is

This is **not** an application — it is a collection of minimal, self-contained bug
reproductions for the third-party HTML→PDF library `github.com/carlos7ags/folio`
(see `go.mod`). The goal of each case is to be independently reportable upstream.
The bugs live in `folio`, not here; this repo's job is to demonstrate them clearly.

## Commands

```sh
go run .          # renders every cases/<name>/sample.html -> cases/<name>/output.pdf
go build .        # compile only
```

There are no tests, linters, or CI configured in this repo. The committed
`cases/<name>/output.pdf` files are intentionally checked in as rendered evidence
of each bug — regenerate them with `go run .` after changing any `sample.html`.

## Structure

`main.go` is a ~70-line driver: it iterates every subdirectory of `cases/`, reads
`sample.html`, renders it via folio's `document.AddHTMLWithContext` on an A4 page,
and writes `output.pdf` next to it. A case folder missing `sample.html` is skipped,
not an error.

Each `cases/<name>/` is a complete reproduction triad:
- `sample.html` — a standalone HTML document exercising one CSS construct
- `issue.md` — the upstream bug report (summary, repro, expected vs observed,
  source pointers into folio, environment, workaround)
- `output.pdf` — folio's actual rendering (the evidence)

## The findings (context for any new case)

The README and `issue.md` files document a coherent investigation, not isolated
bugs. The throughline: in folio, a **single** float works, but **two** floats
(`left`+`right`), `clear:both`, float containment via `overflow:hidden` (BFC), and
`display:table-cell` columns all fail. `display:flex` and CSS grid work and are the
recommended workaround. Cases are split into failing repros and passing controls
(e.g. `minimal-float`, `plain-blocks`) so the boundary of the bug is provable.

Shared root-cause pointers (into the folio source, referenced across issue files):
- `html/converter.go:1538-1549` — each child of a floated element becomes its own
  `layout.NewFloat`, splitting one floated `<div>` into several floats.
- `layout/float.go:67-74` — a float's width is derived from content, so explicit
  CSS `width` is ignored.
- `layout/float.go:97` — a float reports `Consumed: 0` (no normal-flow height).
- `html/converter.go:658` — `ConvertFull` returns `<body>` children as a flat list
  with no float context established, so following content doesn't clear floats.

## Conventions when adding or editing cases

- Keep `sample.html` minimal and standalone — one construct per case, inline CSS,
  no external assets.
- Keep `issue.md` self-contained and upstream-ready: it should make sense to a
  folio maintainer who has never seen this repo.
- Mark each case clearly as a failing bug (❌) or a passing control (✅), and keep
  the README's case table and summary in sync with the cases on disk.
- After any HTML change, run `go run .` to regenerate the committed `output.pdf`.
