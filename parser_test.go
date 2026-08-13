package main

import (
	"reflect"
	"testing"
)

func TestParseHTML(t *testing.T) {
	doc := []byte(`
		<html>
		<body>
			<div class="term">Fall 2026</div>
			<table>
				<tr>
					<td>Alice</td>
					<td>CS101</td>
				</tr>
				<tr>
					<td>Bob</td>
					<td>CS202</td>
				</tr>
			</table>
		</body>
		</html>
	`)

	cfg := Config{
		Columns: []Column{
			{Field: "name", Header: "Name", Group: GroupData, Order: 1},
			{Field: "course", Header: "Course", Group: GroupData, Order: 2},
			{Field: "term", Header: "Term", Group: GroupMetadata, Order: 0},
		},
	}

	got, err := ParseHTML(doc, cfg)
	if err != nil {
		t.Fatalf("ParseHTML() error = %v", err)
	}

	want := Dataset{
		Metadata: []MetaEntry{
			{
				Field:  "term",
				Header: "Term",
				Value:  "Fall 2026",
			},
		},
		Columns: []Column{
			{Field: "name", Header: "Name", Group: GroupData, Order: 1},
			{Field: "course", Header: "Course", Group: GroupData, Order: 2},
		},
		Rows: []map[string]string{
			{
				"name":   "Alice",
				"course": "CS101",
			},
			{
				"name":   "Bob",
				"course": "CS202",
			},
		},
	}

	if !reflect.DeepEqual(got, want) {
		t.Errorf("ParseHTML() = %+v, want %+v", got, want)
	}
}
