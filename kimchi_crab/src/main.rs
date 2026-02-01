mod app;
mod config;
mod event;
mod git;
mod github;
mod ui;

use anyhow::Result;
use clap::Parser;
use crossterm::{
    event::{DisableMouseCapture, EnableMouseCapture},
    execute,
    terminal::{disable_raw_mode, enable_raw_mode, EnterAlternateScreen, LeaveAlternateScreen},
};
use ratatui::prelude::*;
use std::io;
use std::path::PathBuf;
use std::time::Duration;

use app::App;
use event::{AppEvent, EventHandler};

/// Kimchi - AI-native code review TUI
#[derive(Parser, Debug)]
#[command(name = "kimchi")]
#[command(author, version, about, long_about = None)]
struct Args {
    /// Path to git repository
    #[arg(default_value = ".")]
    path: PathBuf,
}

fn main() -> Result<()> {
    let args = Args::parse();

    // Resolve path
    let path = args.path.canonicalize().unwrap_or(args.path);

    // Initialize terminal
    enable_raw_mode()?;
    let mut stdout = io::stdout();
    execute!(stdout, EnterAlternateScreen, EnableMouseCapture)?;
    let backend = CrosstermBackend::new(stdout);
    let mut terminal = Terminal::new(backend)?;

    // Create app
    let mut app = App::new(path.to_str().unwrap_or("."))?;

    // Create event handler
    let events = EventHandler::new(Duration::from_millis(100));

    // Main loop
    let result = run_app(&mut terminal, &mut app, &events);

    // Restore terminal
    disable_raw_mode()?;
    execute!(
        terminal.backend_mut(),
        LeaveAlternateScreen,
        DisableMouseCapture
    )?;
    terminal.show_cursor()?;

    result
}

fn run_app<B: Backend>(
    terminal: &mut Terminal<B>,
    app: &mut App,
    events: &EventHandler,
) -> Result<()> {
    while app.running {
        // Draw
        terminal.draw(|frame| {
            app.render(frame);
        })?;

        // Handle events
        match events.next()? {
            AppEvent::Key(key) => {
                app.handle_key(key)?;
            }
            AppEvent::Resize(_, _) => {
                // Terminal will redraw automatically
            }
            AppEvent::Tick => {
                // Could do periodic refresh here
            }
            AppEvent::FileChanged => {
                app.refresh()?;
            }
            AppEvent::PrLoaded => {
                // PR data updated
            }
        }
    }

    Ok(())
}
