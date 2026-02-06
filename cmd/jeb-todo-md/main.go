package main

import (
	"fmt"
	"os"

	"github.com/Jevs21/jeb-todo-md/internal/tui"
)

func main() {
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
