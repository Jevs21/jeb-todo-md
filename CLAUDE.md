# jeb-todo-md

A minimal TUI application for editing markdown todo files with linked file navigation.

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
CHANGELOG.md                # Feature history by date, grouped by release
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
- [ ] todo:work/tasks.md
- [ ] todo:/absolute/path/to/file.md
```

Parsed with regex `^\s*- \[([ xX])\] (.*)$`. First line checked for `# ` prefix for title. All other lines are preserved but ignored in the TUI.

Todo items with `todo:<filepath>` text are linked todos. `space`/`enter` navigates into the linked file; `x` still toggles the checkbox. Relative paths resolve from the current file's directory. Linked items render with blue underline styling.

## Commands

- `go build -o jeb-todo-md ./cmd/jeb-todo-md` - Build
- `go test ./...` - Run tests
- `./jeb-todo-md -f /path/to/todo.md` - Run with file flag
- `JEB_TODO_FILE=/path/to/todo.md ./jeb-todo-md` - Run with env var

## CLI Flags

| Flag | Description |
|------|-------------|
| `-f`, `--file` | Path to markdown todo file (overrides `JEB_TODO_FILE`) |
| `-v`, `--version` | Show version information |
| `--return` | Comma-separated file paths for back-navigation stack |
| `-h`, `--help` | Show help text |

**Precedence**: `-f`/`--file` flag > `JEB_TODO_FILE` environment variable. If neither is set, the program exits with an error.

## Keybindings

| Key | Mode | Action |
|-----|------|--------|
| j/k | Normal | Navigate up/down |
| space/enter | Normal | Toggle checkbox, or navigate into linked todo |
| x | Normal | Toggle checkbox (always toggles, even on linked items) |
| e | Normal | Edit current item |
| c | Normal | Create new item below cursor |
| r | Normal | Enter rearrange mode |
| d, d | Normal | Delete (press twice to confirm) |
| q/esc | Normal | Quit (or go back if navigated into a linked file) |
| j/k | Rearrange | Swap item with neighbor |
| r/esc | Rearrange | Exit rearrange mode |
| enter | Edit/Create | Commit change |
| esc | Edit/Create | Cancel |
| ctrl+c | Any | Force quit |

## Linked Files

Todo items can link to other todo files using the `todo:<filepath>` syntax. This enables organizing todos across multiple files (e.g., work vs personal).

- **Navigation**: Press `space` or `enter` on a linked item to open the linked file. Press `q` or `esc` to go back.
- **Toggle**: Press `x` to toggle a linked item's checkbox without navigating.
- **Path resolution**: Relative paths resolve from the current file's directory. Absolute paths are used as-is.
- **Stack-based**: Navigation uses an internal stack, so you can drill multiple levels deep and return to each previous file with cursor position preserved.
- **Max depth**: Navigation stack is capped at 50 levels.
- **Visual**: Linked items appear with blue underline styling. A breadcrumb showing the current filename appears when navigated into a linked file.
- **Errors**: If a linked file doesn't exist or links to itself, an error message appears inline and is cleared on the next keypress.

## Changelog

`CHANGELOG.md` must be kept up to date. When a PR adds, removes, or changes a feature, add a single-line entry at the top of the changelog in the format:

```
- YYYY-MM-DD - Description of change
```

When a new version is released (git tag), group all entries since the last release under a `## vX.Y.Z (YYYY-MM-DD)` heading.
