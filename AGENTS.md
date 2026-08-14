# AGENTS.md

Single-package Go CLI (no `internal/` dirs). Module `github.com/TheRogueNico/roseview`.

## Commands

- `go test ./...` — all tests run offline; they stub the Fuse.js cache via `ROSEVIEW_FUSE_CACHE` or swap the `fetchFuse`/`npmCommand` globals (see `fuse_test.go`, `render_test.go`). Never make tests depend on the network.
- `go build .` — produces the `roseview` binary (gitignored).
- Local smoke test: `go run . -input example.html` (writes to `out/`, gitignored).

## Pipeline

`main.go` drives: `ParseTables` (parse.go) → `Build` (build.go, binds rows to `config.json` columns) → `Render` (render.go, emits `index.html`/`style.css`/`app.js`/`fuse.min.js`). `Config` column `group` is `skip` | `data` | `metadata`; positive `order` sorts columns first (ascending), the rest keep source order.

## Gotchas

- `assets/*` is `go:embed`-ed into the binary. `assets/fuse.min.js` is gitignored and fetched at first render via `npm install fuse.js@6.6.2` into a temp dir, then cached; a real render requires `npm` on PATH. Version is pinned to v6 because only v6 ships a UMD build usable over `file://` (v7 is ESM-only, blocked by CORS).
- `Normalize` (normalize.go) is applied to both config headers and cell text before matching: Arabic forms map to Persian (`ك`→`ک`, `ي`→`ی`), harakat/ZWNJ are stripped, whitespace collapsed. So `config.json` headers match source HTML regardless of variant spelling — keep that behavior when editing.
- `ParseTables` only reads `<thead>/<th>` headers and `<tbody>/<td>` rows of top-level tables; nested tables are skipped, cells are raw text (normalization happens in `Build`).
- Output text is Persian (RTL); `Render` prepends a `آخرین بروزرسانی` metadata entry with a Jalali date (jalali.go, Borkowski algorithm, valid 1900–2100).
- `config.json` is the column config for the real input; the shipped config contains Arabic-kaf headers (e.g. `كد درس`) that only match after `Normalize`.