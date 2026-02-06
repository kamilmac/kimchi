use std::fmt;

/// File status in git
#[derive(Debug, Clone, Copy, PartialEq, Eq)]
pub enum FileStatus {
    Modified,
    Added,
    Deleted,
    Renamed,
    Unchanged,
}

impl FileStatus {
    pub fn as_char(&self) -> char {
        match self {
            Self::Modified => 'M',
            Self::Added => 'A',
            Self::Deleted => 'D',
            Self::Renamed => 'R',
            Self::Unchanged => ' ',
        }
    }
}

impl fmt::Display for FileStatus {
    fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
        write!(f, "{}", self.as_char())
    }
}

/// A file with its status
#[derive(Debug, Clone)]
pub struct StatusEntry {
    pub path: String,
    pub status: FileStatus,
    /// True if file has uncommitted changes
    pub uncommitted: bool,
}


/// Diff statistics
#[derive(Debug, Clone, Default)]
pub struct DiffStats {
    pub added: usize,
    pub removed: usize,
}

/// Blame information for a single line
#[derive(Debug, Clone)]
pub struct BlameInfo {
    pub line: usize,
    pub author: String,
    pub date: String,
    pub commit_id: String,
    pub summary: String,
}

/// Blame data for a file
#[derive(Debug, Clone, Default)]
pub struct FileBlame {
    pub path: String,
    pub lines: Vec<BlameInfo>,
}

/// Timeline position for viewing PR history
/// Order: Browse → Wip → FullDiff → -1 → -2 → ... → -16
/// FullDiff is the default (primary code review view)
#[derive(Debug, Clone, Copy, PartialEq, Eq, Default)]
pub enum TimelinePosition {
    /// Browse full file content (no diff)
    Browse,
    /// View only uncommitted changes: HEAD → working tree
    Wip,
    /// View all committed changes: base → HEAD (default)
    #[default]
    FullDiff,
    /// View changes from a single commit: HEAD~N → HEAD~(N-1)
    CommitDiff(usize),
}

impl TimelinePosition {
    /// Move to next position (towards older commits: Browse → Wip → FullDiff → -1 → -2 → ...)
    pub fn next(self, max_commits: usize) -> Self {
        match self {
            Self::Browse => Self::Wip,
            Self::Wip => Self::FullDiff,
            Self::FullDiff => {
                if max_commits > 0 {
                    Self::CommitDiff(1)
                } else {
                    Self::FullDiff
                }
            }
            Self::CommitDiff(n) if n < max_commits && n < 16 => Self::CommitDiff(n + 1),
            other => other,
        }
    }

    /// Move to previous position (towards newer: ... → -1 → FullDiff → Wip → Browse)
    pub fn prev(self) -> Self {
        match self {
            Self::Browse => Self::Browse, // Can't go newer than browse
            Self::Wip => Self::Browse,
            Self::FullDiff => Self::Wip,
            Self::CommitDiff(1) => Self::FullDiff,
            Self::CommitDiff(n) => Self::CommitDiff(n - 1),
        }
    }
}
