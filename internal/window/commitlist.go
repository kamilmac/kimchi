package window

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/kmacinski/blocks/internal/config"
	"github.com/kmacinski/blocks/internal/git"
)

// CommitList displays recent commits
type CommitList struct {
	Base
	commits []git.Commit
	cursor  int
	width   int
	height  int
}

// NewCommitList creates a new commit list window
func NewCommitList(styles config.Styles) *CommitList {
	return &CommitList{
		Base: NewBase("commitlist", styles),
	}
}

// SetCommits sets the commits to display
func (c *CommitList) SetCommits(commits []git.Commit) {
	c.commits = commits
	if c.cursor >= len(commits) {
		c.cursor = 0
	}
}

// Update handles input
func (c *CommitList) Update(msg tea.Msg) (Window, tea.Cmd) {
	// Commits window is read-only for now
	return c, nil
}

// View renders the commit list
func (c *CommitList) View(width, height int) string {
	c.width = width
	c.height = height

	// Account for border
	contentWidth := width - 2
	contentHeight := height - 2

	if contentWidth < 1 || contentHeight < 1 {
		return ""
	}

	var lines []string

	if len(c.commits) == 0 {
		lines = append(lines, c.styles.Muted.Render("No commits"))
	} else {
		// Show commits that fit
		maxCommits := contentHeight
		if maxCommits > len(c.commits) {
			maxCommits = len(c.commits)
		}

		for i := 0; i < maxCommits; i++ {
			commit := c.commits[i]
			line := c.formatCommit(commit, contentWidth)
			lines = append(lines, line)
		}
	}

	// Pad to fill height
	for len(lines) < contentHeight {
		lines = append(lines, "")
	}

	content := strings.Join(lines[:contentHeight], "\n")

	// Apply window style
	style := c.styles.WindowUnfocused
	if c.focused {
		style = c.styles.WindowFocused
	}

	return style.
		Width(contentWidth).
		Height(contentHeight).
		Render(content)
}

// formatCommit formats a single commit line
func (c *CommitList) formatCommit(commit git.Commit, width int) string {
	// Format: hash subject
	hash := c.styles.Muted.Render(commit.Hash[:7])

	// Calculate available space for subject
	hashWidth := 8 // 7 chars + space
	subjectWidth := width - hashWidth
	if subjectWidth < 10 {
		subjectWidth = 10
	}

	subject := commit.Subject
	if lipgloss.Width(subject) > subjectWidth {
		subject = subject[:subjectWidth-3] + "..."
	}

	return fmt.Sprintf("%s %s", hash, subject)
}
