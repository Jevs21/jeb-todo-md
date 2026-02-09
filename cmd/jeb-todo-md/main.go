package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/Jevs21/jeb-todo-md/internal/tui"
)

// Version information injected via ldflags at build time by GoReleaser.
var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func main() {
	var filePath string
	var showVersion bool
	var returnPaths string

	flag.StringVar(&filePath, "file", "", "Path to markdown todo file (overrides JEB_TODO_FILE)")
	flag.StringVar(&filePath, "f", "", "Path to markdown todo file (shorthand)")
	flag.BoolVar(&showVersion, "version", false, "Show version information")
	flag.BoolVar(&showVersion, "v", false, "Show version information (shorthand)")
	flag.StringVar(&returnPaths, "return", "", "Comma-separated file paths for back-navigation stack")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [OPTIONS]\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "A minimal TUI for editing markdown todo files.\n\n")
		fmt.Fprintf(os.Stderr, "Options:\n")
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nIf -f/--file is not provided, reads from JEB_TODO_FILE environment variable.\n")
	}

	flag.Parse()

	if showVersion {
		fmt.Printf("jeb-todo-md %s (commit: %s, built: %s)\n", version, commit, date)
		os.Exit(0)
	}

	// Precedence: --file flag > JEB_TODO_FILE env var
	if filePath == "" {
		filePath = os.Getenv("JEB_TODO_FILE")
	}

	if filePath == "" {
		fmt.Fprintln(os.Stderr, "Error: no todo file specified")
		fmt.Fprintln(os.Stderr, "  Set JEB_TODO_FILE environment variable, or use -f/--file flag")
		os.Exit(1)
	}

	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		fmt.Fprintf(os.Stderr, "Error: file not found: %s\n", filePath)
		os.Exit(1)
	}

	// Parse return stack paths
	var returnStack []string
	if returnPaths != "" {
		for _, returnPath := range strings.Split(returnPaths, ",") {
			trimmedPath := strings.TrimSpace(returnPath)
			if trimmedPath == "" {
				continue
			}
			if _, err := os.Stat(trimmedPath); os.IsNotExist(err) {
				fmt.Fprintf(os.Stderr, "Warning: return path not found: %s\n", trimmedPath)
				continue
			}
			returnStack = append(returnStack, trimmedPath)
		}
	}

	if err := tui.Run(filePath, returnStack); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
