// Package config loads the column configuration from JSON.
package config

import (
	"encoding/json"
	"fmt"
	"os"
)

type Group string

const (
	GroupSkip     Group = "skip"     // ignored entirely
	GroupData     Group = "data"     // shown as a column in the results table
	GroupMetadata Group = "metadata" // constant across rows, shown once above the table
)

type Column struct {
	Field  string `json:"field"`
	Header string `json:"header"`
	Group  Group  `json:"group"`
	Order  int    `json:"order"`
}

type Config struct {
	Columns []Column `json:"columns"`
}

func LoadConfig(path string) (Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Config{}, fmt.Errorf("reading config: %w", err)
	}
	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return Config{}, fmt.Errorf("parsing config %s: %w", path, err)
	}
	for _, col := range cfg.Columns {
		if col.Field == "" {
			return Config{}, fmt.Errorf("config %s: column with empty field", path)
		}
		switch col.Group {
		case GroupSkip:
			// Skip columns are dropped entirely; they may be headerless
			// UI markers (e.g. expand icons) that never resolve to a cell.
		case GroupData, GroupMetadata:
			if col.Header == "" {
				return Config{}, fmt.Errorf("config %s: column %q: empty header", path, col.Field)
			}
		default:
			return Config{}, fmt.Errorf("config %s: column %q: invalid group %q (want %q, %q or %q)",
				path, col.Field, col.Group, GroupSkip, GroupData, GroupMetadata)
		}
	}
	return cfg, nil
}
