# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Build & Run

```bash
go build -o kimchi .    # Build binary
./kimchi                # Run in current directory
./kimchi /path/to/repo  # Run in specific directory
go build ./...          # Verify all packages compile
```

No tests currently exist.

## Architecture

Kimchi is a TUI app built with Go/Bubbletea following Elm architecture (Model-Update-View with message passing).

### Core Concepts

- **Windows** (`internal/window/`): Self-contained UI components (FileList, CommitList, DiffView, FileView, Help). Each implements the `Window` interface and knows nothing about layout.
- **Layout** (`internal/layout/`): Responsive slot-based system. Windows are assigned to named slots. Layout switches based on terminal width (ThreeSlot ≥80 cols, StackedThree <80).
- **State** (`internal/app/state.go`): Centralized state struct with `Selection` and `AppMode`. All updates happen via messages processed in `App.Update()`.
- **Mode** (`internal/app/mode.go`): Unified mode system with `AppMode` (working/branch/browse/docs) and `Selection` types.
- **Git Client** (`internal/git/`): Interface-based, shells out to git CLI. No go-git library.

### Message Flow

```
User Input → App.Update() → Global keys or delegate to focused window
    → Window returns tea.Cmd → Message sent → State updated → Re-render
```

### Key Files

- `internal/app/app.go` - Main tea.Model, orchestrates everything
- `internal/app/state.go` - State struct and transition methods
- `internal/app/mode.go` - AppMode, Selection, SelectionType definitions
- `internal/app/messages.go` - All message types
- `internal/config/config.go` - Centralized config: colors, styles, keybindings
- `internal/git/git.go` - Types, interfaces, enums (DiffMode, Status)
- `internal/git/client.go` - GitClient implementation
- `internal/window/filelist.go` - Tree view with directory structure
- `internal/window/diffview.go` - Diff view with PreviewContent types

### Unified Mode System

Press `m` to cycle through modes, or use number keys for direct access:
- `1` - changed:working (uncommitted changes)
- `2` - changed:branch (all changes vs base)
- `3` - browse (all files)
- `4` - docs (markdown only)

### Adding a New Window

1. Create `internal/window/newwindow.go` implementing `Window` interface
2. Add to window registry in `app.go` `New()` function
3. Assign to slot in `assignments` map
4. Handle any new message types in `App.Update()`

### Adding a Keybinding

1. Add binding to `KeyMap` struct in `internal/config/config.go`
2. Add to `DefaultKeyMap` with key definition
3. Update help.go for help display
4. Handle in `App.Update()` or delegate to window
