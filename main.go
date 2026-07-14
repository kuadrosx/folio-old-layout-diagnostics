package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/carlos7ags/folio/document"
	foliohtml "github.com/carlos7ags/folio/html"
)

// This program renders every cases/<name>/sample.html to cases/<name>/output.pdf
// using folio, so each case folder is a self-contained, independently reportable
// reproduction (sample.html + issue.md + the rendered output.pdf).
//
//	go run .
const casesDir = "cases"

func render(htmlStr, dir, path string) error {
	doc := document.NewDocument(document.PageSizeA4)

	opts := &foliohtml.Options{
		PageWidth:  document.PageSizeA4.Width,
		PageHeight: document.PageSizeA4.Height,
		// Resolve local assets (e.g. @font-face url('Poppins-Regular.ttf'))
		// relative to the case folder, so font-metric cases can ship their
		// own TTFs case-locally. Chrome resolves the same relative url()
		// against the sample.html location, keeping both engines in sync.
		BaseFS: os.DirFS(dir),
	}

	if err := doc.AddHTMLWithContext(context.Background(), htmlStr, opts); err != nil {
		return err
	}

	pdf, err := doc.ToBytes()
	if err != nil {
		return err
	}

	return os.WriteFile(path, pdf, 0o600)
}

func main() {
	entries, err := os.ReadDir(casesDir)
	if err != nil {
		panic(err)
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		dir := filepath.Join(casesDir, entry.Name())
		samplePath := filepath.Join(dir, "sample.html")

		htmlBytes, err := os.ReadFile(samplePath) //nolint:gosec // fixed repo-local path
		if err != nil {
			fmt.Printf("%-18s SKIP (no sample.html): %v\n", entry.Name(), err)
			continue
		}

		outPath := filepath.Join(dir, "output.pdf")
		if err := render(string(htmlBytes), dir, outPath); err != nil {
			fmt.Printf("%-18s ERROR: %v\n", entry.Name(), err)
			continue
		}

		fmt.Printf("%-18s wrote %s\n", entry.Name(), outPath)
	}
}
