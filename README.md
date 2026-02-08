# jeb-todo-md

[![Test](https://github.com/Jevs21/jeb-todo-md/actions/workflows/test.yml/badge.svg)](https://github.com/Jevs21/jeb-todo-md/actions/workflows/test.yml)

A minimal TUI for editing a single markdown todo file.

## Features

- Navigate, toggle, create, edit, delete, and rearrange todos with vim-style keys
- Reads and writes standard markdown checkboxes (`- [ ]` / `- [x]`)
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
```

Any other content in the file (headings, blank lines, notes) is preserved as-is.

This works well with a synced Obsidian vault — point `-f` at a markdown file in your vault and edits stay in sync across devices:

```bash
jeb-todo-md -f ~/ObsidianVault/todo.md
```

## CLI Options

| Flag | Description |
|------|-------------|
| `-f`, `--file` | Path to markdown todo file (overrides `JEB_TODO_FILE`) |
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
