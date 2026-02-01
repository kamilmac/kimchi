package app

import "github.com/kmacinski/blocks/internal/git"

// AppMode represents the main application mode
type AppMode int

const (
	ModeChangedWorking AppMode = iota // Review uncommitted changes
	ModeChangedBranch                 // Review all changes vs base branch
	ModeBrowse                        // Browse all files
	ModeDocs                          // Browse documentation
)

// ModeCount is the total number of modes (for cycling)
const ModeCount = 4

// String returns the display name for the mode
func (m AppMode) String() string {
	switch m {
	case ModeChangedWorking:
		return "changed:working"
	case ModeChangedBranch:
		return "changed:branch"
	case ModeBrowse:
		return "browse"
	case ModeDocs:
		return "docs"
	default:
		return "unknown"
	}
}

// ShortString returns a short display name
func (m AppMode) ShortString() string {
	switch m {
	case ModeChangedWorking:
		return "working"
	case ModeChangedBranch:
		return "branch"
	case ModeBrowse:
		return "browse"
	case ModeDocs:
		return "docs"
	default:
		return "?"
	}
}

// IsChangedMode returns true if this is a changed files mode
func (m AppMode) IsChangedMode() bool {
	return m == ModeChangedWorking || m == ModeChangedBranch
}

// DiffMode returns the git diff mode for this app mode
func (m AppMode) DiffMode() git.DiffMode {
	if m == ModeChangedWorking {
		return git.DiffModeWorking
	}
	return git.DiffModeBranch
}

// FileViewMode returns the file view mode for this app mode
func (m AppMode) FileViewMode() git.FileViewMode {
	switch m {
	case ModeBrowse:
		return git.FileViewAll
	case ModeDocs:
		return git.FileViewDocs
	default:
		return git.FileViewChanged
	}
}

// Next returns the next mode (for cycling)
func (m AppMode) Next() AppMode {
	return AppMode((int(m) + 1) % ModeCount)
}

// SelectionType represents what kind of item is selected
type SelectionType int

const (
	SelectionNone SelectionType = iota
	SelectionFile
	SelectionFolder
	SelectionCommit
)

// Selection represents the current selection state
type Selection struct {
	Type       SelectionType
	FilePath   string
	FolderPath string
	Children   []string    // child paths for folder selection
	Commit     *git.Commit // selected commit
}

// Clear resets the selection
func (s *Selection) Clear() {
	s.Type = SelectionNone
	s.FilePath = ""
	s.FolderPath = ""
	s.Children = nil
	s.Commit = nil
}

// SelectFile sets file selection
func (s *Selection) SelectFile(path string) {
	s.Type = SelectionFile
	s.FilePath = path
	s.FolderPath = ""
	s.Children = nil
	s.Commit = nil
}

// SelectFolder sets folder selection
func (s *Selection) SelectFolder(path string, children []string) {
	s.Type = SelectionFolder
	s.FilePath = ""
	s.FolderPath = path
	s.Children = children
	s.Commit = nil
}

// SelectCommit sets commit selection
func (s *Selection) SelectCommit(commit *git.Commit) {
	s.Type = SelectionCommit
	s.FilePath = ""
	s.FolderPath = ""
	s.Children = nil
	s.Commit = commit
}
