package tui

import (
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
)

const defaultFilePermission = 0644

var todoRegex = regexp.MustCompile(`^(\s*- \[)([ xX])(\] )(.*)$`)

// TodoItem represents a single parsed todo line.
type TodoItem struct {
	Text    string
	Checked bool
}

// IsLinkedTodo returns true if the todo text starts with "todo:" prefix.
func (item TodoItem) IsLinkedTodo() bool {
	return strings.HasPrefix(item.Text, "todo:")
}

// LinkedPath returns the file path portion of a linked todo, trimmed of whitespace.
// Returns empty string if the item is not a link or the path is empty.
func (item TodoItem) LinkedPath() string {
	if !item.IsLinkedTodo() {
		return ""
	}
	return strings.TrimSpace(strings.TrimPrefix(item.Text, "todo:"))
}

// ResolveLinkedPath resolves a linked file path relative to the current file's directory.
// If linkedPath is absolute, it is returned cleaned as-is.
// Otherwise, it is joined with the directory of currentFilePath and cleaned.
func ResolveLinkedPath(currentFilePath, linkedPath string) string {
	if filepath.IsAbs(linkedPath) {
		return filepath.Clean(linkedPath)
	}
	return filepath.Clean(filepath.Join(filepath.Dir(currentFilePath), linkedPath))
}

// TodoFile holds the entire file state for round-trip editing.
type TodoFile struct {
	Path        string
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
	tf.rebuildIndices()

	return tf, nil
}

// ParseTodoLine extracts a TodoItem from a raw line.
func ParseTodoLine(line string) *TodoItem {
	matches := todoRegex.FindStringSubmatch(line)
	if matches == nil {
		return nil
	}
	return &TodoItem{
		Text:    matches[4],
		Checked: matches[2] != " ",
	}
}

// FormatTodoLine creates a raw markdown line from a TodoItem.
func FormatTodoLine(item TodoItem) string {
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
	item := ParseTodoLine(line)
	if item == nil {
		return TodoItem{}
	}
	return *item
}

// SetTodoText updates the text of a todo at logical index.
func (tf *TodoFile) SetTodoText(todoIdx int, text string) {
	lineIdx := tf.TodoIndices[todoIdx]
	item := ParseTodoLine(tf.RawLines[lineIdx])
	if item == nil {
		return
	}
	item.Text = text
	tf.RawLines[lineIdx] = FormatTodoLine(*item)
}

// ToggleTodo flips the checked state of a todo.
func (tf *TodoFile) ToggleTodo(todoIdx int) {
	lineIdx := tf.TodoIndices[todoIdx]
	item := ParseTodoLine(tf.RawLines[lineIdx])
	if item == nil {
		return
	}
	item.Checked = !item.Checked
	tf.RawLines[lineIdx] = FormatTodoLine(*item)
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
	newLine := FormatTodoLine(item)

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
	if err := os.WriteFile(tmpPath, []byte(content), defaultFilePermission); err != nil {
		return err
	}
	if err := os.Rename(tmpPath, tf.Path); err != nil {
		os.Remove(tmpPath)
		return err
	}
	return nil
}
