# Kimchi Design Document

A terminal user interface for code review and repository browsing, built with Rust and Ratatui.

## Vision

When AI writes code, you don't need a traditional IDE. You need:
- **Visibility** - see what the AI is changing
- **Navigation** - understand the codebase
- **Review** - approve changes with confidence
- **Context** - specs and docs alongside code

This is read-heavy, not write-heavy. The human reviews, the AI writes.

## Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                         main.rs                              │
│              Terminal setup, event loop                      │
└─────────────────────────────────────────────────────────────┘
                              │
              ┌───────────────┼───────────────┐
              ▼               ▼               ▼
┌─────────────────┐  ┌─────────────┐  ┌─────────────────┐
│   EventHandler  │  │     App     │  │    Terminal     │
│   (event.rs)    │  │  (app.rs)   │  │   (ratatui)     │
│                 │  │             │  │                 │
│ • Keyboard      │  │ • State     │  │ • Raw mode      │
│ • File watcher  │  │ • Logic     │  │ • Rendering     │
│ • Tick events   │  │ • Commands  │  │                 │
└─────────────────┘  └─────────────┘  └─────────────────┘
                              │
         ┌────────────────────┼────────────────────┐
         ▼                    ▼                    ▼
┌─────────────────┐  ┌─────────────────┐  ┌─────────────────┐
│   GitClient     │  │  GitHubClient   │  │   UI Widgets    │
│  (git2 crate)   │  │   (gh CLI)      │  │                 │
│                 │  │                 │  │ • FileList      │
│ • Status        │  │ • PR info       │  │ • CommitList    │
│ • Diff          │  │ • Comments      │  │ • DiffView      │
│ • Log           │  │ • Reviews       │  │ • HelpModal     │
└─────────────────┘  └─────────────────┘  └─────────────────┘
```

## Event Flow

```
User Input → EventHandler → App.handle_key() → State Update → render()
     ▲                                              │
     │              Widget State Updates ◄──────────┘
     │                      │
     └──────────────────────┘
```

1. EventHandler runs in separate thread, sends events via mpsc channel
2. App receives events, updates state, delegates to focused widget
3. On each frame, App renders all widgets with current state
4. Commands (like open editor) are queued and executed after render

## Features

### Implemented
- [x] Tree view of changed files with directories
- [x] Side-by-side diff viewer with line numbers
- [x] Vim-style navigation (j/k, J/K for fast)
- [x] Auto-refresh on git changes (watches .git/index)
- [x] Syntax highlighting for diffs (+ green, - red)
- [x] Status bar (branch, mode, file count, diff stats)
- [x] Yank path (y to copy file path to clipboard)
- [x] Open in editor (o to open in $EDITOR with line number)
- [x] Help modal with keybindings
- [x] Unified mode system with 4 modes
- [x] File content viewer for browse mode
- [x] PR comments - inline comments in diff view
- [x] Folder selection - combined diff for directories
- [x] PR summary view with commit details
- [x] Line selection in diff view with cursor
- [x] Commit list window (last 8 commits)
- [x] Collapsible folders with status indicators

### Future
- [ ] Markdown rendering
- [ ] Hooks for AI agent integration

## Modes

| Key | Mode | Description |
|-----|------|-------------|
| 1 | ChangedWorking | Uncommitted changes (`git diff`) |
| 2 | ChangedBranch | All changes vs base branch (`git diff <base>`) - default |
| 3 | Browse | All tracked files in repository |
| 4 | Docs | Markdown files only |

Press `m` to cycle through modes.

## Layout

### Wide Layout (≥80 columns)

```
┌─────────────────┬──────────────────────────────────┐
│                 │                                  │
│    FileList     │            DiffView              │
│                 │                                  │
├─────────────────┤         (preview panel)          │
│                 │                                  │
│   CommitList    │                                  │
│                 │                                  │
└─────────────────┴──────────────────────────────────┘
│                    Status Bar                      │
└────────────────────────────────────────────────────┘
```

### Narrow Layout (<80 columns)

```
┌────────────────────┐
│     FileList       │
├────────────────────┤
│    CommitList      │
├────────────────────┤
│     DiffView       │
├────────────────────┤
│    Status Bar      │
└────────────────────┘
```

## Widgets

### FileList

Tree view of files with directory nesting.

```
Files (4)
▼ src/
  > main.rs           M
    app.rs            M
▼ internal/
  ▼ git/
      client.rs       A
  README.md           M
```

- Directories shown with ▼/▶ prefix (expanded/collapsed)
- h collapses folder, l expands it
- Status indicators: M (modified), A (added), D (deleted), R (renamed)
- C marker for files with PR comments
- Color-coded by git status

### CommitList

Recent commits (default: 8).

```
Commits (8)
> abc1234 Add new feature for handling...
  def5678 Fix bug in authentication
```

- When focused, preview shows commit summary with PR info

### DiffView

Side-by-side diff viewer with inline PR comments.

**Content Types:**
- `FileDiff` - Unified diff for single file
- `FolderDiff` - Combined diff for directory
- `FileContent` - Raw file content (browse mode, single column)
- `CommitSummary` - Commit metadata with PR reviews

**Display Format:**
```
  12 │ context line            │   12 │ context line
  13 │-removed line            │      │
     │                         │   13 │+added line
  14 │ context line            │   14 │ context line
```

**Inline Comments:**
```
  37 │ let result = process(); │   37 │ let result = process();
     │ 💬 kamilmac
     │    cool!
  38 │ return result;          │   38 │ return result;
```

### HelpModal

Modal overlay showing all keybindings, toggled with `?`.

## Key Bindings

### Navigation

| Key | Action |
|-----|--------|
| j / ↓ | Move down |
| k / ↑ | Move up |
| J / K | Fast move (5 lines) |
| h | Collapse folder |
| l | Expand folder |
| Tab | Next window |
| Shift+Tab | Previous window |
| Ctrl+d | Page down |
| Ctrl+u | Page up |
| g | Go to top |
| G | Go to bottom |

### Actions

| Key | Action |
|-----|--------|
| y | Yank path to clipboard (with line number in diff) |
| o | Open in $EDITOR |
| r | Refresh |
| ? | Toggle help |
| q / Ctrl+C | Quit |

## Data Structures

### Core Types

```rust
pub enum FileStatus {
    Modified, Added, Deleted, Renamed, Untracked, Unchanged
}

pub enum DiffMode {
    Working,  // git diff
    Branch,   // git diff <base>
}

pub enum AppMode {
    ChangedWorking,  // Mode 1
    ChangedBranch,   // Mode 2
    Browse,          // Mode 3
    Docs,            // Mode 4
}

pub struct StatusEntry {
    pub path: String,
    pub status: FileStatus,
}

pub struct Commit {
    pub hash: String,
    pub short_hash: String,
    pub author: String,
    pub date: String,
    pub subject: String,
}
```

### GitHub Types

```rust
pub struct PrInfo {
    pub number: u64,
    pub title: String,
    pub body: String,
    pub author: String,
    pub state: String,
    pub url: String,
    pub reviews: Vec<Review>,
    pub comments: Vec<Comment>,
    pub file_comments: HashMap<String, Vec<Comment>>,
}

pub struct Comment {
    pub author: String,
    pub body: String,
    pub path: Option<String>,
    pub line: Option<u32>,
}
```

## Git Integration

Uses libgit2 (git2 crate) for native performance:

- Repository opening with path resolution
- Status checking via index/workdir comparison
- Diff generation between commits/trees
- Commit history traversal
- Base branch auto-detection (main, master, origin/*)
- File content reading from HEAD tree

## GitHub Integration

Uses gh CLI for GitHub API access:

- PR detection for current branch
- Review fetching with approval state
- Inline comment fetching mapped to file paths and lines
- Polling every 60 seconds for updates

## File Watching

Watches `.git/index` for changes using notify crate:

- Debounced at 500ms to avoid excessive refreshes
- Triggers FileChanged event on git operations
- Auto-refreshes file list and diff

## Configuration

Hardcoded with Catppuccin Mocha color scheme:

```rust
Colors {
    added: Rgb(166, 227, 161),      // Green
    removed: Rgb(243, 139, 168),    // Red
    modified: Rgb(250, 179, 135),   // Peach
    renamed: Rgb(203, 166, 247),    // Mauve
    header: Rgb(137, 180, 250),     // Blue
    text: Rgb(205, 214, 244),       // Text
    comment: Rgb(249, 226, 175),    // Yellow
    border: Rgb(69, 71, 90),        // Surface1
    border_focused: Rgb(137, 180, 250), // Blue
}
```

Layout settings:
- Left panel: 30%
- Right panel: 70%
- Responsive breakpoint: 80 columns
- Max commits: 8
- File watcher debounce: 500ms
- PR poll interval: 60 seconds

## External Editor Support

Opens files in `$EDITOR` with line number support:

- **vim/nvim**: `+{line}` argument
- **helix**: `{file}:{line}` format
- Terminal suspended during editor session
- Event polling paused to prevent interference
- Auto-refresh on editor close

## Project Structure

```
src/
├── main.rs           # Entry point, terminal setup, event loop
├── app.rs            # Main application state and logic
├── event.rs          # Event handling (keyboard, file watching, ticks)
├── config.rs         # Configuration, colors, layout settings
├── git/
│   ├── mod.rs        # Git module exports
│   ├── types.rs      # Git data structures
│   └── client.rs     # Git operations using libgit2
├── github/
│   └── mod.rs        # GitHub API client using gh CLI
└── ui/
    ├── mod.rs        # UI module exports
    ├── layout.rs     # Layout computation (responsive grid)
    └── widgets/
        ├── mod.rs
        ├── file_list.rs    # Tree view widget
        ├── commit_list.rs  # Commit history widget
        ├── diff_view.rs    # Diff/content preview widget
        └── help.rs         # Help modal widget
```

## Dependencies

```toml
ratatui = "0.29"           # TUI framework
crossterm = "0.28"         # Terminal handling
git2 = "0.19"              # Native git operations
notify = "7"               # File watching
notify-debouncer-mini      # Debounced events
clap = "4"                 # CLI parsing
arboard = "3"              # Clipboard
serde/serde_json           # JSON parsing
anyhow/thiserror           # Error handling
chrono = "0.4"             # Date formatting
unicode-width = "0.2"      # Text width calculation
```

## Performance

- Native libgit2 (no shell overhead for git)
- Debounced file watching (500ms)
- Lazy PR polling (60s intervals)
- Offset-based viewport rendering
- Release build: LTO, single codegen unit, stripped binary

## Error Handling

- Uses `anyhow::Result<T>` throughout
- Context wrapping for helpful error messages
- Graceful fallbacks:
  - Missing gh CLI: PR features disabled
  - Missing base branch: falls back to working status
  - Binary files: shows "Binary file" message
  - Unreadable files: returns empty/default
