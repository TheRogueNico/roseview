package main

import (
	"reflect"
	"testing"
)

func TestBuild(t *testing.T) {
	cfg := Config{Columns: []Column{
		{Field: "expand", Header: "expand", Group: GroupSkip},
		{Field: "code", Header: "كد درس", Group: GroupData, Order: 1},
		{Field: "name", Header: "نام درس", Group: GroupData},
		{Field: "delivery", Header: "نوع ارائه", Group: GroupMetadata},
	}}
	tables := []Table{
		{
			Headers: []string{"expand", "كد درس", "نام درس", "نوع ارائه"},
			Rows: [][]string{
				{"", "2661144408", "مبانی", "\u062d\u0636\u0648\u0631\u064a"},
				{"", "6033144408", "هندسه", "\u062d\u0636\u0648\u0631\u064a"},
			},
		},
	}

	got, err := Build(cfg, tables, "test")
	if err != nil {
		t.Fatalf("Build() error = %v", err)
	}

	if got.Title != "test" {
		t.Errorf("Title = %q, want %q", got.Title, "test")
	}

	// Ordered data column first, unnumbered column keeps source order.
	if len(got.Columns) != 2 {
		t.Fatalf("got %d columns, want 2", len(got.Columns))
	}
	if got.Columns[0].Field != "code" || got.Columns[1].Field != "name" {
		t.Errorf("column order = %q, %q; want code, name",
			got.Columns[0].Field, got.Columns[1].Field)
	}

	if len(got.Rows) != 2 {
		t.Fatalf("got %d rows, want 2", len(got.Rows))
	}
	if got.Rows[0]["name"] != "مبانی" {
		t.Errorf("row name = %q, want %q", got.Rows[0]["name"], "مبانی")
	}
	if _, ok := got.Rows[0]["expand"]; ok {
		t.Error("skip column leaked into row fields")
	}

	// Arabic yeh (U+064A) in the source is normalized to Persian yeh (U+06CC).
	wantMeta := []KV{{Header: "نوع ارائه", Value: "\u062d\u0636\u0648\u0631\u06cc"}}
	if !reflect.DeepEqual(got.Metadata, wantMeta) {
		t.Errorf("metadata = %+v, want %+v", got.Metadata, wantMeta)
	}
}

func TestBuild_MissingColumn(t *testing.T) {
	cfg := Config{Columns: []Column{
		{Field: "code", Header: "كد درس", Group: GroupData},
	}}
	tables := []Table{{Headers: []string{"نام درس"}, Rows: [][]string{{"x"}}}}

	if _, err := Build(cfg, tables, ""); err == nil {
		t.Fatal("Build() succeeded, want error for missing column")
	}
}

func TestBuild_NoTables(t *testing.T) {
	if _, err := Build(Config{}, nil, ""); err == nil {
		t.Fatal("Build() succeeded, want error for no tables")
	}
}

func TestBuild_NormalizedHeaders(t *testing.T) {
	cfg := Config{Columns: []Column{
		{Field: "code", Header: "\u0643د درس", Group: GroupData}, // Arabic keheh
	}}
	tables := []Table{
		{Headers: []string{"\u0643د درس"}, Rows: [][]string{{"1"}}},
	}

	got, err := Build(cfg, tables, "")
	if err != nil {
		t.Fatalf("Build() error = %v", err)
	}
	if got.Columns[0].Header != "\u06a9د درس" { // normalized Persian kaf
		t.Errorf("header = %q, want normalized %q", got.Columns[0].Header, "\u06a9د درس")
	}
}
