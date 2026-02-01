# Kimchi Architecture

## Overview

Kimchi is a terminal UI for code review built with Rust and Ratatui.

## Component Architecture

```
┌─────────────────────────────────────────────────────────────────────────────────┐
│                                   main.rs                                        │
│  (src/main.rs)                                                                  │
│                                                                                  │
│  Responsibilities:                                                               │
│  ├── Terminal setup (raw mode, alternate screen)                                │
│  ├── Create App and EventHandler                                                │
│  └── Run main event loop                                                        │
└─────────────────────────────────────────────────────────────────────────────────┘
         │                              │
         │ creates                      │ creates
         ▼                              ▼
┌─────────────────────────┐    ┌─────────────────────────────────────────────────┐
│      EventHandler       │    │                      App                         │
│  (src/event.rs)         │    │  (src/app.rs)                                   │
│                         │    │                                                  │
│  State:                 │    │  State:                                          │
│  ├── rx: Receiver       │    │  ├── mode: AppMode                               │
│  ├── paused: AtomicBool │    │  ├── focused: FocusedWindow                      │
│  └── watcher: Debouncer │    │  ├── files: Vec<StatusEntry>                     │
│                         │    │  ├── commits: Vec<Commit>                        │
│  Methods:               │    │  ├── pr: Option<PrInfo>                          │
│  ├── next() → AppEvent  │    │  └── *_state: widget states                      │
│  ├── pause()            │    │                                                  │
│  └── resume()           │    │  Methods:                                        │
│                         │    │  ├── handle_key() → updates state                │
└─────────────────────────┘    │  ├── handle_tick() → async loading               │
         │                     │  ├── refresh() → reload git data                 │
         │ sends               │  └── render() → draw UI                          │
         ▼                     └─────────────────────────────────────────────────┘
    AppEvent                            │
    ├── Key(KeyEvent)                   │ uses
    ├── Tick                            ▼
    ├── FileChanged          ┌─────────────────────────────────────────────────────┐
    └── Resize               │                    Clients                          │
                             │                                                     │
                             │  ┌─────────────────┐    ┌─────────────────────────┐ │
                             │  │   GitClient     │    │    GitHubClient         │ │
                             │  │ (src/git/)      │    │  (src/github/)          │ │
                             │  │                 │    │                         │ │
                             │  │ • status()      │    │ • get_pr_for_branch()   │ │
                             │  │ • diff()        │    │ • get_reviews()         │ │
                             │  │ • log()         │    │ • get_comments()        │ │
                             │  │ • read_file()   │    │                         │ │
                             │  │                 │    │  Uses: gh CLI           │ │
                             │  │ Uses: libgit2   │    │                         │ │
                             │  └─────────────────┘    └─────────────────────────┘ │
                             └─────────────────────────────────────────────────────┘
```

## Event Flow

```
┌──────────┐    ┌──────────────┐    ┌────────────────┐    ┌─────────────┐
│  User    │    │ EventHandler │    │      App       │    │  Terminal   │
└────┬─────┘    └──────┬───────┘    └───────┬────────┘    └──────┬──────┘
     │                 │                    │                    │
     │ keypress        │                    │                    │
     │────────────────>│                    │                    │
     │                 │                    │                    │
     │                 │ AppEvent::Key      │                    │
     │                 │───────────────────>│                    │
     │                 │                    │                    │
     │                 │                    │ handle_key()       │
     │                 │                    │ ─ check global keys│
     │                 │                    │ ─ delegate to      │
     │                 │                    │   focused window   │
     │                 │                    │ ─ update state     │
     │                 │                    │                    │
     │                 │                    │ render()           │
     │                 │                    │───────────────────>│
     │                 │                    │                    │
     │  UI updated     │                    │                    │
     │<───────────────────────────────────────────────────────────
     │                 │                    │                    │
```

## Async Data Loading

```
┌─────────────────┐                      ┌─────────────────┐
│                 │   spawn_stats_loader │                 │
│   Main Thread   │ ──────────────────►  │ Background      │
│   (App)         │                      │ Thread          │
│                 │ ◄──────────────────  │                 │
│                 │   mpsc::channel      │ • GitClient     │
│                 │                      │ • diff_stats()  │
└─────────────────┘                      └─────────────────┘

┌─────────────────┐                      ┌─────────────────┐
│                 │   spawn_pr_loader    │                 │
│   Main Thread   │ ──────────────────►  │ Background      │
│   (App)         │                      │ Thread          │
│                 │ ◄──────────────────  │                 │
│                 │   mpsc::channel      │ • GitHubClient  │
│                 │                      │ • gh CLI        │
└─────────────────┘                      └─────────────────┘

On each Tick:
  1. try_recv() on stats_rx → apply DiffStats if ready
  2. try_recv() on pr_rx → apply PrInfo if ready
  3. Trigger new loaders if needed
```

## UI Widget Hierarchy

```
┌─────────────────────────────────────────────────────────────────────────────────┐
│                              Terminal Frame                                      │
│                                                                                  │
│  ┌─────────────────────────┐  ┌─────────────────────────────────────────────┐   │
│  │                         │  │                                             │   │
│  │      FileList           │  │              DiffView                       │   │
│  │  (file_list.rs)         │  │          (diff_view.rs)                     │   │
│  │                         │  │                                             │   │
│  │  • Tree view of files   │  │  PreviewContent:                            │   │
│  │  • Directory collapse   │  │  ├── FileDiff (side-by-side)                │   │
│  │  • Status indicators    │  │  ├── FolderDiff (combined)                  │   │
│  │                         │  │  ├── FileContent (browse mode)              │   │
│  ├─────────────────────────┤  │  ├── CommitSummary (PR info)                │   │
│  │                         │  │  └── Empty                                  │   │
│  │      CommitList         │  │                                             │   │
│  │  (commit_list.rs)       │  │                                             │   │
│  │                         │  │                                             │   │
│  │  • Recent commits       │  │                                             │   │
│  │  • Short hash + subject │  │                                             │   │
│  │                         │  │                                             │   │
│  └─────────────────────────┘  └─────────────────────────────────────────────┘   │
│                                                                                  │
│  ┌─────────────────────────────────────────────────────────────────────────────┐│
│  │                              Status Bar                                      ││
│  │  branch | mode | file count | +added -removed                               ││
│  └─────────────────────────────────────────────────────────────────────────────┘│
│                                                                                  │
│  ┌─────────────────────────────────────────────────────────────────────────────┐│
│  │                           HelpModal (overlay)                                ││
│  │                         (help.rs, toggled with ?)                           ││
│  └─────────────────────────────────────────────────────────────────────────────┘│
└─────────────────────────────────────────────────────────────────────────────────┘
```

## File Watcher Flow

```
┌─────────────────┐
│   .git/index    │
│   (file)        │
└────────┬────────┘
         │
         │ notify crate watches
         ▼
┌─────────────────┐
│   Debouncer     │
│   (500ms)       │
└────────┬────────┘
         │
         │ DebouncedEventKind::Any
         ▼
┌─────────────────┐
│  EventHandler   │
│  tx.send(       │
│    FileChanged) │
└────────┬────────┘
         │
         │ rx.recv()
         ▼
┌─────────────────┐
│      App        │
│  refresh()      │
│  ─ reload files │
│  ─ reload diff  │
│  ─ update UI    │
└─────────────────┘
```

## Application Modes

```
                    ┌─────────────────┐
                    │                 │
         ┌─────────>│ ChangedWorking  │<─────────┐
         │          │   (Mode 1)      │          │
         │          │  git diff       │          │
         │          └────────┬────────┘          │
         │                   │                   │
    press 1             press m              press 1
         │                   │                   │
         │                   ▼                   │
┌────────┴────────┐  ┌───────────────┐  ┌───────┴────────┐
│                 │  │               │  │                │
│     Docs        │  │ ChangedBranch │──│    Browse      │
│   (Mode 4)      │  │   (Mode 2)    │  │   (Mode 3)     │
│                 │  │  git diff     │  │                │
│  *.md files     │  │  <base>       │  │  All files     │
│                 │  │               │  │                │
└─────────────────┘  └───────────────┘  └────────────────┘
         ▲                   │                   ▲
         │              press m                  │
         │                   │                   │
         └───────────────────┴───────────────────┘
```

## Key Input Handling

```
┌──────────────────────────┐
│ KeyEvent received        │
└────────────┬─────────────┘
             │
             ▼
    ╔════════════════════╗
   ╱  show_help == true? ╲
  ╱                       ╲
 yes                      no
  │                        │
  ▼                        ▼
┌────────────────┐   ╔═══════════════╗
│ Only handle    │  ╱ Global key?    ╲
│ ? or Esc       │ ╱  (q, ?, r, Tab,  ╲
│ to close help  │╱   m, 1-4, y, o)   ╲
└────────────────┘yes                  no
                   │                    │
                   ▼                    ▼
          ┌──────────────┐    ┌──────────────────┐
          │ Handle       │    │ Delegate to      │
          │ globally:    │    │ focused window:  │
          │ • quit       │    │                  │
          │ • mode switch│    │ • FileList keys  │
          │ • yank/open  │    │ • CommitList keys│
          │ • refresh    │    │ • Preview keys   │
          └──────────────┘    └──────────────────┘
```

## External Editor Integration

```
┌─────────┐     ┌─────────┐     ┌─────────┐     ┌─────────┐
│ User    │────>│ App     │────>│ Command │────>│ Editor  │
│ press o │     │ queues  │     │ execute │     │ opens   │
└─────────┘     │ command │     └─────────┘     └─────────┘
                └─────────┘
                     │
                     ▼
            ┌─────────────────┐
            │ Terminal state: │
            │ 1. Pause events │
            │ 2. Leave alt    │
            │    screen       │
            │ 3. Disable raw  │
            │    mode         │
            └────────┬────────┘
                     │
                     ▼
            ┌─────────────────┐
            │ Run $EDITOR     │
            │ with path:line  │
            │                 │
            │ vim +42 file.rs │
            │ hx file.rs:42   │
            └────────┬────────┘
                     │
                     ▼
            ┌─────────────────┐
            │ Restore:        │
            │ 1. Enable raw   │
            │ 2. Enter alt    │
            │    screen       │
            │ 3. Resume events│
            │ 4. Refresh data │
            └─────────────────┘
```
