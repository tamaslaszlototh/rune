package main

import (
	"fmt"
	"os"

	"rune/internal/tui"
)

func main() {
	if len(os.Args) > 1 {
		fmt.Fprintf(os.Stderr, "Usage: rune\n")
		os.Exit(1)
	}
	if err := tui.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
