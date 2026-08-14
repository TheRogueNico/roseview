package main

import (
	"fmt"
	"sort"
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
	Columns  []Column
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
func Build(cfg Config, tables []Table, title string) (Output, error) {
	out := Output{Title: title}
	if len(tables) == 0 {
		return out, fmt.Errorf("input contains no tables")
	}

	type mapping struct {
		column Column
		src    int
	}

	// Resolve each configured column to a source index. The first table
	// provides the reference headers; the rest are expected to match it.
	mappings := make([]mapping, 0, len(cfg.Columns))
	srcIndex := make(map[string]int, len(cfg.Columns))
	for _, col := range cfg.Columns {
		idx := findHeader(tables[0].Headers, col.Header)
		if idx < 0 {
			return out, fmt.Errorf("column %q not found in table headers", col.Header)
		}
		mappings = append(mappings, mapping{column: col, src: idx})
		srcIndex[col.Field] = idx
	}

	// Split columns into display and metadata groups.
	var metaCols []mapping
	for _, m := range mappings {
		switch m.column.Group {
		case GroupData:
			col := m.column
			col.Header = Normalize(col.Header)
			out.Columns = append(out.Columns, col)
		case GroupMetadata:
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
	// are excluded entirely.
	for _, tbl := range tables {
		for _, cells := range tbl.Rows {
			row := make(map[string]string, len(mappings))
			for _, m := range mappings {
				if m.column.Group == GroupSkip {
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
				Header: Normalize(m.column.Header),
				Value:  out.Rows[0][m.column.Field],
			})
		}
	}

	return out, nil
}

// findHeader returns the index of want within headers, or -1 if absent.
func findHeader(headers []string, want string) int {
	want = Normalize(want)
	for i, h := range headers {
		if Normalize(h) == want {
			return i
		}
	}
	return -1
}

// cellAt returns the normalized text of the cell at index i, or "" when the
// row has fewer cells than expected.
func cellAt(cells []string, i int) string {
	if i < 0 || i >= len(cells) {
		return ""
	}
	return Normalize(cells[i])
}
