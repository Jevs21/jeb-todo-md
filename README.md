# jeb-todo-md

[![Test](https://github.com/Jevs21/jeb-todo-md/actions/workflows/test.yml/badge.svg)](https://github.com/Jevs21/jeb-todo-md/actions/workflows/test.yml)

A minimal TUI for editing a single markdown todo file.

## Usage

| Command | Description |
|---------|-------------|
| `make build` | Compile the binary |
| `make test` | Run all tests |
| `make run` | Build and launch (requires `JEB_TODO_FILE`) |
| `make clean` | Remove the binary |

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
