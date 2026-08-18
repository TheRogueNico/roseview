# AGENTS.md

Go CLI with supporting packages. The `main.go` command lives at the repo root; all supporting logic lives in packages under `internal/`. Module `github.com/TheRogueNico/roseview`.

## Commands

- `go test ./...` — all tests run offline; never make tests depend on the network.
- `go build .` — produces the `roseview` binary (gitignored).
- Local smoke test: `go run . -input table.html` (writes to `site/`, gitignored; `table.html` is the gitignored sample export).

## Pipeline

`main.go` drives: `config.LoadConfig` (`internal/config`) → `parse.ParseTables` (`internal/parse`) → `build.Build` (`internal/build`, binds rows to `config.json` columns) → `render.Render` (`internal/render`, emits `index.html`/`rose-pine.css`/`style.css`/`app.js`/`fuse.min.js`). `config.Config` column `group` is `skip` | `data` | `metadata`; positive `order` sorts columns first (ascending), the rest keep source order.

Package dependencies (acyclic): `main` → config, parse, build, render; `render` → build; `build` → config, parse, normalize.

## Gotchas

- `internal/render/assets/*` is `go:embed`-ed into the binary. `internal/render/assets/fuse.min.js` is vendored Fuse.js v7.2.0, pinned because it is the last release shipping a browser UMD build (global `Fuse` via plain `<script>`, works over `file://`); 7.3.0+ ships ESM/CJS only, blocked by CORS. Do not bump it without a compatible browser build.
- `app.js` runs every non-empty query through Fuse and intersects per-word hits across columns (multi-word queries can match in different columns). There is deliberately no minimum-query-length or `minMatchCharLength` filter — Fuse's own defaults judge match quality, including single-character queries. Do not re-add length-based filters; they predate the Fuse-based search.
- `normalize.Normalize` (`internal/normalize`) is applied to both config headers and cell text before matching: Arabic forms map to Persian (`ك`→`ک`, `ي`→`ی`), harakat/ZWNJ are stripped, whitespace collapsed. So `config.json` headers match source HTML regardless of variant spelling — keep that behavior when editing.
- `parse.ParseTables` only reads `<thead>/<th>` headers and `<tbody>/<td>` rows of top-level tables; nested tables are skipped, cells are raw text (normalization happens in `build.Build`). Cells whose visible text ends with `...` fall back to the cell's `title` attribute (some source exports truncate the text and keep the full string in the tooltip).
- Output text is Persian (RTL); `render.Render` prepends a `آخرین بروزرسانی` metadata entry with the system date in Gregorian `YYYY/MM/DD` form.
- UI icons are Lucide SVGs in `internal/render/assets/` (`rose.svg`, `search.svg`, `chevron-up.svg`, `chevron-down.svg`, `minus.svg`). They are inlined into `index.tmpl` as a sprite and referenced via same-document `<use href="#icon-…">` — do not move them to an external sprite file, which would fail over `file://` (CORS). Only `rose.svg` is copied to the output as the favicon; the rest are compile-time sources for the sprite.
- `rose-pine.css` is the single source of truth for colors (official Rose Pine dark palette as `--rp-*` variables). `style.css` and the sprite reference it via `var(--rp-…)`; sprite icons use `stroke="currentColor"` and get their color from CSS classes in `style.css`. Do not hardcode hex colors in `style.css` or `index.tmpl`. The favicon `rose.svg` is standalone and embeds its own copy of the palette variables it uses in a `<style>` block (SVG presentation attributes cannot use `var()`).
- `config.json` is the column config for the real input; the shipped config contains Arabic-kaf headers (e.g. `كد درس`) that only match after `normalize.Normalize`.