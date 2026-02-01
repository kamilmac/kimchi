# Kimchi - AI-Native IDE

A read-first terminal IDE for AI-driven development workflows.

## Vision

When AI writes code, you don't need a traditional IDE. You need:
- **Visibility** - see what the AI is changing
- **Navigation** - understand the codebase
- **Review** - approve changes with confidence
- **Context** - specs and docs alongside code

This is read-heavy, not write-heavy. The human reviews, the AI writes.

## Problem

Current IDEs are built for humans writing code. In an AI workflow:
- You constantly run `git diff` to see changes
- You switch between terminal and editor
- You lose track of what changed vs your spec
- You trust blindly and review later

## Solution

A lightweight TUI that serves as your primary interface when working with AI:
- **Changed files** - what did the AI touch?
- **Diff view** - what exactly changed?
- **Commit history** - recent commits on the branch
- **PR summary** - overview of the pull request
- **All files** - browse the entire codebase
- **Docs** - your specs alongside the implementation

## Features

### Implemented
- [x] Display list of changed files (git status)
- [x] Tree view for file list with directories
- [x] Show side-by-side diff for selected file
- [x] Vim-style navigation (`j`/`k`)
- [x] Fast navigation (`J`/`K` - 5 lines at a time)
- [x] Auto-refresh on file changes (500ms debounce via fsnotify)
- [x] Syntax highlighting for diffs (+ green, - red)
- [x] Status bar (branch, mode, file count, diff stats)
- [x] Yank path (`y` to copy file path to clipboard)
- [x] Open in editor (`o` to open in $EDITOR)
- [x] Help modal with keybindings
- [x] Unified mode system with 4 modes
- [x] File content viewer for unchanged files
- [x] PR comments - inline comments in diff view, "C" indicator on files with comments
- [x] Folder selection - select directories to view combined diff
- [x] PR summary view - when commit is selected, shows commit details + PR summary
- [x] Line selection in diff view - navigate to specific lines with cursor
- [x] Commit list window - shows last 8 commits
- [x] Collapsible folders with aggregate status indicators

### Future
- [ ] FileExplorer - full project tree navigation with expand/collapse
- [ ] Markdown rendering - render markdown with formatting
- [ ] Hooks/events for integration with AI agents (Claude Code, Cursor, Aider, etc.)

## Architecture

### Terminology

| Term | Definition |
|------|------------|
| **Window** | A renderable component with its own state. Knows nothing about where it's placed. Given width/height, renders content. |
| **Slot** | A named rectangular area in a layout. Has position and dimensions. Holds one window. |
| **Layout** | Defines slot structure and arrangement. Knows nothing about window types. Handles responsive resizing. |
| **Modal** | A presentation mode. Any window can be shown as a modal (floating, overlays layout, captures input). |

### Windows

Windows implement a common interface. Layout doesn't care what type they are.

| Window | Description |
|--------|-------------|
| `FileList` | Tree view of changed files with status indicators |
| `CommitList` | List of recent commits on the branch |
| `DiffView` | Diff preview for selected file (or folder diff, commit/PR summary) |
| `FileView` | File content viewer for browse mode |
| `Help` | Keybinding reference (modal) |

Each window:
- Has its own state (cursor position, scroll offset, etc.)
- Has a header with relevant title
- Can be focused or unfocused
- Renders itself given width/height (doesn't know about layout)
- Handles its own key events when focused

### Modes

Kimchi uses a unified mode system. All modes are accessed via `m` (cycle) or number keys (`1-4`):

| Mode | Key | Description |
|------|-----|-------------|
| changed:working | `1` | Uncommitted changes only (`git diff`) |
| changed:branch | `2` | All changes vs base branch (`git diff <base>`) - **default** |
| browse | `3` | Browse all tracked files in repository |
| docs | `4` | Browse markdown files only (*.md) |

Mode switching:
- `m` - Cycle through all modes in order
- `1`/`2`/`3`/`4` - Jump directly to specific mode

When browsing all files or docs, selecting an unchanged file shows its full content instead of a diff.

### Selection Types

The app tracks what is currently selected:

| Selection | Preview Content |
|-----------|-----------------|
| File | Diff (in changed modes) or file content (in browse/docs modes) |
| Folder | Combined diff of all changed files in folder |
| Commit | Commit details + PR summary |

### Layouts

Three-slot layout for wider terminals, stacked for narrow:

```
ThreeSlot (width >= 80)              StackedThree (width < 80)
┌───────────┬───────────────────┐    ┌─────────────────────────┐
│ FileList  │                   │    │        FileList         │
│   (30%)   │                   │    ├─────────────────────────┤
├───────────┤   Preview (70%)   │    │       CommitList        │
│CommitList │                   │    ├─────────────────────────┤
│   (30%)   │                   │    │        Preview          │
└───────────┴───────────────────┘    └─────────────────────────┘
```

Window assignments:
```go
assignments := map[string]string{
    "left-top":    "filelist",
    "left-bottom": "commitlist",
    "right":       "diffview",  // or "fileview" in browse mode
}
```

### Modal Presentation

Help window is displayed as a modal overlay, centered on screen.

## Default UI

```
┌─────────────────────┬──────────────────────────────────────┐
│ Files (4)           │ src/main.go ────────────────── 42%   │
│ ▼ src/              │  12 │ func main() {          │       │
│   > main.go      M  │  13 │-    oldLine()          │       │
│     app.go       M  │  14 │+    newLine()          │       │
│ ▼ internal/         │  15 │+    anotherLine()      │       │
│   ▼ git/            │  16 │ }                      │       │
│       git.go     A  │                                      │
├─────────────────────┤                                      │
│ Commits (8)         │                                      │
│ > abc1234 Add feat  │                                      │
│   def5678 Fix bug   │                                      │
├─────────────────────┴──────────────────────────────────────┤
│ feature/blocks  [branch]  4 files  +127 -43                │
└────────────────────────────────────────────────────────────┘
```

## Status Bar

Shows at-a-glance context:

```
┌──────────────────────────────────────────────────────────┐
│ feature/blocks  [branch]  4 files  +127 -43              │
└──────────────────────────────────────────────────────────┘
  │                │        │        │
  │                │        │        └── Total diff stats
  │                │        └── File count
  │                └── Current mode (working/branch/browse/docs)
  └── Current branch
```

## FileList Window

### Tree View

Files are displayed as a tree with directories:
```
Files (4)
▼ src/
  > main.go           M
    app.go            M
▼ internal/
  ▼ git/
      git.go          A
  README.md           M
```

- Header shows "Files (N)" or "Browse (N)" based on mode
- Directories shown with `▼`/`▶` prefix (expanded/collapsed)
- `h` collapses focused folder, `l` expands it
- Collapsed folders show aggregated status indicators
- Root entry (`./`) shows combined view or PR summary
- `j`/`k` for up/down, `J`/`K` for fast navigation (5 lines)
- `g`/`G` for top/bottom

### Status Indicators
- `M` - Modified (orange)
- `A` - Added (green)
- `D` - Deleted (red)
- `?` - Untracked (muted)
- `R` - Renamed (purple)
- `C` - Has PR comments (shown alongside status)
- ` ` - Unchanged (no indicator, in browse/docs modes)

## CommitList Window

Shows recent commits on the branch:

```
Commits (8)
> abc1234 Add new feature for handling...
  def5678 Fix bug in authentication
  ghi9012 Update documentation
```

- Header shows "Commits (N)"
- Displays last 8 commits
- Selectable - when selected, preview shows commit details + PR summary
- `j`/`k` for navigation when focused

## DiffView Window

### Content Types

The DiffView displays different content based on selection:

| Selection | Content |
|-----------|---------|
| File with changes | Side-by-side diff |
| File without changes | File content with line numbers |
| Folder | Combined diff of all changed files in folder |
| Commit | Commit details + PR summary |

### Display Format

Side-by-side diff with syntax highlighting:
```
  12 │ context line            │   12 │ context line
  13 │-removed line            │      │
     │                         │   13 │+added line
  14 │ context line            │   14 │ context line
```

Colors:
- Green (`#a6e3a1`) for additions (`+`)
- Red (`#f38ba8`) for removals (`-`)
- Muted for context lines

### Line Selection

DiffView supports line-by-line navigation with a cursor:
- Cursor highlights the current line (reverse video)
- `j`/`k` moves cursor up/down one line
- `J`/`K` moves cursor 5 lines (fast navigation)
- `y` copies file path with current line number (`path/to/file.go:42`)

### Scrolling

- `j`/`k`: move cursor line by line
- `J`/`K`: move cursor 5 lines (fast navigation)
- `Ctrl+d`/`Ctrl+u`: half-page scroll
- `g`/`G`: top/bottom

Title shows file path and scroll position percentage.

### Commit & PR Summary

When a commit is selected in CommitList:
- Shows commit hash, author, date
- Shows commit message
- Shows PR summary if PR exists (title, description, reviews, comments)

## Keybindings

### Navigation
| Key | Action |
|-----|--------|
| `j` / `↓` | Move down |
| `k` / `↑` | Move up |
| `J` | Fast down (5 lines) |
| `K` | Fast up (5 lines) |
| `h` | Collapse folder |
| `l` | Expand folder |
| `Tab` | Cycle focus clockwise |
| `Shift+Tab` | Cycle focus counter-clockwise |
| `Ctrl+d` | Scroll half-page down |
| `Ctrl+u` | Scroll half-page up |
| `g` | Go to top |
| `G` | Go to bottom |

### Modes
| Key | Action |
|-----|--------|
| `m` | Cycle through all modes |
| `1` | changed:working mode |
| `2` | changed:branch mode |
| `3` | browse mode |
| `4` | docs mode |

### Actions
| Key | Action |
|-----|--------|
| `Enter` | Select item |
| `y` | Yank (copy) file path to clipboard (with line number in diff view) |
| `o` | Open file in $EDITOR |
| `r` | Refresh |
| `?` | Toggle help modal |
| `Escape` | Close modal |
| `q` | Quit |

## CLI Arguments

```
blocks [flags] [path]

Arguments:
  path              Target directory (default: current dir)

Flags:
  -m, --mode        Start in mode: working, branch (default: branch)
  -b, --base        Base branch for branch mode (default: auto-detect)
  -h, --help        Show help
  -v, --version     Show version
```

## Auto-Refresh

File changes are detected via fsnotify watching:
- `.git/index` - staging changes
- `.git/HEAD` - branch changes
- `.git/refs/heads/` - commits
- Working directory (excluding `.git`, `node_modules`, `vendor`, `__pycache__`, hidden dirs)

Changes trigger refresh after 500ms debounce.

## Git Integration

### Base Branch Detection
1. Try configured base branch (via `--base` flag)
2. Try `git config init.defaultBranch`
3. Try common names: `main`, `master`
4. Try remotes: `origin/main`, `origin/master`

## GitHub Integration

Kimchi integrates with GitHub via the `gh` CLI for PR-related features:

### PR Detection
- Automatically detects if current branch has an open PR
- Polls for updates every 60 seconds

### PR Information
When a PR exists, the following is available:
- PR title and description
- Author and status (open/merged/closed)
- Review comments and inline comments
- Comment authors and timestamps

### Inline Comments
- Files with PR comments show "C" indicator in file list
- Comments displayed inline in diff view at the relevant lines
- Comment threading supported

## Error States

| Condition | Behavior |
|-----------|----------|
| Not a git repo | Show message: "Not a git repository" with hint |
| No changes | Show empty state: "No changes" |
| Git command fails | Show error, keep last good state |
| Base branch not found | Fall back gracefully |
| Large diffs | Truncate at 10,000 lines with message |
| No `gh` CLI | PR features disabled gracefully |

## Technical Stack

- **Language**: Go
- **TUI Framework**: Bubbletea
- **Styling**: Lipgloss
- **File Watching**: fsnotify
- **Git Operations**: Shell out to git CLI
- **GitHub Operations**: Shell out to gh CLI

## Project Structure

```
blocks/
├── main.go                 # Entry point, CLI flags, app bootstrap
├── internal/
│   ├── app/
│   │   ├── app.go          # tea.Model, orchestration, Update/View
│   │   ├── state.go        # Shared state struct & transitions
│   │   ├── mode.go         # AppMode, Selection, SelectionType
│   │   └── messages.go     # All message types
│   ├── config/
│   │   └── config.go       # Centralized config: colors, styles, keybindings
│   ├── layout/
│   │   └── layout.go       # Layout definitions & rendering
│   ├── window/
│   │   ├── window.go       # Window interface
│   │   ├── base.go         # Common window functionality
│   │   ├── filelist.go     # FileList with tree view
│   │   ├── commitlist.go   # CommitList for recent commits
│   │   ├── diffview.go     # DiffView with syntax highlighting
│   │   ├── fileview.go     # FileView for browse mode
│   │   ├── prsummary.go    # PR summary renderer
│   │   └── help.go         # Help modal
│   ├── git/
│   │   ├── git.go          # Types, interface, enums
│   │   └── client.go       # GitClient implementation
│   ├── github/
│   │   └── github.go       # GitHub client for PR data
│   └── watcher/
│       └── watcher.go      # File system watcher
├── docs/
│   └── design.md
├── go.mod
└── go.sum
```

## Key Interfaces

```go
// window/window.go
type Window interface {
    Update(msg tea.Msg) (Window, tea.Cmd)
    View(width, height int) string
    Focused() bool
    SetFocus(bool)
    Name() string
}

// git/git.go
type Client interface {
    Status(mode DiffMode) ([]FileStatus, error)
    ListAllFiles() ([]FileStatus, error)
    ListDocFiles() ([]FileStatus, error)
    Diff(path string, mode DiffMode) (string, error)
    ReadFile(path string) (string, error)
    Log() ([]Commit, error)
    BaseBranch() (string, error)
    CurrentBranch() (string, error)
    DiffStats(mode DiffMode) (added, removed int, err error)
    IsRepo() bool
}

// github/github.go
type Client interface {
    IsAvailable() bool
    HasRemote() bool
    GetPRForBranch() (*PRInfo, error)
}
```

## State Management

Centralized state with message-based updates (Elm architecture):

```go
// AppMode represents the unified mode
type AppMode int
const (
    ModeChangedWorking AppMode = iota  // 1
    ModeChangedBranch                   // 2
    ModeBrowse                          // 3
    ModeDocs                            // 4
)

// SelectionType represents what is selected
type SelectionType int
const (
    SelectionNone SelectionType = iota
    SelectionFile
    SelectionFolder
    SelectionCommit
)

// Selection tracks current selection
type Selection struct {
    Type       SelectionType
    FilePath   string
    FolderPath string
    Children   []string
    Commit     *git.Commit
}

// State holds all app state
type State struct {
    Mode      AppMode
    Selection Selection

    Files         []git.FileStatus
    SelectedIndex int
    DiffContent   string

    Branch     string
    BaseBranch string
    DiffAdded   int
    DiffRemoved int

    FocusedWindow string
    ActiveModal   string

    PR    *github.PRInfo
    Error string
}
```

Message flow:
```
User Input → App.Update() → Global keys or delegate to window
    → Window returns command → App receives message
    → State update → Re-render
```

## Configuration

All configuration is centralized in `internal/config/config.go`:

### Window/Modal Names
```go
const (
    WindowFileList   = "filelist"
    WindowCommitList = "commitlist"
    WindowDiffView   = "diffview"
    WindowFileView   = "fileview"
    ModalHelp        = "help"
)
```

### Timing
```go
const (
    PRPollInterval      = 60 * time.Second
    FileWatcherDebounce = 500 * time.Millisecond
)
```

### Layout
```go
const (
    LayoutLeftRatio  = 30  // percentage
    LayoutRightRatio = 70
    LayoutBreakpoint = 80  // columns
)
```

### Diff View
```go
const (
    DiffPaneMinWidth = 40
    DiffLineNumWidth = 4
    DiffMaxLines     = 10000
    DiffTabWidth     = 4
)
```

### Keybindings
```go
var DefaultKeyMap = KeyMap{
    Up:       key.NewBinding(key.WithKeys("k", "up")),
    Down:     key.NewBinding(key.WithKeys("j", "down")),
    Left:     key.NewBinding(key.WithKeys("h")),
    Right:    key.NewBinding(key.WithKeys("l")),
    FastUp:   key.NewBinding(key.WithKeys("K")),
    FastDown: key.NewBinding(key.WithKeys("J")),
    Tab:      key.NewBinding(key.WithKeys("tab")),
    ShiftTab: key.NewBinding(key.WithKeys("shift+tab")),
    // ... etc
}
```

### Colors (Catppuccin Mocha)
```go
var DefaultColors = Colors{
    Added:    "#a6e3a1",  // Green
    Removed:  "#f38ba8",  // Red
    Modified: "#fab387",  // Peach
    Renamed:  "#cba6f7",  // Mauve
    Header:   "#89b4fa",  // Blue
    Muted:    "#6c7086",  // Overlay0
    // ...
}
```

## Future: Docs Integration

The workflow: write markdown specs → AI implements → review changes.

Potential features:
- **DocsView** - Markdown viewer in terminal with rendering
- **Spec linking** - Associate a spec with current work/branch
- **Split context** - Spec on left, implementation diff on right
