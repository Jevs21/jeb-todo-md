# jeb-todo-md

A minimal TUI application for editing a single markdown todo file.

## Tech Stack

- Go 1.24
- [Bubbletea](https://github.com/charmbracelet/bubbletea) v1 - TUI framework (Elm architecture)
- [Bubbles](https://github.com/charmbracelet/bubbles) - textinput component for inline editing
- [Lipgloss](https://github.com/charmbracelet/lipgloss) v1 - terminal styling

## Project Structure

```
cmd/jeb-todo-md/main.go    # Thin entry point (package main)
internal/tui/model.go       # Bubbletea model & TUI logic (package tui)
internal/tui/styles.go      # Lipgloss style constants (package tui)
internal/tui/todofile.go    # File I/O, parsing, data types (package tui)
tests/todofile_test.go      # Unit tests (package tests)
.github/workflows/test.yml  # CI: test on push
```

## Architecture

### Data Model

`TodoFile` keeps the entire file as `[]string` of raw lines plus `[]int` of indices pointing to todo lines. This preserves all non-todo content (comments, headings, blank lines) on round-trip. Mutations modify `RawLines` directly, then `Save()` writes back atomically (write .tmp, rename).

### TUI Modes

Four modes via a state machine in `model.Update()`:

- **ModeNormal** - Browse with j/k, trigger actions
- **ModeEditing** - Inline textinput replaces current todo text
- **ModeCreating** - Inline textinput for new todo below cursor
- **ModeRearrange** - j/k swaps items instead of just moving cursor

Delete uses a `pendingDelete` flag in Normal mode (press d twice to confirm).

### Markdown Format

```markdown
# Optional Title

- [ ] Unchecked todo
- [x] Checked todo
```

Parsed with regex `^\s*- \[([ xX])\] (.*)$`. First line checked for `# ` prefix for title. All other lines are preserved but ignored in the TUI.

## Commands

- `go build -o jeb-todo-md ./cmd/jeb-todo-md` - Build
- `go test ./...` - Run tests
- `JEB_TODO_FILE=/path/to/todo.md ./jeb-todo-md` - Run

## Keybindings

| Key | Mode | Action |
|-----|------|--------|
| j/k | Normal | Navigate up/down |
| space/x | Normal | Toggle checkbox |
| e | Normal | Edit current item |
| c | Normal | Create new item below cursor |
| r | Normal | Enter rearrange mode |
| d, d | Normal | Delete (press twice to confirm) |
| q | Normal | Quit |
| j/k | Rearrange | Swap item with neighbor |
| r/esc | Rearrange | Exit rearrange mode |
| enter | Edit/Create | Commit change |
| esc | Edit/Create | Cancel |
| ctrl+c | Any | Force quit |
