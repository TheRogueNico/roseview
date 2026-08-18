// Package render emits the static site from the bound model.
package render

import (
	"embed"
	"encoding/json"
	"fmt"
	"html/template"
	"os"
	"path/filepath"
	"time"

	"github.com/TheRogueNico/roseview/internal/build"
)

//go:embed assets/*
var assets embed.FS

// Render writes index.html, rose-pine.css, style.css, app.js, rose.svg and
// fuse.min.js into outDir. The page embeds the bound data as JSON; the
// client-side script renders and filters the table from it. fuse.min.js is
// vendored in assets/, so the generated site works fully offline.
// rose-pine.css is the theme's single color source; style.css and the SVG
// icons reference its variables instead of hardcoding colors.
func Render(outDir string, out build.Output) error {
	if err := os.MkdirAll(outDir, 0o755); err != nil {
		return fmt.Errorf("creating output dir: %w", err)
	}

	// The generated page records when it was created. Prepend it to the
	// metadata so it appears first above the table.
	out.Metadata = append([]build.KV{{
		Header: "آخرین بروزرسانی",
		Value:  time.Now().Format("2006/01/02"),
	}}, out.Metadata...)

	tmpl, err := template.ParseFS(assets, "assets/index.tmpl")
	if err != nil {
		return fmt.Errorf("parsing index template: %w", err)
	}

	// encoding/json escapes <, >, & and U+2028/U+2029, so the payload is safe
	// to inline verbatim as JavaScript.
	data, err := json.Marshal(out)
	if err != nil {
		return fmt.Errorf("encoding page data: %w", err)
	}

	page := map[string]any{
		"Title":    out.Title,
		"Metadata": out.Metadata,
		"Data":     template.JS(data),
	}

	f, err := os.Create(filepath.Join(outDir, "index.html"))
	if err != nil {
		return fmt.Errorf("creating index.html: %w", err)
	}

	if err := tmpl.Execute(f, page); err != nil {
		f.Close()
		return fmt.Errorf("rendering index.html: %w", err)
	}
	if err := f.Close(); err != nil {
		return fmt.Errorf("closing index.html: %w", err)
	}

	for _, name := range []string{"rose-pine.css", "style.css", "app.js", "rose.svg", "fuse.min.js"} {
		if err := copyAsset(name, outDir); err != nil {
			return err
		}
	}

	return nil
}

// copyAsset writes a single embedded asset into outDir.
func copyAsset(name, outDir string) error {
	b, err := assets.ReadFile("assets/" + name)
	if err != nil {
		return fmt.Errorf("reading %s: %w", name, err)
	}
	if err := os.WriteFile(filepath.Join(outDir, name), b, 0o644); err != nil {
		return fmt.Errorf("writing %s: %w", name, err)
	}
	return nil
}
