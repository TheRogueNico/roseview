package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestEnsureFuse_CacheHit(t *testing.T) {
	cache := filepath.Join(t.TempDir(), "fuse.min.js")
	if err := os.WriteFile(cache, []byte("cached"), 0o644); err != nil {
		t.Fatal(err)
	}
	t.Setenv("ROSEVIEW_FUSE_CACHE", cache)

	b, err := ensureFuse()
	if err != nil {
		t.Fatalf("ensureFuse() error = %v", err)
	}
	if string(b) != "cached" {
		t.Errorf("ensureFuse() = %q, want %q", b, "cached")
	}
}

func TestEnsureFuse_Fetch(t *testing.T) {
	orig := fetchFuse
	fetchFuse = func() ([]byte, error) { return []byte("library"), nil }
	defer func() { fetchFuse = orig }()

	cache := filepath.Join(t.TempDir(), "fuse.min.js")
	t.Setenv("ROSEVIEW_FUSE_CACHE", cache)

	b, err := ensureFuse()
	if err != nil {
		t.Fatalf("ensureFuse() error = %v", err)
	}
	if string(b) != "library" {
		t.Errorf("ensureFuse() = %q, want %q", b, "library")
	}

	cached, err := os.ReadFile(cache)
	if err != nil {
		t.Fatalf("reading cache after fetch: %v", err)
	}
	if string(cached) != "library" {
		t.Errorf("cached file = %q, want %q", cached, "library")
	}
}

func TestEnsureFuse_FetchFailure(t *testing.T) {
	orig := fetchFuse
	fetchFuse = func() ([]byte, error) { return nil, os.ErrNotExist }
	defer func() { fetchFuse = orig }()

	cache := filepath.Join(t.TempDir(), "fuse.min.js")
	t.Setenv("ROSEVIEW_FUSE_CACHE", cache)

	if _, err := ensureFuse(); err == nil {
		t.Fatal("ensureFuse() succeeded, want error")
	}
	if _, err := os.Stat(cache); !os.IsNotExist(err) {
		t.Error("ensureFuse() wrote a cache file despite the failed fetch")
	}
}

func TestPickDistFile(t *testing.T) {
	root := t.TempDir()
	pkg := filepath.Join(root, "node_modules", "fuse.js")
	if err := os.MkdirAll(filepath.Join(pkg, "dist"), 0o755); err != nil {
		t.Fatal(err)
	}
	umd := filepath.Join(pkg, "dist", "fuse.min.js")
	esm := filepath.Join(pkg, "dist", "fuse.mjs")
	if err := os.WriteFile(umd, []byte("umd"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(esm, []byte("esm"), 0o644); err != nil {
		t.Fatal(err)
	}

	b, err := pickDistFile(root)
	if err != nil {
		t.Fatalf("pickDistFile() error = %v", err)
	}
	if string(b) != "umd" {
		t.Errorf("pickDistFile() = %q, want the UMD build %q", b, "umd")
	}

	if err := os.Remove(umd); err != nil {
		t.Fatal(err)
	}
	b, err = pickDistFile(root)
	if err != nil {
		t.Fatalf("pickDistFile() without UMD build error = %v", err)
	}
	if string(b) != "esm" {
		t.Errorf("pickDistFile() = %q, want fallback %q", b, "esm")
	}

	if err := os.Remove(esm); err != nil {
		t.Fatal(err)
	}
	if _, err := pickDistFile(root); err == nil {
		t.Fatal("pickDistFile() succeeded, want error with no dist files")
	}
}

func TestNpmFetchFuse_MissingNpm(t *testing.T) {
	orig := npmCommand
	npmCommand = "roseview-no-such-npm-binary"
	defer func() { npmCommand = orig }()

	if _, err := npmFetchFuse(); err == nil {
		t.Fatal("npmFetchFuse() succeeded, want error for missing npm")
	}
}
