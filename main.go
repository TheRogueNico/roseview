// Command roseview converts an HTML course-schedule export into a
// static, searchable website.
package main

import (
	"fmt"
	"os"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "roseview:", err)
		os.Exit(1)
	}
}

func run() error {
	return nil
}
