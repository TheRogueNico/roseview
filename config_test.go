package main

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestLoadConfig_Read(t *testing.T) {
	tests := []struct {
		name string // description of this test case
		// Named input parameters for target function.
		path    string
		want    *Config
		wantErr bool
	}{
		{
			name:    "missing file",
			path:    "testdata/does-not-exist",
			wantErr: true,
		},
		{
			name:    "empty path",
			path:    "",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, gotErr := LoadConfig(tt.path)
			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("LoadConfig() failed: %v", gotErr)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("LoadConfig() succeeded unexpectedly")
			}
			if got == nil {
				t.Errorf("LoadConfig() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestLoadConfig_Parse(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.json")

	config := `{
		"columns": [
			{
				"field": "name",
				"header": "Name",
				"group": "data",
				"order": 1
			},
			{
				"field": "id",
				"header": "ID",
				"group": "metadata",
				"order": 0
			},
			{
				"field": "debug",
				"header": "Debug",
				"group": "skip",
				"order": 99
			}
		]
	}`

	if err := os.WriteFile(path, []byte(config), 0o644); err != nil {
		t.Fatal(err)
	}

	got, err := LoadConfig(path)
	if err != nil {
		t.Fatalf("LoadConfig() error = %v", err)
	}

	want := &Config{
		Columns: []Column{
			{
				Field:  "name",
				Header: "Name",
				Group:  GroupData,
				Order:  1,
			},
			{
				Field:  "id",
				Header: "ID",
				Group:  GroupMetadata,
				Order:  0,
			},
			{
				Field:  "debug",
				Header: "Debug",
				Group:  GroupSkip,
				Order:  99,
			},
		},
	}

	if !reflect.DeepEqual(got, want) {
		t.Errorf("LoadConfig() = %+v, want %+v", got, want)
	}
}
