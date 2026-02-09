package tests

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Jevs21/jeb-todo-md/internal/tui"
)

const testMarkdown = `# Weekend Tasks

Some notes about this weekend.

- [x] Clean the kitchen
- [ ] Buy groceries
- [ ] Call dentist

## Later

- [ ] Fix the fence
`

func writeTempFile(t *testing.T, content string) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "test.md")
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestParseFile_TodoCount(t *testing.T) {
	path := writeTempFile(t, testMarkdown)
	tf, err := tui.ParseFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if tf.TodoCount() != 4 {
		t.Errorf("expected 4 todos, got %d", tf.TodoCount())
	}
}

func TestParseFile_TodoContent(t *testing.T) {
	path := writeTempFile(t, testMarkdown)
	tf, err := tui.ParseFile(path)
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		idx     int
		text    string
		checked bool
	}{
		{0, "Clean the kitchen", true},
		{1, "Buy groceries", false},
		{2, "Call dentist", false},
		{3, "Fix the fence", false},
	}

	for _, tt := range tests {
		item := tf.GetTodo(tt.idx)
		if item.Text != tt.text {
			t.Errorf("todo[%d]: expected text %q, got %q", tt.idx, tt.text, item.Text)
		}
		if item.Checked != tt.checked {
			t.Errorf("todo[%d]: expected checked=%v, got %v", tt.idx, tt.checked, item.Checked)
		}
	}
}

func TestToggleTodo(t *testing.T) {
	path := writeTempFile(t, testMarkdown)
	tf, _ := tui.ParseFile(path)

	// Toggle unchecked -> checked
	tf.ToggleTodo(1)
	item := tf.GetTodo(1)
	if !item.Checked {
		t.Error("expected todo 1 to be checked after toggle")
	}

	// Toggle checked -> unchecked
	tf.ToggleTodo(1)
	item = tf.GetTodo(1)
	if item.Checked {
		t.Error("expected todo 1 to be unchecked after second toggle")
	}
}

func TestSwapTodos(t *testing.T) {
	path := writeTempFile(t, testMarkdown)
	tf, _ := tui.ParseFile(path)

	tf.SwapTodos(1, 2)

	item1 := tf.GetTodo(1)
	item2 := tf.GetTodo(2)
	if item1.Text != "Call dentist" {
		t.Errorf("after swap, todo[1] expected 'Call dentist', got %q", item1.Text)
	}
	if item2.Text != "Buy groceries" {
		t.Errorf("after swap, todo[2] expected 'Buy groceries', got %q", item2.Text)
	}
}

func TestDeleteTodo(t *testing.T) {
	path := writeTempFile(t, testMarkdown)
	tf, _ := tui.ParseFile(path)

	tf.DeleteTodo(1) // Remove "Buy groceries"

	if tf.TodoCount() != 3 {
		t.Errorf("expected 3 todos after delete, got %d", tf.TodoCount())
	}
	item := tf.GetTodo(1)
	if item.Text != "Call dentist" {
		t.Errorf("after delete, todo[1] expected 'Call dentist', got %q", item.Text)
	}
}

func TestInsertTodo(t *testing.T) {
	path := writeTempFile(t, testMarkdown)
	tf, _ := tui.ParseFile(path)

	tf.InsertTodo(1, tui.TodoItem{Text: "New task", Checked: false})

	if tf.TodoCount() != 5 {
		t.Errorf("expected 5 todos after insert, got %d", tf.TodoCount())
	}
	item := tf.GetTodo(2)
	if item.Text != "New task" {
		t.Errorf("inserted todo expected 'New task', got %q", item.Text)
	}
}

func TestSetTodoText(t *testing.T) {
	path := writeTempFile(t, testMarkdown)
	tf, _ := tui.ParseFile(path)

	tf.SetTodoText(0, "Scrub the kitchen")
	item := tf.GetTodo(0)
	if item.Text != "Scrub the kitchen" {
		t.Errorf("expected 'Scrub the kitchen', got %q", item.Text)
	}
	// Should preserve checked state
	if !item.Checked {
		t.Error("expected checked state to be preserved")
	}
}

func TestRoundTrip(t *testing.T) {
	path := writeTempFile(t, testMarkdown)
	tf, _ := tui.ParseFile(path)
	if err := tf.Save(); err != nil {
		t.Fatal(err)
	}

	data, _ := os.ReadFile(path)
	if string(data) != testMarkdown {
		t.Errorf("round trip mismatch.\nExpected:\n%s\nGot:\n%s", testMarkdown, string(data))
	}
}

func TestInsertTodo_EmptyFile(t *testing.T) {
	path := writeTempFile(t, "# My List\n")
	tf, _ := tui.ParseFile(path)

	tf.InsertTodo(-1, tui.TodoItem{Text: "First task", Checked: false})

	if tf.TodoCount() != 1 {
		t.Errorf("expected 1 todo, got %d", tf.TodoCount())
	}
	item := tf.GetTodo(0)
	if item.Text != "First task" {
		t.Errorf("expected 'First task', got %q", item.Text)
	}
}

func TestParseTodoLine(t *testing.T) {
	tests := []struct {
		line    string
		isValid bool
		text    string
		checked bool
	}{
		{"- [ ] Buy milk", true, "Buy milk", false},
		{"- [x] Done thing", true, "Done thing", true},
		{"- [X] Also done", true, "Also done", true},
		{"  - [ ] Indented", true, "Indented", false},
		{"Not a todo", false, "", false},
		{"## Heading", false, "", false},
		{"- Regular list item", false, "", false},
	}

	for _, tt := range tests {
		item := tui.ParseTodoLine(tt.line)
		if tt.isValid {
			if item == nil {
				t.Errorf("expected %q to be a valid todo", tt.line)
				continue
			}
			if item.Text != tt.text {
				t.Errorf("line %q: expected text %q, got %q", tt.line, tt.text, item.Text)
			}
			if item.Checked != tt.checked {
				t.Errorf("line %q: expected checked=%v, got %v", tt.line, tt.checked, item.Checked)
			}
		} else if item != nil {
			t.Errorf("expected %q to NOT be a valid todo", tt.line)
		}
	}
}

func TestFormatTodoLine(t *testing.T) {
	item := tui.TodoItem{Text: "Buy milk", Checked: false}
	if got := tui.FormatTodoLine(item); got != "- [ ] Buy milk" {
		t.Errorf("expected '- [ ] Buy milk', got %q", got)
	}
	item.Checked = true
	if got := tui.FormatTodoLine(item); got != "- [x] Buy milk" {
		t.Errorf("expected '- [x] Buy milk', got %q", got)
	}
}

func TestIsLinkedTodo(t *testing.T) {
	tests := []struct {
		text     string
		expected bool
	}{
		{"todo:/path/to/file.md", true},
		{"todo:work.md", true},
		{"todo: spaced.md", true},
		{"todo:", true},
		{"Regular item", false},
		{"TODO:uppercase", false},
		{"not a todo:link", false},
		{"", false},
	}

	for _, tt := range tests {
		item := tui.TodoItem{Text: tt.text}
		if got := item.IsLinkedTodo(); got != tt.expected {
			t.Errorf("IsLinkedTodo(%q) = %v, want %v", tt.text, got, tt.expected)
		}
	}
}

func TestLinkedPath(t *testing.T) {
	tests := []struct {
		text     string
		expected string
	}{
		{"todo:/path/to/file.md", "/path/to/file.md"},
		{"todo:work.md", "work.md"},
		{"todo: spaced.md", "spaced.md"},
		{"todo:  extra-spaces.md  ", "extra-spaces.md"},
		{"todo:", ""},
		{"todo:   ", ""},
		{"Regular item", ""},
		{"", ""},
	}

	for _, tt := range tests {
		item := tui.TodoItem{Text: tt.text}
		if got := item.LinkedPath(); got != tt.expected {
			t.Errorf("LinkedPath(%q) = %q, want %q", tt.text, got, tt.expected)
		}
	}
}

func TestResolveLinkedPath(t *testing.T) {
	tests := []struct {
		currentFile string
		linkedPath  string
		expected    string
	}{
		// Absolute paths pass through cleaned
		{"/home/user/todos/main.md", "/absolute/path.md", "/absolute/path.md"},
		{"/home/user/todos/main.md", "/absolute/../clean.md", "/clean.md"},
		// Relative paths join with current file's directory
		{"/home/user/todos/main.md", "work.md", "/home/user/todos/work.md"},
		{"/home/user/todos/main.md", "sub/nested.md", "/home/user/todos/sub/nested.md"},
		{"/home/user/todos/main.md", "../parent.md", "/home/user/parent.md"},
		{"/home/user/todos/main.md", "./same-dir.md", "/home/user/todos/same-dir.md"},
	}

	for _, tt := range tests {
		got := tui.ResolveLinkedPath(tt.currentFile, tt.linkedPath)
		if got != tt.expected {
			t.Errorf("ResolveLinkedPath(%q, %q) = %q, want %q", tt.currentFile, tt.linkedPath, got, tt.expected)
		}
	}
}

func TestLinkedTodoRoundTrip(t *testing.T) {
	content := `# Project Tasks

- [ ] todo:work.md
- [x] todo:/absolute/done.md
- [ ] Regular task
`
	path := writeTempFile(t, content)
	todoFile, err := tui.ParseFile(path)
	if err != nil {
		t.Fatal(err)
	}

	if todoFile.TodoCount() != 3 {
		t.Fatalf("expected 3 todos, got %d", todoFile.TodoCount())
	}

	// Verify linked items are parsed correctly
	firstItem := todoFile.GetTodo(0)
	if !firstItem.IsLinkedTodo() {
		t.Error("expected todo[0] to be a linked todo")
	}
	if firstItem.LinkedPath() != "work.md" {
		t.Errorf("expected linked path 'work.md', got %q", firstItem.LinkedPath())
	}

	secondItem := todoFile.GetTodo(1)
	if !secondItem.IsLinkedTodo() {
		t.Error("expected todo[1] to be a linked todo")
	}
	if !secondItem.Checked {
		t.Error("expected todo[1] to be checked")
	}

	thirdItem := todoFile.GetTodo(2)
	if thirdItem.IsLinkedTodo() {
		t.Error("expected todo[2] to NOT be a linked todo")
	}

	// Save and verify round-trip
	if err := todoFile.Save(); err != nil {
		t.Fatal(err)
	}

	savedData, _ := os.ReadFile(path)
	if string(savedData) != content {
		t.Errorf("round trip mismatch.\nExpected:\n%s\nGot:\n%s", content, string(savedData))
	}
}

func TestSave_PreservesNonTodoLines(t *testing.T) {
	path := writeTempFile(t, testMarkdown)
	tf, _ := tui.ParseFile(path)

	// Modify a todo
	tf.ToggleTodo(1)
	tf.Save()

	data, _ := os.ReadFile(path)
	content := string(data)

	// Non-todo lines should still be there
	if !strings.Contains(content, "Some notes about this weekend.") {
		t.Error("non-todo content was lost")
	}
	if !strings.Contains(content, "## Later") {
		t.Error("sub-heading was lost")
	}
}
