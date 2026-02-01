use ratatui::{
    buffer::Buffer,
    layout::Rect,
    style::Modifier,
    text::{Line, Span},
    widgets::{Block, Borders, Clear, StatefulWidget, Widget},
};

use crate::config::Colors;
use crate::github::PrSummary;

/// PR list modal state
#[derive(Debug, Default)]
pub struct PrListState {
    pub prs: Vec<PrSummary>,
    pub cursor: usize,
    pub loading: bool,
    pub error: Option<String>,
}

impl PrListState {
    pub fn new() -> Self {
        Self {
            loading: true,
            ..Default::default()
        }
    }

    pub fn set_prs(&mut self, prs: Vec<PrSummary>) {
        self.prs = prs;
        self.loading = false;
        self.cursor = 0;
    }

    pub fn set_error(&mut self, err: String) {
        self.error = Some(err);
        self.loading = false;
    }

    pub fn move_down(&mut self) {
        if self.cursor < self.prs.len().saturating_sub(1) {
            self.cursor += 1;
        }
    }

    pub fn move_up(&mut self) {
        self.cursor = self.cursor.saturating_sub(1);
    }

    pub fn selected(&self) -> Option<&PrSummary> {
        self.prs.get(self.cursor)
    }
}

/// PR list modal widget
pub struct PrListModal<'a> {
    colors: &'a Colors,
}

impl<'a> PrListModal<'a> {
    pub fn new(colors: &'a Colors) -> Self {
        Self { colors }
    }
}

impl<'a> StatefulWidget for PrListModal<'a> {
    type State = PrListState;

    fn render(self, area: Rect, buf: &mut Buffer, state: &mut Self::State) {
        // Clear background
        Clear.render(area, buf);

        let title = format!(" Open Pull Requests ({}) ", state.prs.len());
        let block = Block::default()
            .borders(Borders::ALL)
            .border_style(self.colors.style_border_focused())
            .title(Span::styled(title, self.colors.style_header()));

        let inner = block.inner(area);
        block.render(area, buf);

        if state.loading {
            let msg = Line::from(Span::styled("Loading...", self.colors.style_muted()));
            buf.set_line(inner.x + 2, inner.y + 1, &msg, inner.width - 4);
            return;
        }

        if let Some(err) = &state.error {
            let msg = Line::from(Span::styled(
                format!("Error: {}", err),
                self.colors.style_removed(),
            ));
            buf.set_line(inner.x + 2, inner.y + 1, &msg, inner.width - 4);
            return;
        }

        if state.prs.is_empty() {
            let msg = Line::from(Span::styled("No open PRs", self.colors.style_muted()));
            buf.set_line(inner.x + 2, inner.y + 1, &msg, inner.width - 4);
            return;
        }

        // Render PR list
        let visible_height = inner.height.saturating_sub(2) as usize;
        let start = if state.cursor >= visible_height {
            state.cursor - visible_height + 1
        } else {
            0
        };

        for (i, pr) in state.prs.iter().skip(start).take(visible_height).enumerate() {
            let y = inner.y + 1 + i as u16;
            let is_selected = start + i == state.cursor;

            let pr_num = format!("#{:<5}", pr.number);
            let author = format!("{:<12}", truncate(&pr.author, 12));
            let date = format!("{}", pr.updated_at);
            let title_width = inner.width as usize - 30;
            let title = truncate(&pr.title, title_width);

            let style = if is_selected {
                self.colors.style_selected().add_modifier(Modifier::REVERSED)
            } else {
                ratatui::style::Style::default().fg(self.colors.text)
            };

            let line = Line::from(vec![
                Span::styled(pr_num, self.colors.style_header()),
                Span::raw(" "),
                Span::styled(author, self.colors.style_muted()),
                Span::raw(" "),
                Span::styled(title, style),
                Span::raw(" "),
                Span::styled(date, self.colors.style_muted()),
            ]);

            buf.set_line(inner.x + 1, y, &line, inner.width - 2);
        }

        // Footer with instructions
        let footer_y = inner.y + inner.height - 1;
        let footer = Line::from(vec![
            Span::styled("Enter", self.colors.style_header()),
            Span::styled(" checkout  ", self.colors.style_muted()),
            Span::styled("o", self.colors.style_header()),
            Span::styled(" open in browser  ", self.colors.style_muted()),
            Span::styled("Esc", self.colors.style_header()),
            Span::styled(" close", self.colors.style_muted()),
        ]);
        buf.set_line(inner.x + 1, footer_y, &footer, inner.width - 2);
    }
}

fn truncate(s: &str, max: usize) -> String {
    if s.chars().count() > max {
        s.chars().take(max.saturating_sub(1)).collect::<String>() + "…"
    } else {
        s.to_string()
    }
}
