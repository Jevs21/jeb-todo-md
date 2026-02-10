# jeb-todo-md

[![Test](https://github.com/Jevs21/jeb-todo-md/actions/workflows/test.yml/badge.svg)](https://github.com/Jevs21/jeb-todo-md/actions/workflows/test.yml)

A minimal TUI for editing markdown todo files with linked file navigation.

## Features

- Navigate, toggle, create, edit, delete, and rearrange todos with vim-style keys
- Reads and writes standard markdown checkboxes (`- [ ]` / `- [x]`)
- Linked todo files with `todo:<filepath>` syntax for organizing across multiple files
- Stack-based navigation into linked files with breadcrumb header
- Dynamic header with file basename, date, and depth icons
- Preserves all non-todo content (headings, comments, blank lines) on save
- Atomic file writes (write to tmp, rename) to prevent data loss

## Usage

Pass a markdown file with the `-f` flag:

```bash
jeb-todo-md -f ~/todo.md
```

Or set the `JEB_TODO_FILE` environment variable:

```bash
export JEB_TODO_FILE="$HOME/todo.md"
jeb-todo-md
```

If both are provided, the `-f` flag takes precedence over the environment variable.

The file should use standard markdown checkbox syntax:

```markdown
- [ ] Unchecked item
- [x] Completed item
- [ ] todo:work/tasks.md
```

Any other content in the file (headings, blank lines, notes) is preserved as-is.

This works well with a synced Obsidian vault — point `-f` at a markdown file in your vault and edits stay in sync across devices:

```bash
jeb-todo-md -f ~/ObsidianVault/todo.md
```

## Linked Files

Todo items can link to other todo files using the `todo:<filepath>` syntax. This enables organizing todos across multiple files (e.g., work vs personal).

- **Navigation**: Press `space` or `enter` on a linked item to open the linked file. Press `q` or `esc` to go back.
- **Toggle**: Press `x` to toggle a linked item's checkbox without navigating.
- **Path resolution**: Relative paths resolve from the current file's directory. Absolute paths are used as-is.
- **Stack-based**: Navigation uses an internal stack, so you can drill multiple levels deep and return to each previous file with cursor position preserved.
- **Visual**: Linked items appear with blue underline styling. The header shows the current file's basename and navigation depth.

## Keybindings

| Key | Mode | Action |
|-----|------|--------|
| `j`/`k` | Normal | Navigate up/down |
| `space`/`enter` | Normal | Toggle checkbox, or navigate into linked todo |
| `x` | Normal | Toggle checkbox (always toggles, even on linked items) |
| `e` | Normal | Edit current item |
| `c` | Normal | Create new item below cursor |
| `r` | Normal | Enter rearrange mode |
| `d`, `d` | Normal | Delete (press twice to confirm) |
| `q`/`esc` | Normal | Quit (or go back if navigated into a linked file) |
| `j`/`k` | Rearrange | Swap item with neighbor |
| `r`/`esc` | Rearrange | Exit rearrange mode |
| `enter` | Edit/Create | Commit change |
| `esc` | Edit/Create | Cancel |
| `ctrl+c` | Any | Force quit |

## CLI Options

| Flag | Description |
|------|-------------|
| `-f`, `--file` | Path to markdown todo file (overrides `JEB_TODO_FILE`) |
| `--return` | Comma-separated file paths for back-navigation stack |
| `-v`, `--version` | Show version information |
| `-h`, `--help` | Show help text |

## Install

### Homebrew

```bash
brew install Jevs21/tap/jeb-todo-md
```

Then run directly with `-f`:

```bash
jeb-todo-md -f ~/todo.md
```

Or set up an alias:

```bash
echo 'alias todo="jeb-todo-md -f ~/todo.md"' >> ~/.bashrc
source ~/.bashrc
```

### From Source

```bash
git clone https://github.com/Jevs21/jeb-todo-md.git
cd jeb-todo-md
go build -o jeb-todo-md ./cmd/jeb-todo-md
```

Then run directly with `-f`:

```bash
./jeb-todo-md -f ~/todo.md
```

Or set up an alias:

```bash
echo 'alias todo="/path/to/jeb-todo-md -f ~/todo.md"' >> ~/.bashrc
source ~/.bashrc
```
