# roseview

Turn raw HTML course-schedule tables into a static, searchable web page.

The CLI parses the HTML export (e.g. an HTML Tidy dump of a course catalog),
binds it to a column configuration, normalizes the Persian text, and emits a
static site: `index.html`, `style.css`, and `app.js`.

The generated page features a clean, right-to-left (RTL) UI themed with
[Rose Pine](https://rosepinetheme.com/) (dark) and Rose Pine Dawn (light,
switchable), a fuzzy search bar with match highlighting that matches across
columns, sortable columns, and a low-transparency "empty" icon for blank
cells.

## Usage

```sh
roseview -input example.html
```

All flags are optional except `-input`:

| Flag      | Default      | Description                                  |
|-----------|--------------|----------------------------------------------|
| `-input`  | *(required)* | Path to the HTML course-schedule export      |
| `-out`    | `site`      | Directory to write the generated site into   |
| `-config` | `config.json`| Path to the column config JSON               |
| `-title`  | `برنامه درسی` | Page title shown in the site header          |

## Output

The `-out` directory receives:

- `index.html` — page skeleton with the extracted data embedded as JSON
- `style.css`  — Rose Pine / Rose Pine Dawn themes and layout
- `app.js`     — rendering, fuzzy search, sorting, and theme toggle
- `fuse.min.js` — the Fuse.js fuzzy-search library, bundled with the site

Open `index.html` in a browser; no server or network is needed.

The first build fetches Fuse.js via npm and caches it in
`assets/fuse.min.js`; later builds reuse the cache, and the generated page
itself is fully offline.

## Configuration

`config.json` maps source-table columns to fields. Each entry has:

- `field` — stable identifier used in the output data
- `header` — the column header text as it appears in the source HTML
- `group` — `skip` (ignored), `data` (shown as a table column), or
  `metadata` (constant across rows, shown once above the table)
- `order` — optional display priority; ordered columns come first
  (ascending), then remaining columns in source order

## Development

```sh
go test ./...
go build .
```

The command lives at the repo root (`main.go`); the pipeline is split into
supporting packages under `internal/`:

- `internal/config` — column configuration loading
- `internal/parse` — HTML table extraction
- `internal/build` — binds parsed tables to the config
- `internal/render` — emits the static site (embeds `assets/`)
- `internal/fuse` — fetches/caches the Fuse.js library
- `internal/normalize` — Persian text normalization