// Package fuse fetches and caches the Fuse.js fuzzy-search library.
package fuse

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// fuseVersion is the pinned version of the Fuse.js package fetched via npm.
// v6 is used because it ships a UMD build that exposes the global Fuse
// constructor in a classic <script> tag, which works when the page is opened
// straight from disk; v7 ships only ESM/CJS builds, and ESM is blocked over
// file:// by browser CORS.
const fuseVersion = "6.6.2"

// distCandidates are the library files, in preference order, that expose the
// Fuse global when loaded with a plain <script> tag. The UMD build is
// preferred; later package versions may rename the dist artifacts.
var distCandidates = []string{
	"dist/fuse.min.js",
	"dist/fuse.js",
	"dist/fuse.min.mjs",
	"dist/fuse.mjs",
}

// fuseCacheFile returns the path where the fetched library is cached, so
// later builds skip npm. ROSEVIEW_FUSE_CACHE overrides the location (used by
// tests); by default the cache lives in render's assets/ so the embed picks
// it up on the next build.
func fuseCacheFile() string {
	if p := os.Getenv("ROSEVIEW_FUSE_CACHE"); p != "" {
		return p
	}
	return filepath.Join("internal", "render", "assets", "fuse.min.js")
}

// npmFetchFuse installs the pinned fuse.js package into a temp dir with npm
// and returns the dist file bytes.
func npmFetchFuse() ([]byte, error) {
	dir, err := os.MkdirTemp("", "roseview-fuse-*")
	if err != nil {
		return nil, fmt.Errorf("creating temp dir: %w", err)
	}
	defer os.RemoveAll(dir)

	cmd := exec.Command("npm", "install",
		"--prefix", dir,
		"--no-save",
		"--no-audit",
		"--no-fund",
		"--no-package-lock",
		"fuse.js@"+fuseVersion,
	)
	if out, err := cmd.CombinedOutput(); err != nil {
		return nil, fmt.Errorf("running npm install fuse.js@%s: %w (%s)",
			fuseVersion, err, strings.TrimSpace(string(out)))
	}
	return pickDistFile(dir)
}

// pickDistFile returns the first existing, non-empty library file inside the
// installed fuse.js package.
func pickDistFile(pkgDir string) ([]byte, error) {
	for _, rel := range distCandidates {
		p := filepath.Join(pkgDir, "node_modules", "fuse.js", rel)
		if b, err := os.ReadFile(p); err == nil && len(b) > 0 {
			return b, nil
		}
	}
	return nil, fmt.Errorf("no dist file found in fuse.js@%s (tried %s)",
		fuseVersion, strings.Join(distCandidates, ", "))
}

// Ensure returns the Fuse.js library bytes, fetching and caching them on
// first use. A failed fetch with no cached copy is an error.
func Ensure() ([]byte, error) {
	cache := fuseCacheFile()
	if b, err := os.ReadFile(cache); err == nil {
		return b, nil
	}

	b, err := npmFetchFuse()
	if err != nil {
		return nil, fmt.Errorf("fetching fuse.js (no cached copy in %s): %w", cache, err)
	}

	if err := os.MkdirAll(filepath.Dir(cache), 0o755); err != nil {
		return nil, fmt.Errorf("creating cache dir for fuse.js: %w", err)
	}
	if err := os.WriteFile(cache, b, 0o644); err != nil {
		return nil, fmt.Errorf("caching fuse.js in %s: %w", cache, err)
	}
	return b, nil
}
