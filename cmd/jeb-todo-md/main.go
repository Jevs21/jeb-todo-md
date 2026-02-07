package main

import (
	"fmt"
	"os"

	"github.com/Jevs21/jeb-todo-md/internal/tui"
)

// Version information injected via ldflags at build time by GoReleaser.
var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func main() {
	if len(os.Args) > 1 && (os.Args[1] == "--version" || os.Args[1] == "-v") {
		fmt.Printf("jeb-todo-md %s (commit: %s, built: %s)\n", version, commit, date)
		os.Exit(0)
	}

	filePath := os.Getenv("JEB_TODO_FILE")
	if filePath == "" {
		fmt.Fprintln(os.Stderr, "Error: JEB_TODO_FILE environment variable not set")
		os.Exit(1)
	}

	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		fmt.Fprintf(os.Stderr, "Error: file not found: %s\n", filePath)
		os.Exit(1)
	}

	if err := tui.Run(filePath); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
