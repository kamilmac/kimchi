package app

import (
	"github.com/kmacinski/blocks/internal/git"
	"github.com/kmacinski/blocks/internal/github"
)

// Selection messages

// FileSelectedMsg is sent when a file is selected
type FileSelectedMsg struct {
	Path string
}

// FolderSelectedMsg is sent when a folder is selected
type FolderSelectedMsg struct {
	Path     string
	Children []string
}

// CommitSelectedMsg is sent when a commit is selected
type CommitSelectedMsg struct {
	Commit git.Commit
}

// Data loaded messages

// FilesLoadedMsg is sent when files are loaded
type FilesLoadedMsg struct {
	Files []git.FileStatus
}

// ContentLoadedMsg is sent when diff or file content is loaded
type ContentLoadedMsg struct {
	Content string
}

// CommitsLoadedMsg is sent when commits are loaded
type CommitsLoadedMsg struct {
	Commits []git.Commit
}

// BranchInfoMsg is sent with branch information
type BranchInfoMsg struct {
	Branch     string
	BaseBranch string
}

// DiffStatsMsg is sent with diff statistics
type DiffStatsMsg struct {
	Added   int
	Removed int
}

// PRLoadedMsg is sent when PR info is loaded
type PRLoadedMsg struct {
	PR  *github.PRInfo
	Err error
}

// System messages

// ErrorMsg is sent when an error occurs
type ErrorMsg struct {
	Err error
}

// RefreshMsg triggers a refresh
type RefreshMsg struct{}

// GitChangedMsg is sent when git repository changes
type GitChangedMsg struct{}

// PRPollTickMsg triggers PR refresh
type PRPollTickMsg struct{}
