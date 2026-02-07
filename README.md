# jeb-todo-md

[![Test](https://github.com/Jevs21/jeb-todo-md/actions/workflows/test.yml/badge.svg)](https://github.com/Jevs21/jeb-todo-md/actions/workflows/test.yml)

A minimal TUI for editing a single markdown todo file.

## Features

- Navigate, toggle, create, edit, delete, and rearrange todos with vim-style keys
- Reads and writes standard markdown checkboxes (`- [ ]` / `- [x]`)
- Preserves all non-todo content (headings, comments, blank lines) on save
- Atomic file writes (write to tmp, rename) to prevent data loss

## Usage

Point `JEB_TODO_FILE` at any markdown file containing todo items and run the binary:

```bash
export JEB_TODO_FILE="$HOME/todo.md"
./jeb-todo-md
```

The file should use standard markdown checkbox syntax:

```markdown
- [ ] Unchecked item
- [x] Completed item
```

Any other content in the file (headings, blank lines, notes) is preserved as-is.

This works well with a synced Obsidian vault — point `JEB_TODO_FILE` at a markdown file in your vault and edits stay in sync across devices.

## Install

```bash
# clone and build
git clone https://github.com/Jevs21/jeb-todo-md.git
cd jeb-todo-md
make build

# set your todo file and add an alias
echo 'export JEB_TODO_FILE="$HOME/todo.md"' >> ~/.bashrc
echo 'alias todo="$HOME/path/to/jeb-todo-md"' >> ~/.bashrc
source ~/.bashrc
```
