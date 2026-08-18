# roseview

Turn raw HTML course-schedule tables into a static, searchable web page.

The CLI parses the HTML export (e.g. an HTML Tidy dump of a course catalog),
binds it to a column configuration, normalizes the Persian text, and emits a
static site: `index.html`, `rose-pine.css`, `style.css`, `app.js`, `rose.svg`,
and `fuse.min.js`.

The generated page features a clean, right-to-left (RTL) UI themed with
[Rose Pine](https://rosepinetheme.com/) (dark), a fuzzy search bar that
matches across columns with per-cell highlighting (every non-empty query is
searched; Fuse decides match quality on its own), sortable columns, a
bookmark column that pins selected rows into a separate pinned table (shown
only while items are pinned, search-independent, sortable on its own), and a
low-transparency "empty" icon for blank cells. All colors are defined once in
`rose-pine.css` (the official `--rp-*` palette variables) and referenced
everywhere else, so the theme can be swapped without touching any other file.

## Usage

```sh
roseview -input /path/to/schedule.html
```

All flags are optional except `-input`:

| Flag      | Default                    | Description                                |
| --------- | -------------------------- | ------------------------------------------ |
| `-input`  | _(required)_               | Path to the HTML course-schedule export    |
| `-out`    | `site`                     | Directory to write the generated site into |
| `-config` | `config.json`              | Path to the column config JSON             |
| `-title`  | `دروس تخصصی علوم کامپیوتر` | Page title shown in the site header        |

## Output

The `-out` directory receives:

- `index.html` — page skeleton with the extracted data embedded as JSON
- `rose-pine.css` — Rose Pine (dark) palette; the single color source for the
  whole site, defining the official `--rp-*` theme variables
- `style.css` — layout and styling, referencing only theme variables
- `app.js` — rendering, fuzzy search, and sorting
- `rose.svg` — favicon
- `fuse.min.js` — the Fuse.js fuzzy-search library, bundled with the site

Open `index.html` in a browser; no server or network is needed. Fuse.js
(v7.2.0, the last release shipping a browser UMD build) is vendored in
`internal/render/assets/fuse.min.js` and bundled into the output, so the
generated page is fully offline.

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
- `internal/normalize` — Persian text normalization

## License

BSD Zero Clause — see `LICENSE`. UI icons are from
[Lucide](https://lucide.dev/) (ISC license).

