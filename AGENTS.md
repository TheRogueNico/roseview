# AGENTS.md

Go CLI with supporting packages. The `main.go` command lives at the repo root; all supporting logic lives in packages under `internal/`. Module `github.com/TheRogueNico/roseview`.

## Commands

- `go test ./...` — all tests run offline; they stub the Fuse.js cache via `ROSEVIEW_FUSE_CACHE` (see `internal/render/render_test.go`). Never make tests depend on the network.
- `go build .` — produces the `roseview` binary (gitignored).
- Local smoke test: `go run . -input example.html` (writes to `site/`, gitignored).

## Pipeline

`main.go` drives: `config.LoadConfig` (`internal/config`) → `parse.ParseTables` (`internal/parse`) → `build.Build` (`internal/build`, binds rows to `config.json` columns) → `render.Render` (`internal/render`, emits `index.html`/`style.css`/`app.js`/`fuse.min.js`). `config.Config` column `group` is `skip` | `data` | `metadata`; positive `order` sorts columns first (ascending), the rest keep source order.

Package dependencies (acyclic): `main` → config, parse, build, render; `render` → build, fuse; `build` → config, parse, normalize.

## Gotchas

- `internal/render/assets/*` is `go:embed`-ed into the binary. `internal/render/assets/fuse.min.js` is gitignored and fetched at first render via `npm install fuse.js@6.6.2` into a temp dir, then cached there (or at `ROSEVIEW_FUSE_CACHE`); a real render requires `npm` on PATH. Version is pinned to v6 because only v6 ships a UMD build usable over `file://` (v7 is ESM-only, blocked by CORS).
- `normalize.Normalize` (`internal/normalize`) is applied to both config headers and cell text before matching: Arabic forms map to Persian (`ك`→`ک`, `ي`→`ی`), harakat/ZWNJ are stripped, whitespace collapsed. So `config.json` headers match source HTML regardless of variant spelling — keep that behavior when editing.
- `parse.ParseTables` only reads `<thead>/<th>` headers and `<tbody>/<td>` rows of top-level tables; nested tables are skipped, cells are raw text (normalization happens in `build.Build`).
- Output text is Persian (RTL); `render.Render` prepends a `آخرین بروزرسانی` metadata entry with the system date in Gregorian `YYYY/MM/DD` form.
- `config.json` is the column config for the real input; the shipped config contains Arabic-kaf headers (e.g. `كد درس`) that only match after `normalize.Normalize`.