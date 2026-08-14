// Command roseview converts an HTML course-schedule export into a
// static, searchable website.
package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/TheRogueNico/roseview/internal/build"
	"github.com/TheRogueNico/roseview/internal/config"
	"github.com/TheRogueNico/roseview/internal/parse"
	"github.com/TheRogueNico/roseview/internal/render"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "roseview:", err)
		os.Exit(1)
	}
}

func run() error {
	var (
		input  = flag.String("input", "", "path to the HTML course-schedule export (required)")
		outDir = flag.String("out", "out", "directory to write the generated site into")
		cfg    = flag.String("config", "config.json", "path to the column config JSON")
		title  = flag.String("title", "برنامه درسی", "page title shown in the site header")
	)
	flag.Parse()

	if *input == "" {
		return fmt.Errorf("missing required -input flag")
	}

	conf, err := config.LoadConfig(*cfg)
	if err != nil {
		return err
	}

	f, err := os.Open(*input)
	if err != nil {
		return fmt.Errorf("opening input: %w", err)
	}
	defer f.Close()

	tables, err := parse.ParseTables(f)
	if err != nil {
		return err
	}

	out, err := build.Build(conf, tables, *title)
	if err != nil {
		return err
	}

	if err := render.Render(*outDir, out); err != nil {
		return err
	}

	fmt.Printf("wrote %d rows to %s\n", len(out.Rows), *outDir)
	return nil
}
