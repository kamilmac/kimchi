# Kimchi

Terminal UI for reviewing code changes. Built for the AI coding era.

## Why

AI agents write code now. Claude Code, Cursor, Copilot — they commit faster than you can review. This creates problems:

1. **You lose track of what's changing.** Agents commit frequently. `git diff` against your working tree shows nothing useful. What you need is diff against your base branch — the full picture of what's being built.

2. **PRs need faster access.** More code means more PRs, more comments, more review cycles. Switching between terminal and browser kills flow. Review actions should be a keystroke away.

3. **Code is secondary.** Controversial, but: when an agent writes the code, your job shifts from writing to reviewing. The diff becomes the primary artifact, not the source file. I want diffs side-by-side with my agent, not buried in a git command.

Kimchi sits in a terminal pane next to your AI agent. It shows what changed, refreshes automatically, and lets you review PRs without leaving the terminal.

```
┌─ Files (4) ────────┬─ src/app.rs ─────────────────────────────┐
│ ▼ src/             │  12 │ fn main() {         fn main() {   │
│   > app.rs      M  │  13 │-    old_call();                    │
│     config.rs   M  │  14 │+                    new_call();    │
├────────────────────┤  15 │ }                   }              │
│                    │                                          │
├─ PRs (2) ──────────┴──────────────────────────────────────────┤
│●  #42 │ you        │ Add feature                     │  today │
│ ◆ #38 │ teammate   │ Fix bug                         │     2d │
└───────────────────────────────────────────────────────────────┘
```

## Install

```bash
git clone https://github.com/kmacinski/kimchi
cd kimchi
cargo build --release
cp target/release/kimchi ~/.local/bin/
```

## Requirements

- **Git**
- **gh CLI** (optional) — for PR list, reviews, and comments. [Install here](https://cli.github.com/). Without it, PR features are disabled but local git operations work fine.

## Usage

```bash
kimchi              # current directory
kimchi /path/to/repo
```

## Keys

| Key | Action |
|-----|--------|
| `j/k` | Navigate |
| `J/K` | Fast scroll |
| `h/l` | Scroll diff horizontally |
| `g/G` | Top/bottom |
| `Tab` | Switch panes |
| `1-4` | Switch mode |
| `y` | Copy path |
| `o` | Open in $EDITOR |
| `?` | Help |
| `q` | Quit |

**PR actions** (requires `gh`): `a` approve, `x` request changes, `c` comment

## Modes

| `1` working | Uncommitted changes |
|-------------|---------------------|
| `2` branch | Changes vs base branch |
| `3` browse | All tracked files |
| `4` docs | Markdown files |

## License

MIT
