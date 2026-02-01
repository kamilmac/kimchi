package app

import (
	"github.com/kmacinski/blocks/internal/config"
	"github.com/kmacinski/blocks/internal/git"
	"github.com/kmacinski/blocks/internal/github"
)

// State holds the shared application state
type State struct {
	// Mode and selection
	Mode      AppMode
	Selection Selection

	// File list data
	Files         []git.FileStatus
	SelectedIndex int

	// Cached content
	DiffContent string

	// Branch info
	Branch     string
	BaseBranch string

	// Stats
	DiffAdded   int
	DiffRemoved int

	// UI state
	FocusedWindow string
	ActiveModal   string

	// PR data
	PR *github.PRInfo

	// Errors
	Error string
}

// NewState creates a new state with defaults
func NewState() *State {
	return &State{
		Mode:          ModeChangedBranch,
		FocusedWindow: config.WindowFileList,
	}
}

// SetMode changes the app mode and resets selection
func (s *State) SetMode(mode AppMode) {
	s.Mode = mode
	s.Selection.Clear()
	s.SelectedIndex = 0
}

// CycleMode advances to the next mode
func (s *State) CycleMode() {
	s.SetMode(s.Mode.Next())
}

// SetFiles updates the file list
func (s *State) SetFiles(files []git.FileStatus) {
	s.Files = files
	if s.SelectedIndex >= len(files) {
		s.SelectedIndex = 0
	}
}

// ToggleModal toggles a modal on/off
func (s *State) ToggleModal(name string) {
	if s.ActiveModal == name {
		s.ActiveModal = ""
	} else {
		s.ActiveModal = name
	}
}

// CloseModal closes any open modal
func (s *State) CloseModal() {
	s.ActiveModal = ""
}

// CycleWindow cycles focus to the next window
func (s *State) CycleWindow(windows []string, reverse bool) {
	if len(windows) == 0 {
		return
	}
	currentIdx := 0
	for i, w := range windows {
		if w == s.FocusedWindow {
			currentIdx = i
			break
		}
	}
	if reverse {
		currentIdx = (currentIdx - 1 + len(windows)) % len(windows)
	} else {
		currentIdx = (currentIdx + 1) % len(windows)
	}
	s.FocusedWindow = windows[currentIdx]
}
