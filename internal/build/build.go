// Package build binds parsed tables to the column configuration.
package build

import (
	"fmt"
	"sort"

	"github.com/TheRogueNico/roseview/internal/config"
	"github.com/TheRogueNico/roseview/internal/normalize"
	"github.com/TheRogueNico/roseview/internal/parse"
)

// KV is an ordered key/value pair, used for the metadata shown above the table.
type KV struct {
	Header string
	Value  string
}

// Output is the fully bound model handed to the renderer. Column order
// reflects the configured sort order, not the source layout.
type Output struct {
	Title    string
	Columns  []config.Column
	Rows     []map[string]string
	Metadata []KV
}

// Build binds parsed tables to the column configuration, producing the
// normalized output model.
//
// Every configured column is resolved to a source column index by matching
// its (normalized) header text against the first table's headers. Data
// columns are kept for display, metadata columns are folded into the
// constant Metadata list, and skip columns are dropped. All cell text is
// run through Normalize.
func Build(cfg config.Config, tables []parse.Table, title string) (Output, error) {
	out := Output{Title: title}
	if len(tables) == 0 {
		return out, fmt.Errorf("input contains no tables")
	}

	type mapping struct {
		column config.Column
		src    int
	}

	// Resolve each configured column to a source index. The first table
	// provides the reference headers; the rest are expected to match it.
	// Skip columns are dropped anyway, so a missing header for one does not
	// fail the build.
	mappings := make([]mapping, 0, len(cfg.Columns))
	srcIndex := make(map[string]int, len(cfg.Columns))
	for _, col := range cfg.Columns {
		idx := findHeader(tables[0].Headers, col.Header)
		if idx < 0 {
			if col.Group == config.GroupSkip {
				continue
			}
			return out, fmt.Errorf("column %q not found in table headers", col.Header)
		}
		mappings = append(mappings, mapping{column: col, src: idx})
		srcIndex[col.Field] = idx
	}

	// Split columns into display and metadata groups.
	var metaCols []mapping
	for _, m := range mappings {
		switch m.column.Group {
		case config.GroupData:
			col := m.column
			col.Header = normalize.Normalize(col.Header)
			out.Columns = append(out.Columns, col)
		case config.GroupMetadata:
			metaCols = append(metaCols, m)
		}
	}

	// Explicitly ordered columns come first (ascending); the rest keep their
	// source order.
	sort.Slice(out.Columns, func(i, j int) bool {
		oi, oj := out.Columns[i].Order, out.Columns[j].Order
		switch {
		case oi > 0 && oj == 0:
			return true
		case oi == 0 && oj > 0:
			return false
		case oi != oj:
			return oi < oj
		default:
			return srcIndex[out.Columns[i].Field] < srcIndex[out.Columns[j].Field]
		}
	})

	// Extract rows across every table, normalizing each cell. Skip columns
	// are excluded entirely. Later tables must share the reference header
	// layout, otherwise their rows would silently map to wrong columns.
	for ti, tbl := range tables {
		if ti > 0 {
			if err := checkHeaders(tbl.Headers, tables[0].Headers); err != nil {
				return out, fmt.Errorf("table %d: %w", ti+1, err)
			}
		}
		for _, cells := range tbl.Rows {
			row := make(map[string]string, len(mappings))
			for _, m := range mappings {
				if m.column.Group == config.GroupSkip {
					continue
				}
				row[m.column.Field] = cellAt(cells, m.src)
			}
			out.Rows = append(out.Rows, row)
		}
	}

	// Metadata is constant across rows by definition; show the first row's.
	if len(out.Rows) > 0 {
		for _, m := range metaCols {
			out.Metadata = append(out.Metadata, KV{
				Header: normalize.Normalize(m.column.Header),
				Value:  out.Rows[0][m.column.Field],
			})
		}
	}

	return out, nil
}

// findHeader returns the index of want within headers, or -1 if absent.
func findHeader(headers []string, want string) int {
	want = normalize.Normalize(want)
	for i, h := range headers {
		if normalize.Normalize(h) == want {
			return i
		}
	}
	return -1
}

// checkHeaders verifies that headers line up with the reference header row
// (positionally, after normalization). Rows are mapped by column index, so
// a diverging header layout would silently misalign every row.
func checkHeaders(headers, ref []string) error {
	if len(headers) != len(ref) {
		return fmt.Errorf("header count %d does not match reference %d", len(headers), len(ref))
	}
	for i := range ref {
		if normalize.Normalize(headers[i]) != normalize.Normalize(ref[i]) {
			return fmt.Errorf("header %q does not match reference %q", headers[i], ref[i])
		}
	}
	return nil
}

// cellAt returns the normalized text of the cell at index i, or "" when the
// row has fewer cells than expected.
func cellAt(cells []string, i int) string {
	if i < 0 || i >= len(cells) {
		return ""
	}
	return normalize.Normalize(cells[i])
}
