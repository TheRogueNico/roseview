package parse

import (
	"fmt"
	"io"
	"strings"

	"golang.org/x/net/html"
)

// Table holds the header and body cells of a single <table> element. Cells
// are kept as raw text; normalization happens later at bind time.
type Table struct {
	Headers []string
	Rows    [][]string
}

// ParseTables extracts every <table> from r, returning them in document
// order. Nested tables are ignored to avoid double counting.
func ParseTables(r io.Reader) ([]Table, error) {
	doc, err := html.Parse(r)
	if err != nil {
		return nil, fmt.Errorf("parsing html: %w", err)
	}

	var tables []Table
	var visit func(*html.Node)
	visit = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "table" {
			tables = append(tables, parseTable(n))
			return
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			visit(c)
		}
	}
	visit(doc)
	return tables, nil
}

// parseTable extracts the header row (from <thead>) and body rows (from
// <tbody>) of a single table.
func parseTable(table *html.Node) Table {
	var t Table
	for c := table.FirstChild; c != nil; c = c.NextSibling {
		if c.Type != html.ElementNode {
			continue
		}
		switch c.Data {
		case "thead":
			if rows := parseRows(c, "th"); len(rows) > 0 {
				t.Headers = rows[0]
			}
		case "tbody":
			for _, row := range parseRows(c, "td") {
				t.Rows = append(t.Rows, row)
			}
		}
	}
	return t
}

// parseRows collects the text of every cellTag cell inside each <tr> child
// of parent. Rows without any matching cells are skipped.
func parseRows(parent *html.Node, cellTag string) [][]string {
	var rows [][]string
	for c := parent.FirstChild; c != nil; c = c.NextSibling {
		if c.Type != html.ElementNode || c.Data != "tr" {
			continue
		}
		var row []string
		for cell := c.FirstChild; cell != nil; cell = cell.NextSibling {
			if cell.Type == html.ElementNode && cell.Data == cellTag {
				row = append(row, textContent(cell))
			}
		}
		if len(row) > 0 {
			rows = append(rows, row)
		}
	}
	return rows
}

// textContent returns the concatenated text of all descendant text nodes.
func textContent(n *html.Node) string {
	var sb strings.Builder
	var walk func(*html.Node)
	walk = func(n *html.Node) {
		if n.Type == html.TextNode {
			sb.WriteString(n.Data)
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			walk(c)
		}
	}
	walk(n)
	return sb.String()
}
