package main

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

func LoadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading config: %w", err)
	}
	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parsing config %s: %w", path, err)
	}
	return &cfg, nil
}
