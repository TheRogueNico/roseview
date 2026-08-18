package render

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/TheRogueNico/roseview/internal/build"
	"github.com/TheRogueNico/roseview/internal/config"
)

func TestRender(t *testing.T) {
	dir := t.TempDir()
	out := build.Output{
		Title: "برنامه درسی",
		Columns: []config.Column{
			{Field: "code", Header: "كد درس", Group: config.GroupData, Order: 1},
			{Field: "name", Header: "نام درس", Group: config.GroupData},
		},
		Rows: []map[string]string{
			{"code": "2661144408", "name": "مبانی"},
		},
		Metadata: []build.KV{{Header: "نوع ارائه", Value: "حضوري"}},
	}

	if err := Render(dir, out); err != nil {
		t.Fatalf("Render() error = %v", err)
	}

	for _, name := range []string{"index.html", "rose-pine.css", "style.css", "app.js", "fuse.min.js", "rose.svg"} {
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
		`dir="rtl"`,
		"برنامه درسی",
		"نوع ارائه",
		"حضوري",
		"window.ROSEVIEW",
		"rose-pine.css",
		"style.css",
		"fuse.min.js",
		"app.js",
		`rel="icon"`,
		"rose.svg",
		"icon-search",
		"icon-minus",
		"icon-chevron-up",
		"icon-chevron-down",
		"icon-bookmark-plus",
		"icon-bookmark-minus",
		"pinned-wrap",
		"pinned-head",
		"آخرین بروزرسانی",
		"search-icon",
		"No roses were harmed in the making of this project",
	} {
		if !strings.Contains(html, want) {
			t.Errorf("index.html missing %q", want)
		}
	}

	// The last-updated chip must hold a Gregorian (system) date.
	date := regexp.MustCompile(`\d{4}/\d{2}/\d{2}`)
	if !date.MatchString(html) {
		t.Errorf("index.html missing a Gregorian date for آخرین بروزرسانی")
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
	dir := t.TempDir()
	out := build.Output{
		Title:   "",
		Columns: []config.Column{{Field: "name", Header: "نام درس", Group: config.GroupData}},
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
