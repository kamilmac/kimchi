# Kimchi

A read-first terminal IDE for AI-driven development workflows.

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
│ feature/branch  [branch]  4 files  +127 -43                │
└────────────────────────────────────────────────────────────┘
```

## Why Kimchi?

When AI writes code, you don't need a traditional IDE. You need:

- **Visibility** - see what the AI is changing
- **Navigation** - understand the codebase
- **Review** - approve changes with confidence
- **Context** - specs and docs alongside code

This is read-heavy, not write-heavy. The human reviews, the AI writes.

## Features

- **Side-by-side diffs** - See changes with syntax highlighting
- **File tree** - Navigate changed files with collapsible folders
- **Commit history** - View recent commits with PR integration
- **Multiple modes** - Switch between working changes, branch diff, browse, and docs
- **PR integration** - See inline comments and review status via GitHub CLI
- **Vim-style navigation** - `j`/`k`, `g`/`G`, `Ctrl+d`/`Ctrl+u`
- **Auto-refresh** - Changes appear automatically (500ms debounce)
- **Quick actions** - `y` to yank path, `o` to open in editor

## Installation

```bash
go install github.com/kmacinski/blocks@latest
```

Or build from source:

```bash
git clone https://github.com/kmacinski/blocks
cd blocks
go build -o kimchi .
```

## Usage

```bash
kimchi              # Run in current directory
kimchi /path/to/repo   # Run in specific directory
```

### Modes

Press `m` to cycle through modes, or use number keys:

| Key | Mode | Description |
|-----|------|-------------|
| `1` | changed:working | Uncommitted changes only |
| `2` | changed:branch | All changes vs base branch (default) |
| `3` | browse | Browse all files in repo |
| `4` | docs | Browse markdown files only |

### Navigation

| Key | Action |
|-----|--------|
| `j`/`k` | Move up/down |
| `J`/`K` | Fast move (5 lines) |
| `h`/`l` | Collapse/expand folder |
| `Tab` | Cycle focus clockwise |
| `Ctrl+d`/`u` | Half-page scroll |
| `g`/`G` | Go to top/bottom |

### Actions

| Key | Action |
|-----|--------|
| `y` | Copy file path (with line number in diff) |
| `o` | Open in `$EDITOR` |
| `r` | Refresh |
| `?` | Show help |
| `q` | Quit |

## GitHub Integration

Kimchi integrates with `gh` CLI for PR features:

- Automatic PR detection for current branch
- Inline comment display in diff view
- Review status and PR description
- Files with comments marked with `C` indicator

Install GitHub CLI: https://cli.github.com/

## Requirements

- Go 1.21+
- Git
- Terminal with 256 color support
- Optional: `gh` CLI for GitHub features

## License

MIT
