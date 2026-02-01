# Kimchi Crab 🦀

The Rust rewrite of Kimchi - faster, leaner, better.

## Why Rust?

- **Native git**: Uses libgit2 instead of shelling out to git
- **Zero-cost abstractions**: Compile-time guarantees, runtime speed
- **Memory safe**: No GC, no runtime overhead
- **Single binary**: No dependencies, just copy and run

## Build

```bash
cargo build --release
```

The binary will be at `target/release/kimchi`.

## Features

Same as Go version, but faster:

- Side-by-side diffs with syntax highlighting
- File tree with collapsible folders
- Commit history with PR integration
- Multiple modes (working/branch/browse/docs)
- Vim-style navigation
- Auto-refresh on file changes
- Clipboard support
- Open in $EDITOR

## Architecture

```
src/
├── main.rs           # Entry point, CLI, terminal setup
├── app.rs            # App state, event handling, rendering
├── config.rs         # Colors, timing, layout config
├── event.rs          # Terminal event handling
├── git/
│   ├── mod.rs
│   ├── types.rs      # FileStatus, AppMode, Commit, etc.
│   └── client.rs     # GitClient using libgit2
├── github/
│   └── mod.rs        # GitHub client (gh CLI)
└── ui/
    ├── mod.rs
    ├── layout.rs     # Layout computation
    └── widgets/
        ├── file_list.rs
        ├── commit_list.rs
        ├── diff_view.rs
        └── help.rs
```

## Dependencies

- **ratatui**: TUI framework (modern tui-rs fork)
- **crossterm**: Cross-platform terminal handling
- **git2**: libgit2 bindings for native git operations
- **tokio**: Async runtime for file watching
- **clap**: CLI argument parsing
- **arboard**: Clipboard support

## License

MIT
