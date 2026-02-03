# TODO

Minor changes and refactors to address.

---

## Widget Refactoring

### Goal
Clean separation between widgets with simple state handling. App owns all state, widgets return Actions describing what happened.

### Current Issues
- Inconsistent widget APIs (some have `handle_key`, some don't)
- Naming inconsistencies (`pr_info.rs` exports `PrListPanel`)
- `diff_parser.rs` is a utility living in widgets folder
- PR details mixed into diff_view
- No clear pattern for widget → app communication

### Architecture

**Simple rule:** App owns state, widgets return Actions, App dispatches.

```rust
// Widget handles input, returns what happened
impl FileListState {
    pub fn handle_key(&mut self, key: KeyEvent) -> Action {
        match key {
            Key::Enter => Action::FileSelected(self.selected_path()),
            Key::Up => { self.move_up(); Action::None }
            _ => Action::None
        }
    }
}

// App dispatches the action
let action = self.file_list.handle_key(key);
if let Action::FileSelected(path) = action {
    self.diff.load(&path);
}
```

### Action Enum

```rust
pub enum Action {
    None,

    // File list
    FileSelected(PathBuf),

    // PR list
    PrSelected(u64),

    // Diff view
    OpenComment { path: String, line: u32 },

    // Input
    SubmitReview(ReviewAction),
    CancelInput,

    // Navigation
    ChangeFocus(FocusedWidget),
    Quit,
}
```

### File Structure

```
src/ui/widgets/
├── mod.rs                 # exports Action + all widgets
├── action.rs              # Action enum
├── file_list/
│   ├── mod.rs             # FileList + FileListState
│   └── tree.rs            # TreeEntry logic
├── diff_view/
│   ├── mod.rs             # DiffView + DiffViewState
│   └── parser.rs          # DiffLine, diff parsing (moved from diff_parser.rs)
├── pr_details/            # NEW - extracted from diff_view
│   └── mod.rs
├── pr_list/               # renamed from pr_info
│   └── mod.rs
├── input/                 # renamed from input_modal
│   └── mod.rs
└── help/
    └── mod.rs
```

### Widget Interface

Based on existing widgets, keep it minimal:

```rust
/// Minimal trait - just key handling
pub trait WidgetState {
    fn handle_key(&mut self, key: KeyEvent) -> Action;
}
```

**What stays the same:**
- `StatefulWidget` for rendering (Ratatui's trait)
- `.focused(bool)` builder on Widget for styling
- Widget-specific methods (`set_files()`, `set_content()`, etc.)

**Pattern for new widgets:**

```rust
// 1. State struct - holds data, implements WidgetState
pub struct MyWidgetState {
    items: Vec<Item>,
    selected: usize,
    // NO focused field - that's on the Widget
}

impl WidgetState for MyWidgetState {
    fn handle_key(&mut self, key: KeyEvent) -> Action {
        match key.code {
            KeyCode::Up => { self.selected = self.selected.saturating_sub(1); Action::None }
            KeyCode::Down => { self.selected += 1; Action::None }
            KeyCode::Enter => Action::ItemSelected(self.items[self.selected].clone()),
            _ => Action::None
        }
    }
}

// 2. Widget struct - rendering only, takes &State
pub struct MyWidget<'a> {
    colors: &'a Colors,
    focused: bool,
}

impl<'a> MyWidget<'a> {
    pub fn new(colors: &'a Colors) -> Self {
        Self { colors, focused: false }
    }

    pub fn focused(mut self, focused: bool) -> Self {
        self.focused = focused;
        self
    }
}

impl<'a> StatefulWidget for MyWidget<'a> {
    type State = MyWidgetState;

    fn render(self, area: Rect, buf: &mut Buffer, state: &mut Self::State) {
        // render using state + self.focused for styling
    }
}
```

**Plug-and-play in App:**

```rust
// App stores states
pub struct App {
    file_list: FileListState,
    diff_view: DiffViewState,
    my_widget: MyWidgetState,  // just add new state
    focus: FocusedWidget,
}

// Uniform key handling
fn handle_key(&mut self, key: KeyEvent) {
    let action = match self.focus {
        Focus::Files => self.file_list.handle_key(key),
        Focus::Diff => self.diff_view.handle_key(key),
        Focus::MyWidget => self.my_widget.handle_key(key),
    };
    self.dispatch(action);
}

// Uniform rendering
fn render(&self, frame: &mut Frame) {
    frame.render_stateful_widget(
        MyWidget::new(&self.colors).focused(self.focus == Focus::MyWidget),
        area,
        &mut self.my_widget,
    );
}
```

### Tasks

- [ ] Create `Action` enum in `src/ui/widgets/action.rs`
- [ ] Create `WidgetState` trait in `src/ui/widgets/mod.rs`
- [ ] Add `handle_key() -> Action` to all widget states
- [ ] Update `App.handle_key()` to dispatch actions
- [ ] Rename `pr_info.rs` → `pr_list/`
- [ ] Rename `input_modal.rs` → `input/`
- [ ] Move `diff_parser.rs` → `diff_view/parser.rs`
- [ ] Extract PR details rendering to `pr_details/`
- [ ] Organize each widget into its own folder

---

## Event-Driven Architecture

### Goal
All state changes flow through events. No polling, no special cases. Timers and watchers spawn events.

### Current Issues
- PR list refreshes on timer via polling in `handle_tick()`
- No branch change detection
- `AsyncLoader` is PR-specific, not reusable
- Mixing polling and event handling

### Unified Event System

```rust
pub enum AppEvent {
    // Input
    Key(KeyEvent),

    // Triggers (from timers, watchers, user actions)
    RefreshPrList,
    RefreshGitStatus,
    BranchChanged(String),

    // Async completions
    PrListLoaded(Vec<PrSummary>),
    PrDetailsLoaded(PrInfo),
    LoadFailed { task: String, error: String },

    // Periodic
    Tick,
}
```

### Generic AsyncLoader

```rust
pub struct AsyncLoader {
    tx: mpsc::Sender<AppEvent>,
}

impl AsyncLoader {
    /// Spawn any async task, result comes back as an event
    pub fn spawn<F, R>(&self, task: F, on_complete: impl Fn(R) -> AppEvent + Send + 'static)
    where
        F: FnOnce() -> R + Send + 'static,
        R: Send + 'static,
    {
        let tx = self.tx.clone();
        thread::spawn(move || {
            let result = task();
            let _ = tx.send(on_complete(result));
        });
    }
}

// Usage - PR list
self.loader.spawn(
    || github.list_open_prs().unwrap_or_default(),
    AppEvent::PrListLoaded,
);

// Usage - PR details
let pr_num = pr_number;
self.loader.spawn(
    move || github.get_pr_by_number(pr_num),
    |result| match result {
        Ok(Some(pr)) => AppEvent::PrDetailsLoaded(pr),
        _ => AppEvent::LoadFailed { task: "pr_details".into(), error: "not found".into() },
    },
);
```

### Branch Watcher

Watch `.git/HEAD` for branch changes:

```rust
// In file watcher setup, watch .git/HEAD separately
watcher.watch(".git/HEAD", RecursiveMode::NonRecursive);

// When .git/HEAD changes
if path.ends_with("HEAD") {
    let branch = git.current_branch();
    tx.send(AppEvent::BranchChanged(branch));
}
```

### Timers Spawn Events

```rust
// Timers managed separately, spawn events when due
impl TimerManager {
    pub fn check(&mut self, tx: &mpsc::Sender<AppEvent>) {
        if self.pr_list_timer.elapsed() >= self.pr_poll_interval {
            let _ = tx.send(AppEvent::RefreshPrList);
            self.pr_list_timer = Instant::now();
        }
    }
}
```

### Event Flow

```
┌─────────────┐    ┌─────────────┐    ┌─────────────┐
│   Timers    │    │  Watchers   │    │  AsyncLoader│
│             │    │             │    │             │
│ PR refresh  │    │ .git/HEAD   │    │ task done   │
│ every 30s   │    │ file changes│    │             │
└──────┬──────┘    └──────┬──────┘    └──────┬──────┘
       │                  │                  │
       ▼                  ▼                  ▼
┌─────────────────────────────────────────────────┐
│              Event Channel (mpsc)               │
└─────────────────────────┬───────────────────────┘
                          │
                          ▼
┌─────────────────────────────────────────────────┐
│                 App.handle_event()              │
│                                                 │
│  RefreshPrList    → loader.spawn(fetch_prs)    │
│  PrListLoaded     → pr_list_state.set_prs()    │
│  BranchChanged    → refresh git, clear PR      │
│  Key              → widget.handle_key()        │
└─────────────────────────────────────────────────┘
```

### Tasks

- [ ] Add new event variants to `AppEvent`
- [ ] Make `AsyncLoader` generic (takes event sender)
- [ ] Add `.git/HEAD` to file watcher for branch detection
- [ ] Extract timer logic from `handle_tick()`
- [ ] Create `App.handle_event()` dispatcher
- [ ] Remove polling from `handle_tick()`

---

## Cleanup

### Dead Code & Obsolete Features

After iterative development, audit the codebase for:

- **Unused functions/methods** - run `cargo clippy` with `dead_code` warnings
- **Unused imports** - `cargo fix` can help
- **Over-engineered features** - simplify where possible
- **Commented-out code** - delete it, git has history
- **Unused dependencies** - check `Cargo.toml`

### Audit Checklist

- [ ] Run `cargo clippy -- -W dead_code` and fix warnings
- [ ] Run `cargo +nightly udeps` to find unused dependencies
- [ ] Review each widget for unused methods
- [ ] Check for overly complex abstractions that can be simplified
- [ ] Remove any feature flags or config options that aren't used
- [ ] Look for TODO/FIXME comments that are stale

### Candidates to Review

- `PreviewContent` variants - are all used?
- `TimelinePosition` - is the full range needed?
- Config options in `config.rs` - which are actually configurable?
- Helper functions in `event.rs` - `KeyInput` methods all used?

---

## Minor Fixes
