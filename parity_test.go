package main

// Chrome parity harness.
//
// For every gap case this test renders the SAME normalized HTML with two
// engines — folio (the library under test) and headless Google Chrome (the
// reference) — on an identical A4, zero-margin page, then decides per gap
// whether folio's rendering MATCHES Chrome's.
//
//	folio-vs-chrome = FAIL  -> folio differs from the browser (gap reproduces)
//	folio-vs-chrome = PASS  -> folio matches the browser (gap absent / fixed)
//
// Each case carries wantReproduce, the state observed when the case was
// authored against this local folio checkout. While reality matches
// wantReproduce the Go test is GREEN. The moment a case's verdict flips —
// a reproducing gap starts matching Chrome (folio fixed it), or a matching
// construct starts differing (folio regressed) — the subtest FAILS with an
// explicit alert. That is the signal to revisit the corresponding CSS
// workaround in downstream.
//
// Why geometric checks instead of a whole-page pixel diff:
// folio and Chrome rasterize fonts differently, so a per-pixel diff of any
// text-bearing region carries irreducible antialiasing noise that is the same
// order of magnitude as the layout gaps we want to detect. Each case therefore
// uses a targeted GEOMETRIC measurement taken from BOTH renders and compared
// relatively — the bounding box of a flat fill colour, whether a pill corner
// is square or rounded, how many rows a set of pills occupies, the vertical
// extent of a block of text. Flat fills and gross geometry are immune to glyph
// antialiasing, so the comparison reflects the construct, not the font raster.
// GAP-03 and GAP-10 use structural PDF checks (link action kind, page count).
//
// The Chrome / poppler / ImageMagick-free design means the only external
// dependencies are Chrome (rendering the reference) and pdftoppm+pdfinfo
// (poppler: rasterize page 1, count pages). All are gated: if any is missing
// the test SKIPS rather than fails, so `go test ./...` degrades gracefully.

import (
	"bytes"
	"compress/zlib"
	"context"
	"fmt"
	"image"
	"image/color"
	_ "image/png"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"testing"

	"github.com/carlos7ags/folio/document"
	foliohtml "github.com/carlos7ags/folio/html"
)

// checkKind selects the per-case verdict strategy.
type checkKind int

const (
	kindGeom      checkKind = iota // decode both page-1 rasters, run geomFn
	kindAnchor                     // GAP-03: folio must emit /GoTo, not /URI (#id)
	kindPageBreak                  // GAP-10: folio page count must be >= Chrome's
)

// geomResult is what a geometric check reports.
type geomResult struct {
	match  bool
	detail string
}

type parityCase struct {
	gap  string
	kind checkKind
	// wantReproduce is the authored-time observation: true if folio differs
	// from Chrome (the gap reproduces) in this local folio checkout.
	wantReproduce bool
	// geomFn measures folio vs chrome for kindGeom cases.
	geomFn func(folio, chrome image.Image) geomResult
}

// caseTable maps each gap case dir to its gap id and verdict strategy. The
// pre-existing float-family cases (FOLIO-GAP-04) are documented by their own
// issue.md files and inspected manually; they are intentionally NOT graded
// here (float geometry is covered by those cases, per the repo README).
var caseTable = map[string]parityCase{
	"gap-01-line-height":            {"FOLIO-GAP-01", kindGeom, false, checkLineHeight},
	"gap-02-glyph-advance":          {"FOLIO-GAP-02", kindGeom, false, checkGlyphAdvance},
	"gap-03-internal-links":         {"FOLIO-GAP-03", kindAnchor, false, nil},
	"gap-05-position-absolute":      {"FOLIO-GAP-05", kindGeom, true, checkAbsolute},
	"gap-06-flex-margins":           {"FOLIO-GAP-06", kindGeom, true, checkFlexMargins},
	"gap-07-grid-pill":              {"FOLIO-GAP-07", kindGeom, true, checkGridPill},
	"gap-08-inline-block-flow":      {"FOLIO-GAP-08", kindGeom, true, checkInlineBlock},
	"gap-09-thead-background":       {"FOLIO-GAP-09", kindGeom, true, checkTheadBg},
	"gap-10-flex-page-break":        {"FOLIO-GAP-10", kindPageBreak, true, nil},
	"gap-11-border-radius-overflow": {"FOLIO-GAP-11", kindGeom, true, checkRadiusOverflow},

	// Multi-font regression coverage for FOLIO-GAP-01/02: the same two
	// constructs re-rendered with fonts other than Poppins, so a future
	// change to line-height/glyph-advance handling that happens to work
	// for one font's metrics but regresses another's is caught here.
	"gap-01-line-height-nimbussans":   {"FOLIO-GAP-01", kindGeom, false, checkLineHeight},
	"gap-01-line-height-inter":        {"FOLIO-GAP-01", kindGeom, false, checkLineHeight},
	"gap-01-line-height-notosans":     {"FOLIO-GAP-01", kindGeom, false, checkLineHeight},
	"gap-01-line-height-opensans":     {"FOLIO-GAP-01", kindGeom, false, checkLineHeight},
	"gap-01-line-height-roboto":       {"FOLIO-GAP-01", kindGeom, false, checkLineHeight},
	"gap-02-glyph-advance-nimbussans": {"FOLIO-GAP-02", kindGeom, false, checkGlyphAdvance},
	"gap-02-glyph-advance-inter":      {"FOLIO-GAP-02", kindGeom, false, checkGlyphAdvance},
	"gap-02-glyph-advance-notosans":   {"FOLIO-GAP-02", kindGeom, false, checkGlyphAdvance},
	"gap-02-glyph-advance-opensans":   {"FOLIO-GAP-02", kindGeom, false, checkGlyphAdvance},
	"gap-02-glyph-advance-roboto":     {"FOLIO-GAP-02", kindGeom, false, checkGlyphAdvance},
}

// ---------------------------------------------------------------------------
// Colours used by the gap samples (RGB). Kept in sync with cases/*/sample.html.
// ---------------------------------------------------------------------------

var (
	colBlueA  = rgb{46, 134, 222}  // #2e86de gap-06 column A
	colOrange = rgb{230, 126, 34}  // #e67e22 gap-06 column B
	colPill   = rgb{74, 144, 217}  // #4a90d9 gap-07 grid pill
	colBadge  = rgb{72, 0, 255}    // #4800ff gap-08 badges
	colGrey   = rgb{204, 204, 204} // #ccc   gap-09 thead bar
	colAmber  = rgb{255, 188, 0}   // #ffbc00 gap-11 rectangle pill
)

type rgb struct{ r, g, b int }

func absi(a int) int {
	if a < 0 {
		return -a
	}
	return a
}

func isColor(c color.Color, want rgb, tol int) bool {
	cr, cg, cb, _ := c.RGBA()
	return absi(int(cr>>8)-want.r) <= tol &&
		absi(int(cg>>8)-want.g) <= tol &&
		absi(int(cb>>8)-want.b) <= tol
}

func isDark(c color.Color) bool {
	cr, cg, cb, _ := c.RGBA()
	return int(cr>>8) < 110 && int(cg>>8) < 110 && int(cb>>8) < 110
}

// box is a pixel bounding box; ok=false means no matching pixels were found.
type box struct {
	x0, y0, x1, y1 int
	count          int
	ok             bool
}

func (b box) w() int { return b.x1 - b.x0 }
func (b box) h() int { return b.y1 - b.y0 }

// maskBBox returns the bounding box of every pixel matching pred.
func maskBBox(img image.Image, pred func(color.Color) bool) box {
	r := img.Bounds()
	b := box{x0: 1 << 30, y0: 1 << 30, x1: -1, y1: -1}
	for y := r.Min.Y; y < r.Max.Y; y++ {
		for x := r.Min.X; x < r.Max.X; x++ {
			if pred(img.At(x, y)) {
				b.count++
				if x < b.x0 {
					b.x0 = x
				}
				if x > b.x1 {
					b.x1 = x
				}
				if y < b.y0 {
					b.y0 = y
				}
				if y > b.y1 {
					b.y1 = y
				}
			}
		}
	}
	if b.x1 >= 0 {
		b.x1++
		b.y1++
		b.ok = true
	}
	return b
}

func colorBBox(img image.Image, want rgb, tol int) box {
	return maskBBox(img, func(c color.Color) bool { return isColor(c, want, tol) })
}

// cornerFilled reports whether the pixel just inside a corner of b is the
// fill colour — true for a square corner, false for a rounded one.
func cornerFilled(img image.Image, b box, want rgb, tol, inset int, right, bottom bool) bool {
	if !b.ok {
		return false
	}
	x := b.x0 + inset
	if right {
		x = b.x1 - 1 - inset
	}
	y := b.y0 + inset
	if bottom {
		y = b.y1 - 1 - inset
	}
	if x < 0 || y < 0 {
		return false
	}
	return isColor(img.At(x, y), want, tol)
}

// countRowBands counts contiguous groups of rows that each contain at least
// minRun pixels of the fill colour — i.e. how many separate lines the coloured
// boxes occupy.
func countRowBands(img image.Image, want rgb, tol, minRun int) int {
	r := img.Bounds()
	bands, inBand := 0, false
	for y := r.Min.Y; y < r.Max.Y; y++ {
		n := 0
		for x := r.Min.X; x < r.Max.X; x++ {
			if isColor(img.At(x, y), want, tol) {
				n++
			}
		}
		if n >= minRun {
			if !inBand {
				bands++
				inBand = true
			}
		} else {
			inBand = false
		}
	}
	return bands
}

// ---------------------------------------------------------------------------
// Per-case geometric checks. Each measures folio and Chrome the same way and
// decides "match" from their agreement, so the check needs no hard-coded pixel
// values and survives rendering-scale changes.
// ---------------------------------------------------------------------------

const colTol = 46

// GAP-01: two lines of Poppins at line-height:normal. folio's leading is
// tighter, so the text block is vertically shorter than Chrome's. Match when
// the block heights agree within 12%.
func checkLineHeight(folio, chrome image.Image) geomResult {
	f := maskBBox(folio, isDark)
	c := maskBBox(chrome, isDark)
	if !f.ok || !c.ok || c.h() == 0 {
		return geomResult{false, "no text found"}
	}
	dev := float64(absi(f.h()-c.h())) / float64(c.h())
	return geomResult{dev <= 0.12, fmt.Sprintf("text-height folio=%d chrome=%d dev=%.0f%% (match<=12%%)", f.h(), c.h(), dev*100)}
}

// GAP-02: a long unbreakable Poppins-Bold token in a fixed-width box. folio's
// wider advances push the token to wrap (roughly doubling its block height)
// where Chrome keeps it on one overflowing line. Match when the token block
// heights agree within 20%.
func checkGlyphAdvance(folio, chrome image.Image) geomResult {
	f := maskBBox(folio, isDark)
	c := maskBBox(chrome, isDark)
	if !f.ok || !c.ok || c.h() == 0 {
		return geomResult{false, "no token found"}
	}
	dev := float64(absi(f.h()-c.h())) / float64(c.h())
	return geomResult{dev <= 0.20, fmt.Sprintf("token-height folio=%d chrome=%d dev=%.0f%% (match<=20%%)", f.h(), c.h(), dev*100)}
}

// GAP-05: an absolutely-positioned <p> (top:12px) inside a relative banner
// whose height comes from padding (its content box collapses to zero). Chrome
// honors position:absolute and paints the text near the banner top (top-y ~22
// at 150dpi), so the banner shows the text. folio ignores the absolute
// positioning and lays the <p> out in normal flow BELOW the banner (top-y
// ~112), leaving the banner's absolute-position region empty. We compare the
// text block's top-y: Match (PASS) only when folio places the text at Chrome's
// position (absolute honored); a large vertical displacement — or the text
// missing entirely — = FAIL (reproduces).
func checkAbsolute(folio, chrome image.Image) geomResult {
	f := maskBBox(folio, isDark)
	c := maskBBox(chrome, isDark)
	if c.count < 200 {
		return geomResult{false, fmt.Sprintf("chrome label not detected (dark=%d)", c.count)}
	}
	if f.count < 200 {
		return geomResult{false, fmt.Sprintf("folio label absent (dark=%d, chrome top-y=%d)", f.count, c.y0)}
	}
	dy := absi(f.y0 - c.y0)
	match := dy <= 20 // folio put the text at the absolute position, like Chrome
	return geomResult{match, fmt.Sprintf("label top-y folio=%d chrome=%d dy=%d (match<=20; folio honors absolute pos=%v)", f.y0, c.y0, dy, match)}
}

// GAP-06: margins on flex children. Measure the main-axis gutter (colB.left -
// colA.right) and the cross-axis offset (colB.top - colA.top). folio drops
// both; Chrome honors them. Match when both agree with Chrome within 10px.
func checkFlexMargins(folio, chrome image.Image) geomResult {
	fa, fb := colorBBox(folio, colBlueA, colTol), colorBBox(folio, colOrange, colTol)
	ca, cb := colorBBox(chrome, colBlueA, colTol), colorBBox(chrome, colOrange, colTol)
	if !fa.ok || !fb.ok || !ca.ok || !cb.ok {
		return geomResult{false, "columns not found"}
	}
	fGut, fOff := fb.x0-fa.x1, fb.y0-fa.y0
	cGut, cOff := cb.x0-ca.x1, cb.y0-ca.y0
	match := absi(fGut-cGut) <= 10 && absi(fOff-cOff) <= 10
	return geomResult{match, fmt.Sprintf("gutter folio=%d chrome=%d | top-offset folio=%d chrome=%d", fGut, cGut, fOff, cOff)}
}

// GAP-07: a display:grid pill. Both engines stretch the grid box to width, so
// the discriminator is the corner: folio draws square corners (ignores
// border-radius), Chrome rounds them. Match when folio's corner rounding
// equals Chrome's (both rounded).
func checkGridPill(folio, chrome image.Image) geomResult {
	f := colorBBox(folio, colPill, colTol)
	c := colorBBox(chrome, colPill, colTol)
	if !f.ok || !c.ok {
		return geomResult{false, "pill not found"}
	}
	fSquare := cornerFilled(folio, f, colPill, colTol, 3, false, false)
	cSquare := cornerFilled(chrome, c, colPill, colTol, 3, false, false)
	return geomResult{fSquare == cSquare, fmt.Sprintf("top-left corner square folio=%v chrome=%v", fSquare, cSquare)}
}

// GAP-08: several inline-block badges back-to-back in a ~48%-width plain block.
// Chrome flows them horizontally and wraps to a few rows; folio lays each out
// as its own block (one badge per line), producing many more row-bands. Match
// only when folio uses the same number of badge row-bands as Chrome.
func checkInlineBlock(folio, chrome image.Image) geomResult {
	f := countRowBands(folio, colBadge, colTol, 5)
	c := countRowBands(chrome, colBadge, colTol, 5)
	return geomResult{f == c && c > 0, fmt.Sprintf("badge row-bands folio=%d chrome=%d", f, c)}
}

// GAP-09: background-color on <thead>. folio paints no header bar; Chrome
// paints a solid grey band (thousands of grey pixels). Match when folio has a
// grey bar iff Chrome does.
func checkTheadBg(folio, chrome image.Image) geomResult {
	f := colorBBox(folio, colGrey, 12)
	c := colorBBox(chrome, colGrey, 12)
	const barMin = 3000 // a painted bar dwarfs stray antialiased greys
	fBar, cBar := f.count >= barMin, c.count >= barMin
	return geomResult{fBar == cBar, fmt.Sprintf("grey-bar folio=%v(%dpx) chrome=%v(%dpx)", fBar, f.count, cBar, c.count)}
}

// rowExtent returns the leftmost and rightmost x of the fill colour on row y
// within b's x-range, or (-1,-1) if none.
func rowExtent(img image.Image, b box, want rgb, tol, y int) (int, int) {
	lx, rx := -1, -1
	for x := b.x0; x < b.x1; x++ {
		if isColor(img.At(x, y), want, tol) {
			if lx < 0 {
				lx = x
			}
			rx = x
		}
	}
	return lx, rx
}

// bottomStraightRun counts consecutive rows upward from the bbox bottom whose
// left (or right) fill edge stays within ±1px of the bottom row's edge. A flat
// square band spilling below the pill has a straight vertical edge reaching the
// very bottom, so this run is long; a rounded bottom's edge moves inward every
// row (the arc), so the run is ~1. Returns the larger of the left/right runs.
func bottomStraightRun(img image.Image, b box, want rgb, tol int) int {
	if !b.ok {
		return 0
	}
	lb, rb := rowExtent(img, b, want, tol, b.y1-1)
	if lb < 0 {
		return 0
	}
	runL, runR := 0, 0
	for y := b.y1 - 1; y >= b.y0; y-- {
		lx, _ := rowExtent(img, b, want, tol, y)
		if lx < 0 || absi(lx-lb) > 1 {
			break
		}
		runL++
	}
	for y := b.y1 - 1; y >= b.y0; y-- {
		_, rx := rowExtent(img, b, want, tol, y)
		if rx < 0 || absi(rx-rb) > 1 {
			break
		}
		runR++
	}
	if runR > runL {
		return runR
	}
	return runL
}

// GAP-11: a fixed-height (34px) flex "rectangle" pill with rounded corners
// whose children are taller than 34px. folio paints the overflow background as
// a SQUARE band that spills below the rounded bottom; Chrome clips the
// background to the border-radius (the bottom stays rounded). We find the
// pill's amber bbox and measure the "bottom straight run" — how many rows up
// from the very bottom keep a straight (constant) vertical edge. A square band
// has a flat vertical edge reaching the bottom (long run); a rounded bottom's
// edge curves inward every row (run ~1). Match (PASS) only when folio's bottom
// is rounded like Chrome; a square band present = FAIL (reproduces).
func checkRadiusOverflow(folio, chrome image.Image) geomResult {
	f := colorBBox(folio, colAmber, colTol)
	c := colorBBox(chrome, colAmber, colTol)
	if !f.ok || !c.ok {
		return geomResult{false, fmt.Sprintf("pill not found folio=%v chrome=%v", f.ok, c.ok)}
	}
	const bandRun = 5 // a flat band reaches the bottom for many rows; an arc does not
	fRun := bottomStraightRun(folio, f, colAmber, colTol)
	cRun := bottomStraightRun(chrome, c, colAmber, colTol)
	fBand := fRun >= bandRun
	cBand := cRun >= bandRun
	return geomResult{fBand == cBand, fmt.Sprintf("bottom-straight-run folio=%d(band=%v) chrome=%d(band=%v)", fRun, fBand, cRun, cBand)}
}

// ---------------------------------------------------------------------------
// Rendering + tool plumbing.
// ---------------------------------------------------------------------------

func chromeBin() string {
	if b := os.Getenv("CHROME_BIN"); b != "" {
		return b
	}
	return "/Applications/Google Chrome.app/Contents/MacOS/Google Chrome"
}

// normalizeHTML forces both engines onto the same page box: A4, zero margin,
// no default body margin. Chrome defaults to US-Letter with printer margins;
// folio defaults to 72pt margins — the injected @page rule pins both to A4/0.
func normalizeHTML(html string) string {
	inject := `<style>@page{size:A4;margin:0}html,body{margin:0;padding:0}</style>`
	if i := strings.Index(strings.ToLower(html), "</head>"); i != -1 {
		return html[:i] + inject + html[i:]
	}
	return inject + html
}

func renderFolio(html, dir string) ([]byte, error) {
	doc := document.NewDocument(document.PageSizeA4)
	opts := &foliohtml.Options{
		PageWidth:  document.PageSizeA4.Width,
		PageHeight: document.PageSizeA4.Height,
		BaseFS:     os.DirFS(dir),
	}
	if err := doc.AddHTMLWithContext(context.Background(), html, opts); err != nil {
		return nil, err
	}
	return doc.ToBytes()
}

// renderChrome writes the normalized HTML into the case dir (so relative
// @font-face url() resolves the same way it does for folio) and prints it to
// PDF with headless Chrome.
func renderChrome(t *testing.T, chrome, dir, html string) (string, bool) {
	t.Helper()
	htmlPath := filepath.Join(dir, ".parity-chrome.html")
	if err := os.WriteFile(htmlPath, []byte(html), 0o600); err != nil {
		t.Fatalf("write chrome html: %v", err)
	}
	t.Cleanup(func() { _ = os.Remove(htmlPath) })

	absHTML, err := filepath.Abs(htmlPath)
	if err != nil {
		t.Fatalf("abs: %v", err)
	}
	outPDF := filepath.Join(t.TempDir(), "chrome.pdf")
	cmd := exec.Command(chrome,
		"--headless", "--disable-gpu", "--no-sandbox",
		"--no-pdf-header-footer",
		"--print-to-pdf="+outPDF,
		"file://"+absHTML,
	)
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Logf("chrome render failed: %v\n%s", err, out)
		return "", false
	}
	if _, err := os.Stat(outPDF); err != nil {
		return "", false
	}
	return outPDF, true
}

// rasterize renders page 1 of pdfPath to a single PNG at 150dpi via pdftoppm
// and decodes it.
func rasterize(t *testing.T, pdftoppm, pdfPath string) (image.Image, bool) {
	t.Helper()
	prefix := filepath.Join(t.TempDir(), strings.TrimSuffix(filepath.Base(pdfPath), ".pdf")+"-p1")
	cmd := exec.Command(pdftoppm, "-png", "-r", "150", "-f", "1", "-l", "1", "-singlefile", pdfPath, prefix)
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Logf("pdftoppm failed: %v\n%s", err, out)
		return nil, false
	}
	png := prefix + ".png"
	f, err := os.Open(png) //nolint:gosec // fixed temp path
	if err != nil {
		return nil, false
	}
	defer func() { _ = f.Close() }()
	img, _, err := image.Decode(f)
	if err != nil {
		t.Logf("decode png: %v", err)
		return nil, false
	}
	return img, true
}

func pdfPageCount(pdfinfo, pdfPath string) int {
	out, err := exec.Command(pdfinfo, pdfPath).Output()
	if err != nil {
		return -1
	}
	for _, line := range strings.Split(string(out), "\n") {
		if strings.HasPrefix(line, "Pages:") {
			n, _ := strconv.Atoi(strings.TrimSpace(strings.TrimPrefix(line, "Pages:")))
			return n
		}
	}
	return -1
}

// decompressPDF returns the raw bytes plus every FlateDecode stream inflated,
// concatenated, so tokens inside compressed object streams (link actions,
// destination trees) are searchable as plain text.
func decompressPDF(pdf []byte) string {
	var b strings.Builder
	b.Write(pdf)
	needle := []byte("stream")
	for i := 0; i+len(needle) < len(pdf); {
		j := bytes.Index(pdf[i:], needle)
		if j < 0 {
			break
		}
		start := i + j + len(needle)
		for start < len(pdf) && (pdf[start] == '\r' || pdf[start] == '\n') {
			start++
		}
		end := bytes.Index(pdf[start:], []byte("endstream"))
		if end < 0 {
			break
		}
		if zr, err := zlib.NewReader(bytes.NewReader(pdf[start : start+end])); err == nil {
			if inflated, err := io.ReadAll(zr); err == nil {
				b.WriteByte('\n')
				b.Write(inflated)
			}
			_ = zr.Close()
		}
		i = start + end + len("endstream")
	}
	return b.String()
}

// folioAnchorBroken reports whether the folio PDF's internal anchor is dead:
// either a dangling /GoTo string action whose named destination was never
// registered (0 /Dests and 0 /Names in the decompressed PDF), or a bare
// external /URI (#id) action. A working link resolves to a registered named
// destination (a /Dests dict) or a direct /Dest array — folio commit 1b17d01
// ("register element ids as PDF named destinations") produces the latter, so
// this returns false (link works) for the current build.
func folioAnchorBroken(pdf []byte) bool {
	blob := decompressPDF(pdf)
	brokenURI := strings.Contains(blob, "/URI (#") || strings.Contains(blob, "/URI(#")
	hasGoTo := strings.Contains(blob, "/GoTo")
	hasRegisteredDest := strings.Contains(blob, "/Dests") || strings.Contains(blob, "/Names")
	danglingGoTo := hasGoTo && !hasRegisteredDest
	return brokenURI || danglingGoTo
}

func TestChromeParity(t *testing.T) {
	chrome := chromeBin()
	if _, err := os.Stat(chrome); err != nil {
		t.Skipf("Chrome not found at %q (set CHROME_BIN); skipping parity harness", chrome)
	}
	pdftoppm, err := exec.LookPath("pdftoppm")
	if err != nil {
		t.Skip("pdftoppm (poppler) not found; skipping parity harness")
	}
	pdfinfo, _ := exec.LookPath("pdfinfo")

	type resultRow struct {
		gap, dir, verdict, detail string
	}
	var rows []resultRow

	// Iterate the gap cases in a stable order.
	names := make([]string, 0, len(caseTable))
	for name := range caseTable {
		names = append(names, name)
	}
	sort.Strings(names)

	for _, name := range names {
		pc := caseTable[name]
		dir := filepath.Join(casesDir, name)
		raw, err := os.ReadFile(filepath.Join(dir, "sample.html"))
		if err != nil {
			t.Errorf("%s: %v", name, err)
			continue
		}

		t.Run(name, func(t *testing.T) {
			html := normalizeHTML(string(raw))

			folioPDF, err := renderFolio(html, dir)
			if err != nil {
				t.Fatalf("folio render: %v", err)
			}
			folioPath := filepath.Join(t.TempDir(), "folio.pdf")
			if err := os.WriteFile(folioPath, folioPDF, 0o600); err != nil {
				t.Fatalf("write folio pdf: %v", err)
			}

			chromePath, ok := renderChrome(t, chrome, dir, html)
			if !ok {
				t.Skip("chrome render unavailable")
			}

			var match bool
			var detail string

			switch pc.kind {
			case kindAnchor:
				broken := folioAnchorBroken(folioPDF)
				match = !broken
				detail = fmt.Sprintf("folio internal link dead (dangling GoTo / no dest)=%v", broken)

			case kindPageBreak:
				if pdfinfo == "" {
					t.Skip("pdfinfo not found")
				}
				fp := pdfPageCount(pdfinfo, folioPath)
				cpg := pdfPageCount(pdfinfo, chromePath)
				match = fp >= cpg && cpg > 0
				detail = fmt.Sprintf("folio pages=%d chrome pages=%d", fp, cpg)

			default: // kindGeom
				fimg, ok1 := rasterize(t, pdftoppm, folioPath)
				cimg, ok2 := rasterize(t, pdftoppm, chromePath)
				if !ok1 || !ok2 {
					t.Skip("rasterization unavailable")
				}
				res := pc.geomFn(fimg, cimg)
				match, detail = res.match, res.detail
			}

			verdict := "FAIL" // folio differs from Chrome (gap reproduces)
			if match {
				verdict = "PASS"
			}
			rows = append(rows, resultRow{pc.gap, name, verdict, detail})
			t.Logf("[%s] %s: folio-vs-chrome=%s (%s)", pc.gap, name, verdict, detail)

			switch {
			case pc.wantReproduce && match:
				t.Errorf("%s (%s): folio now MATCHES Chrome — the gap appears FIXED "+
					"upstream. Verify, then remove the corresponding workaround in "+
					"downstream and set wantReproduce=false here. (%s)", pc.gap, name, detail)
			case !pc.wantReproduce && !match:
				t.Errorf("%s (%s): folio now DIFFERS from Chrome — a construct that "+
					"used to match has REGRESSED. Re-check the gap. (%s)", pc.gap, name, detail)
			}
		})
	}

	sort.Slice(rows, func(i, j int) bool {
		if rows[i].gap != rows[j].gap {
			return rows[i].gap < rows[j].gap
		}
		return rows[i].dir < rows[j].dir
	})
	var b strings.Builder
	b.WriteString("\n=== folio vs Chrome parity ===\n")
	b.WriteString(fmt.Sprintf("%-14s %-32s %-16s %s\n", "GAP", "case", "folio-vs-chrome", "detail"))
	for _, r := range rows {
		b.WriteString(fmt.Sprintf("%-14s %-32s %-16s %s\n", r.gap, r.dir, r.verdict, r.detail))
	}
	b.WriteString("FAIL = folio differs from Chrome (gap reproduces).  " +
		"PASS = folio matches Chrome (gap absent/fixed).\n" +
		"FOLIO-GAP-04 (float) is covered by the float-family cases and inspected manually, not graded here.\n")
	t.Log(b.String())
}
