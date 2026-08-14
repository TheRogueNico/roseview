package main

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

// seedFuseCache points the fuse cache at a fresh temp dir holding a stub
// library, so Render never touches the network or the repo's assets/.
func seedFuseCache(t *testing.T) {
	t.Helper()
	stub := filepath.Join(t.TempDir(), "fuse.min.js")
	if err := os.WriteFile(stub, []byte("stub"), 0o644); err != nil {
		t.Fatal(err)
	}
	t.Setenv("ROSEVIEW_FUSE_CACHE", stub)
}

func TestRender(t *testing.T) {
	seedFuseCache(t)
	dir := t.TempDir()
	out := Output{
		Title: "برنامه درسی",
		Columns: []Column{
			{Field: "code", Header: "كد درس", Group: GroupData, Order: 1},
			{Field: "name", Header: "نام درس", Group: GroupData},
		},
		Rows: []map[string]string{
			{"code": "2661144408", "name": "مبانی"},
		},
		Metadata: []KV{{Header: "نوع ارائه", Value: "حضوري"}},
	}

	if err := Render(dir, out); err != nil {
		t.Fatalf("Render() error = %v", err)
	}

	for _, name := range []string{"index.html", "style.css", "app.js", "fuse.min.js"} {
		if _, err := os.Stat(filepath.Join(dir, name)); err != nil {
			t.Errorf("missing asset %s: %v", name, err)
		}
	}

	index, err := os.ReadFile(filepath.Join(dir, "index.html"))
	if err != nil {
		t.Fatal(err)
	}
	html := string(index)

	for _, want := range []string{
		`data-theme="rose-pine"`,
		`dir="rtl"`,
		"برنامه درسی",
		"نوع ارائه",
		"حضوري",
		"window.ROSEVIEW",
		"style.css",
		"fuse.min.js",
		"app.js",
		"آخرین بروزرسانی",
		"search-icon",
		"No roses were harmed in the making of this project",
	} {
		if !strings.Contains(html, want) {
			t.Errorf("index.html missing %q", want)
		}
	}

	// The last-updated chip must hold a Persian (Jalali) date.
	date := regexp.MustCompile(`[۰-۹]{4}/[۰-۹]{2}/[۰-۹]{2}`)
	if !date.MatchString(html) {
		t.Errorf("index.html missing a Persian date for آخرین بروزرسانی")
	}

	// The JSON payload must be embedded verbatim for the client script.
	for _, want := range []string{
		`"field":"code"`,
		`"header":"كد درس"`,
		`"name":"مبانی"`,
	} {
		if !strings.Contains(html, want) {
			t.Errorf("index.html JSON missing %q", want)
		}
	}

	// Values must be HTML-escaped when rendered by the template.
	if strings.Contains(html, `"<script>"`) {
		t.Error("index.html should not contain raw template data")
	}
}

func TestRender_EmptyTitle(t *testing.T) {
	seedFuseCache(t)
	dir := t.TempDir()
	out := Output{
		Title:   "",
		Columns: []Column{{Field: "name", Header: "نام درس", Group: GroupData}},
		Rows:    []map[string]string{{"name": "مبانی"}},
	}

	if err := Render(dir, out); err != nil {
		t.Fatalf("Render() error = %v", err)
	}

	index, err := os.ReadFile(filepath.Join(dir, "index.html"))
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(index), "<h1") {
		t.Error("expected h1 to be omitted when title is empty")
	}
}
