package main

import (
	"os"
	"regexp"
	"slices"
	"strings"
)

var todoRegex = regexp.MustCompile(`^(\s*- \[)([ xX])(\] )(.*)$`)

// TodoItem represents a single parsed todo line.
type TodoItem struct {
	Text    string
	Checked bool
}

// TodoFile holds the entire file state for round-trip editing.
type TodoFile struct {
	Path        string
	Title       string
	RawLines    []string
	TodoIndices []int
}

// ParseFile reads the file at path and returns a TodoFile.
func ParseFile(path string) (*TodoFile, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	content := string(data)
	// Preserve trailing newline behavior
	lines := strings.Split(content, "\n")

	tf := &TodoFile{Path: path, RawLines: lines}

	// Check first non-empty line for h1
	if len(lines) > 0 {
		trimmed := strings.TrimSpace(lines[0])
		if strings.HasPrefix(trimmed, "# ") {
			tf.Title = strings.TrimPrefix(trimmed, "# ")
		}
	}

	tf.rebuildIndices()

	return tf, nil
}

// parseTodoLine extracts a TodoItem from a raw line.
func parseTodoLine(line string) *TodoItem {
	matches := todoRegex.FindStringSubmatch(line)
	if matches == nil {
		return nil
	}
	return &TodoItem{
		Text:    matches[4],
		Checked: matches[2] != " ",
	}
}

// formatTodoLine creates a raw markdown line from a TodoItem.
func formatTodoLine(item TodoItem) string {
	check := " "
	if item.Checked {
		check = "x"
	}
	return "- [" + check + "] " + item.Text
}

// rebuildIndices rescans RawLines to rebuild TodoIndices.
func (tf *TodoFile) rebuildIndices() {
	tf.TodoIndices = nil
	for i, line := range tf.RawLines {
		if todoRegex.MatchString(line) {
			tf.TodoIndices = append(tf.TodoIndices, i)
		}
	}
}

// TodoCount returns the number of todos.
func (tf *TodoFile) TodoCount() int {
	return len(tf.TodoIndices)
}

// GetTodo returns the TodoItem at logical index.
func (tf *TodoFile) GetTodo(todoIdx int) TodoItem {
	line := tf.RawLines[tf.TodoIndices[todoIdx]]
	item := parseTodoLine(line)
	if item == nil {
		return TodoItem{}
	}
	return *item
}

// SetTodoText updates the text of a todo at logical index.
func (tf *TodoFile) SetTodoText(todoIdx int, text string) {
	lineIdx := tf.TodoIndices[todoIdx]
	item := parseTodoLine(tf.RawLines[lineIdx])
	if item == nil {
		return
	}
	item.Text = text
	tf.RawLines[lineIdx] = formatTodoLine(*item)
}

// ToggleTodo flips the checked state of a todo.
func (tf *TodoFile) ToggleTodo(todoIdx int) {
	lineIdx := tf.TodoIndices[todoIdx]
	item := parseTodoLine(tf.RawLines[lineIdx])
	if item == nil {
		return
	}
	item.Checked = !item.Checked
	tf.RawLines[lineIdx] = formatTodoLine(*item)
}

// SwapTodos swaps two todo lines in RawLines by content.
func (tf *TodoFile) SwapTodos(a, b int) {
	lineA := tf.TodoIndices[a]
	lineB := tf.TodoIndices[b]
	tf.RawLines[lineA], tf.RawLines[lineB] = tf.RawLines[lineB], tf.RawLines[lineA]
}

// DeleteTodo removes a todo line and rebuilds indices.
func (tf *TodoFile) DeleteTodo(todoIdx int) {
	lineIdx := tf.TodoIndices[todoIdx]

	tf.RawLines = append(tf.RawLines[:lineIdx], tf.RawLines[lineIdx+1:]...)
	tf.rebuildIndices()
}

// InsertTodo inserts a new todo after the given logical index.
// If todoIdx is -1 or there are no todos, appends at end of file.
func (tf *TodoFile) InsertTodo(afterTodoIdx int, item TodoItem) {
	newLine := formatTodoLine(item)

	var insertAt int
	if tf.TodoCount() == 0 || afterTodoIdx < 0 {
		// Append before the last empty line (if file ends with newline)
		insertAt = len(tf.RawLines)
		if insertAt > 0 && tf.RawLines[insertAt-1] == "" {
			insertAt = insertAt - 1
		}
	} else {
		insertAt = tf.TodoIndices[afterTodoIdx] + 1
	}

	tf.RawLines = slices.Insert(tf.RawLines, insertAt, newLine)
	tf.rebuildIndices()
}

// Save writes RawLines back to the file atomically.
func (tf *TodoFile) Save() error {
	content := strings.Join(tf.RawLines, "\n")
	tmpPath := tf.Path + ".tmp"
	if err := os.WriteFile(tmpPath, []byte(content), 0644); err != nil {
		return err
	}
	return os.Rename(tmpPath, tf.Path)
}
