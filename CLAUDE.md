# jeb-todo-md

A minimal TUI application for editing a single markdown todo file.

## Tech Stack

- Go 1.24
- [Bubbletea](https://github.com/charmbracelet/bubbletea) v1 - TUI framework (Elm architecture)
- [Bubbles](https://github.com/charmbracelet/bubbles) - textinput component for inline editing
- [Lipgloss](https://github.com/charmbracelet/lipgloss) v1 - terminal styling

## Project Structure

All files are in the `main` package (no sub-packages).

- `main.go` - Entry point: reads `JEB_TODO_FILE` env var, loads file, starts TUI
- `model.go` - Bubbletea Model: struct, Init, Update (state machine), View
- `todofile.go` - File I/O: markdown parsing, atomic save, data types, all mutations
- `todofile_test.go` - Unit tests for parsing and mutations
- `styles.go` - Lipgloss style constants

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

- `go build -o jeb-todo-md .` - Build
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
