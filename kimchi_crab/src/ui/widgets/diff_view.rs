use ratatui::{
    buffer::Buffer,
    layout::Rect,
    text::{Line, Span},
    widgets::{Block, Borders, StatefulWidget, Widget},
};

use crate::config::Colors;
use crate::git::Commit;
use crate::github::PrInfo;

/// What to show in the diff view
#[derive(Debug, Clone)]
pub enum PreviewContent {
    Empty,
    FileDiff {
        path: String,
        content: String,
    },
    FolderDiff {
        path: String,
        content: String,
    },
    FileContent {
        path: String,
        content: String,
    },
    CommitSummary {
        commit: Commit,
        pr: Option<PrInfo>,
    },
}

impl Default for PreviewContent {
    fn default() -> Self {
        Self::Empty
    }
}

/// Diff view widget state
#[derive(Debug, Default)]
pub struct DiffViewState {
    pub content: PreviewContent,
    pub lines: Vec<DiffLine>,
    pub cursor: usize,
    pub offset: usize,
}

#[derive(Debug, Clone)]
pub struct DiffLine {
    pub text: String,
    pub line_type: LineType,
    pub left_num: Option<usize>,
    pub right_num: Option<usize>,
}

#[derive(Debug, Clone, Copy, PartialEq)]
pub enum LineType {
    Context,
    Added,
    Removed,
    Header,
    Info,
}

impl DiffViewState {
    pub fn new() -> Self {
        Self::default()
    }

    pub fn set_content(&mut self, content: PreviewContent) {
        self.content = content;
        self.cursor = 0;
        self.offset = 0;
        self.parse_content();
    }

    fn parse_content(&mut self) {
        self.lines = match &self.content {
            PreviewContent::Empty => vec![],
            PreviewContent::FileDiff { content, .. } | PreviewContent::FolderDiff { content, .. } => {
                if is_binary(content) {
                    vec![DiffLine {
                        text: "Binary file".to_string(),
                        line_type: LineType::Info,
                        left_num: None,
                        right_num: None,
                    }]
                } else {
                    parse_diff(content)
                }
            }
            PreviewContent::FileContent { content, .. } => {
                if is_binary(content) {
                    vec![DiffLine {
                        text: "Binary file".to_string(),
                        line_type: LineType::Info,
                        left_num: None,
                        right_num: None,
                    }]
                } else {
                    parse_file_content(content)
                }
            }
            PreviewContent::CommitSummary { commit, pr } => {
                parse_commit_summary(commit, pr.as_ref())
            }
        };
    }

    pub fn title(&self) -> String {
        match &self.content {
            PreviewContent::Empty => "Preview".to_string(),
            PreviewContent::FileDiff { path, .. } => path.clone(),
            PreviewContent::FolderDiff { path, .. } => format!("{}/", path),
            PreviewContent::FileContent { path, .. } => path.clone(),
            PreviewContent::CommitSummary { .. } => "Commit & PR Summary".to_string(),
        }
    }

    pub fn move_down(&mut self) {
        if self.cursor < self.lines.len().saturating_sub(1) {
            self.cursor += 1;
        }
    }

    pub fn move_up(&mut self) {
        self.cursor = self.cursor.saturating_sub(1);
    }

    pub fn move_down_n(&mut self, n: usize) {
        self.cursor = (self.cursor + n).min(self.lines.len().saturating_sub(1));
    }

    pub fn move_up_n(&mut self, n: usize) {
        self.cursor = self.cursor.saturating_sub(n);
    }

    pub fn page_down(&mut self, height: usize) {
        self.move_down_n(height / 2);
    }

    pub fn page_up(&mut self, height: usize) {
        self.move_up_n(height / 2);
    }

    pub fn go_top(&mut self) {
        self.cursor = 0;
        self.offset = 0;
    }

    pub fn go_bottom(&mut self) {
        self.cursor = self.lines.len().saturating_sub(1);
    }

    pub fn ensure_visible(&mut self, height: usize) {
        let visible_height = height.saturating_sub(3);
        if self.cursor < self.offset {
            self.offset = self.cursor;
        } else if self.cursor >= self.offset + visible_height {
            self.offset = self.cursor.saturating_sub(visible_height) + 1;
        }
    }

    pub fn scroll_percent(&self, height: usize) -> String {
        if self.lines.is_empty() {
            return String::new();
        }
        let visible_height = height.saturating_sub(3);
        if self.lines.len() <= visible_height {
            return String::new();
        }
        let max_offset = self.lines.len().saturating_sub(visible_height);
        let percent = (self.offset * 100) / max_offset.max(1);
        format!("{}%", percent)
    }

    pub fn get_current_line_number(&self) -> Option<usize> {
        self.lines.get(self.cursor).and_then(|l| l.right_num.or(l.left_num))
    }
}

fn is_binary(content: &str) -> bool {
    let check_len = content.len().min(8192);
    content[..check_len].contains('\0')
}

fn parse_diff(content: &str) -> Vec<DiffLine> {
    let mut lines = vec![];
    let mut left_num = 0usize;
    let mut right_num = 0usize;

    for line in content.lines() {
        if line.starts_with("@@") {
            // Parse hunk header for line numbers
            if let Some((left, right)) = parse_hunk_header(line) {
                left_num = left;
                right_num = right;
            }
            lines.push(DiffLine {
                text: line.to_string(),
                line_type: LineType::Header,
                left_num: None,
                right_num: None,
            });
        } else if line.starts_with("diff --git") || line.starts_with("index ")
            || line.starts_with("---") || line.starts_with("+++")
            || line.starts_with("new file") || line.starts_with("deleted file")
        {
            lines.push(DiffLine {
                text: line.to_string(),
                line_type: LineType::Header,
                left_num: None,
                right_num: None,
            });
        } else if line.starts_with('+') {
            lines.push(DiffLine {
                text: line[1..].to_string(),
                line_type: LineType::Added,
                left_num: None,
                right_num: Some(right_num),
            });
            right_num += 1;
        } else if line.starts_with('-') {
            lines.push(DiffLine {
                text: line[1..].to_string(),
                line_type: LineType::Removed,
                left_num: Some(left_num),
                right_num: None,
            });
            left_num += 1;
        } else if line.starts_with(' ') {
            lines.push(DiffLine {
                text: line[1..].to_string(),
                line_type: LineType::Context,
                left_num: Some(left_num),
                right_num: Some(right_num),
            });
            left_num += 1;
            right_num += 1;
        } else {
            lines.push(DiffLine {
                text: line.to_string(),
                line_type: LineType::Context,
                left_num: None,
                right_num: None,
            });
        }
    }

    lines
}

fn parse_hunk_header(line: &str) -> Option<(usize, usize)> {
    // Parse @@ -start,count +start,count @@
    let parts: Vec<&str> = line.split_whitespace().collect();
    if parts.len() < 3 {
        return None;
    }

    let left_start = parts.get(1)?
        .trim_start_matches('-')
        .split(',')
        .next()?
        .parse()
        .ok()?;

    let right_start = parts.get(2)?
        .trim_start_matches('+')
        .split(',')
        .next()?
        .parse()
        .ok()?;

    Some((left_start, right_start))
}

fn parse_file_content(content: &str) -> Vec<DiffLine> {
    content
        .lines()
        .enumerate()
        .map(|(i, line)| DiffLine {
            text: line.to_string(),
            line_type: LineType::Context,
            left_num: Some(i + 1),
            right_num: Some(i + 1),
        })
        .collect()
}

fn parse_commit_summary(commit: &Commit, pr: Option<&PrInfo>) -> Vec<DiffLine> {
    let mut lines = vec![];

    // Commit info
    lines.push(DiffLine {
        text: "Commit".to_string(),
        line_type: LineType::Header,
        left_num: None,
        right_num: None,
    });
    lines.push(DiffLine {
        text: "─".repeat(40),
        line_type: LineType::Info,
        left_num: None,
        right_num: None,
    });
    lines.push(DiffLine {
        text: format!("Hash:   {}", commit.hash),
        line_type: LineType::Context,
        left_num: None,
        right_num: None,
    });
    lines.push(DiffLine {
        text: format!("Author: {}", commit.author),
        line_type: LineType::Context,
        left_num: None,
        right_num: None,
    });
    lines.push(DiffLine {
        text: format!("Date:   {}", commit.date),
        line_type: LineType::Context,
        left_num: None,
        right_num: None,
    });
    lines.push(DiffLine {
        text: String::new(),
        line_type: LineType::Context,
        left_num: None,
        right_num: None,
    });
    lines.push(DiffLine {
        text: commit.subject.clone(),
        line_type: LineType::Info,
        left_num: None,
        right_num: None,
    });
    lines.push(DiffLine {
        text: String::new(),
        line_type: LineType::Context,
        left_num: None,
        right_num: None,
    });

    // PR info
    if let Some(pr) = pr {
        lines.push(DiffLine {
            text: String::new(),
            line_type: LineType::Context,
            left_num: None,
            right_num: None,
        });
        lines.push(DiffLine {
            text: "Pull Request".to_string(),
            line_type: LineType::Header,
            left_num: None,
            right_num: None,
        });
        lines.push(DiffLine {
            text: "─".repeat(40),
            line_type: LineType::Info,
            left_num: None,
            right_num: None,
        });
        lines.push(DiffLine {
            text: pr.title.clone(),
            line_type: LineType::Info,
            left_num: None,
            right_num: None,
        });
        lines.push(DiffLine {
            text: format!("#{} by {} [{}]", pr.number, pr.author, pr.state),
            line_type: LineType::Context,
            left_num: None,
            right_num: None,
        });
        lines.push(DiffLine {
            text: pr.url.clone(),
            line_type: LineType::Context,
            left_num: None,
            right_num: None,
        });

        if !pr.body.is_empty() {
            lines.push(DiffLine {
                text: String::new(),
                line_type: LineType::Context,
                left_num: None,
                right_num: None,
            });
            for line in pr.body.lines() {
                lines.push(DiffLine {
                    text: line.to_string(),
                    line_type: LineType::Context,
                    left_num: None,
                    right_num: None,
                });
            }
        }

        // Reviews
        if !pr.reviews.is_empty() {
            lines.push(DiffLine {
                text: String::new(),
                line_type: LineType::Context,
                left_num: None,
                right_num: None,
            });
            lines.push(DiffLine {
                text: "Reviews".to_string(),
                line_type: LineType::Header,
                left_num: None,
                right_num: None,
            });
            for review in &pr.reviews {
                let state_type = match review.state.as_str() {
                    "APPROVED" => LineType::Added,
                    "CHANGES_REQUESTED" => LineType::Removed,
                    _ => LineType::Context,
                };
                lines.push(DiffLine {
                    text: format!("{} - {}", review.author, review.state),
                    line_type: state_type,
                    left_num: None,
                    right_num: None,
                });
                if !review.body.is_empty() {
                    for line in review.body.lines() {
                        lines.push(DiffLine {
                            text: format!("  {}", line),
                            line_type: LineType::Context,
                            left_num: None,
                            right_num: None,
                        });
                    }
                }
            }
        }
    } else {
        lines.push(DiffLine {
            text: String::new(),
            line_type: LineType::Context,
            left_num: None,
            right_num: None,
        });
        lines.push(DiffLine {
            text: "No PR found for this branch".to_string(),
            line_type: LineType::Info,
            left_num: None,
            right_num: None,
        });
    }

    lines
}

/// Diff view widget
pub struct DiffView<'a> {
    colors: &'a Colors,
    focused: bool,
}

impl<'a> DiffView<'a> {
    pub fn new(colors: &'a Colors) -> Self {
        Self {
            colors,
            focused: false,
        }
    }

    pub fn focused(mut self, focused: bool) -> Self {
        self.focused = focused;
        self
    }
}

impl<'a> StatefulWidget for DiffView<'a> {
    type State = DiffViewState;

    fn render(self, area: Rect, buf: &mut Buffer, state: &mut Self::State) {
        let border_style = if self.focused {
            self.colors.style_border_focused()
        } else {
            self.colors.style_border()
        };

        let scroll_info = state.scroll_percent(area.height as usize);
        let title = if scroll_info.is_empty() {
            state.title()
        } else {
            format!("{} ─── {}", state.title(), scroll_info)
        };

        let block = Block::default()
            .borders(Borders::ALL)
            .border_style(border_style)
            .title(Span::styled(title, self.colors.style_header()));

        let inner = block.inner(area);
        block.render(area, buf);

        if state.lines.is_empty() {
            let msg = match &state.content {
                PreviewContent::Empty => "Select a file to view",
                _ => "No content",
            };
            let line = Line::from(Span::styled(msg, self.colors.style_muted()));
            buf.set_line(inner.x, inner.y, &line, inner.width);
            return;
        }

        state.ensure_visible(inner.height as usize);

        let visible_lines: Vec<_> = state
            .lines
            .iter()
            .enumerate()
            .skip(state.offset)
            .take(inner.height as usize)
            .collect();

        for (i, (idx, diff_line)) in visible_lines.into_iter().enumerate() {
            let y = inner.y + i as u16;
            let is_cursor = self.focused && idx == state.cursor;
            let line = render_diff_line(diff_line, is_cursor, self.colors, inner.width as usize);
            buf.set_line(inner.x, y, &line, inner.width);
        }
    }
}

fn render_diff_line(diff_line: &DiffLine, cursor: bool, colors: &Colors, _width: usize) -> Line<'static> {
    let mut spans = vec![];

    // Line numbers
    let num_width = 4;
    if let Some(num) = diff_line.right_num.or(diff_line.left_num) {
        spans.push(Span::styled(
            format!("{:>width$} │ ", num, width = num_width),
            colors.style_muted(),
        ));
    } else {
        spans.push(Span::styled(
            format!("{:>width$} │ ", "", width = num_width),
            colors.style_muted(),
        ));
    }

    // Content
    let text = diff_line.text.replace('\t', "    ");
    let style = match diff_line.line_type {
        LineType::Added => colors.style_added(),
        LineType::Removed => colors.style_removed(),
        LineType::Header => colors.style_header(),
        LineType::Info => colors.style_muted(),
        LineType::Context => ratatui::style::Style::default().fg(colors.text),
    };

    let content_style = if cursor {
        style.add_modifier(ratatui::style::Modifier::REVERSED)
    } else {
        style
    };

    spans.push(Span::styled(text, content_style));

    Line::from(spans)
}
