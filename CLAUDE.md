# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## What This Is

`ai_rename` is a Neovim plugin for AI-assisted variable renaming in Go files. It combines:

- A **Lua layer** (`rename.lua`) that integrates with Neovim via Tree-sitter and `vim.ui`
- A **Go CLI binary** (`go/ai_rename_bin`) that analyzes Go source files and queries a local LLM (Ollama) for rename suggestions

## Build

```bash
# Build the Go binary (run from go/ directory)
cd go && go build -o ai_rename_bin ./cmd/
```

The binary path is hardcoded in `rename.lua` as `~/.config/nvim/lua/ai_rename/go/ai_rename_bin`.

## Runtime Dependency

Requires [Ollama](https://ollama.com) running locally with the `llama3:8b` model:

```bash
ollama pull llama3:8b
ollama serve
```

## Architecture

### Data Flow

1. **Neovim trigger** → `rename.lua:suggest_and_rename()`
2. **Tree-sitter** resolves the identifier under cursor and its enclosing function name
3. **CLI invocation**: `ai_rename_bin <filepath> <func:var>` (e.g., `file.go Fibonacci:result`)
4. **Go binary** (`go/cmd/main.go` or via `internal/rename/run.go`):
   - Parses the Go file with `go/ast` and `go/parser`
   - Builds a `VarContext` (type, usages, assignments, parameters, imports, file comments)
   - Sends a structured prompt to Ollama via `ollama run llama3:8b`
   - Returns JSON with `suggestions` (name+reason pairs)
5. **Neovim** presents suggestions via `vim.ui.select`, then applies chosen edit bottom-up

### Go Package Layout (`go/internal/rename/`)

| File         | Responsibility                                                                                           |
| ------------ | -------------------------------------------------------------------------------------------------------- |
| `context.go`       | `VarContext` struct; `BuildVarContext()` — parses file and populates all context fields                  |
| `field_context.go` | `FieldContext` struct; `BuildFieldContext()` — context for struct field renames                          |
| `resolve.go`       | `Selector` struct; `ResolveSelector()` — finds an `*ast.Ident` by `funcvar` name or `position` (row:col) |
| `prompt.go`        | `BuildPrompt()` / `BuildTypePrompt()` / `BuildFieldPrompt()` — LLM prompt builders                      |
| `llm.go`           | `CallLLM()` — dispatches to Claude or Ollama; retries up to 3× for exactly 3 output lines               |
| `result.go`        | Shared types: `Suggestion`, `Debug`, `Result`                                                            |
| `run.go`           | `Run()` — top-level orchestrator: context → prompt → LLM → parse suggestions                            |

### Known Inconsistency

`cmd/main.go` accepts `<file.go> <row:col>` positional format but `rename.lua` passes `<func:var>` (e.g., `Fibonacci:result`). The `internal/rename` package supports both via `Selector.Kind` (`"funcvar"` vs `"position"`), but `cmd/main.go` only implements the position path. The working entry point for the Lua plugin is through `Run()` in `run.go`, which uses `BuildVarContext()` with the `funcvar` selector.

### Plugin Entry Point

`init.lua` registers the `:AIRename` user command and is the standard entry point when the plugin is loaded via `require("ai_rename")`. The active rename logic lives in `rename.lua`.

## LLM Output Format

The LLM is prompted to return exactly 3 lines, each formatted as:

```
<name> - <very short justification (max 5 words)>
```

`run.go` splits on ` - `, matching the prompt format.
