package app

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/atotto/clipboard"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/kmacinski/blocks/internal/config"
	"github.com/kmacinski/blocks/internal/git"
	"github.com/kmacinski/blocks/internal/github"
	"github.com/kmacinski/blocks/internal/layout"
	"github.com/kmacinski/blocks/internal/watcher"
	"github.com/kmacinski/blocks/internal/window"
)

// App is the main application model
type App struct {
	state  *State
	git    git.Client
	gh     github.Client
	layout *layout.Manager
	styles config.Styles

	// Windows
	fileList   *window.FileList
	commitList *window.CommitList
	diffView   *window.DiffView
	fileView   *window.FileView
	help       *window.Help

	// Window registry
	windows     map[string]window.Window
	assignments map[string]string

	// Dimensions
	width  int
	height int

	// Status
	statusMessage string

	// File watcher
	watcher *watcher.GitWatcher
	program *tea.Program
}

// New creates a new application
func New(gitClient git.Client) *App {
	styles := config.DefaultStyles
	state := NewState()

	fileList := window.NewFileList(styles)
	commitList := window.NewCommitList(styles)
	diffView := window.NewDiffView(styles)
	fileView := window.NewFileView(styles)
	help := window.NewHelp(styles)

	fileList.SetFocus(true)

	windows := map[string]window.Window{
		config.WindowFileList:   fileList,
		config.WindowCommitList: commitList,
		config.WindowDiffView:   diffView,
		config.WindowFileView:   fileView,
		config.WindowHelp:       help,
	}

	assignments := map[string]string{
		"left-top":    config.WindowFileList,
		"left-bottom": config.WindowCommitList,
		"right":       config.WindowDiffView,
		"top":         config.WindowFileList,
		"middle":      config.WindowCommitList,
		"bottom":      config.WindowDiffView,
	}

	app := &App{
		state:       state,
		git:         gitClient,
		gh:          github.NewClient(),
		layout:      layout.NewManager(layout.DefaultResponsive),
		styles:      styles,
		fileList:    fileList,
		commitList:  commitList,
		diffView:    diffView,
		fileView:    fileView,
		help:        help,
		windows:     windows,
		assignments: assignments,
	}

	// Set selection callbacks
	fileList.SetOnSelect(func(index int, path string) tea.Cmd {
		return func() tea.Msg {
			if fileList.IsFolderSelected() {
				return FolderSelectedMsg{
					Path:     path,
					Children: fileList.SelectedChildren(),
				}
			}
			return FileSelectedMsg{Path: path}
		}
	})

	commitList.SetOnSelect(func(commit git.Commit) tea.Cmd {
		return func() tea.Msg {
			return CommitSelectedMsg{Commit: commit}
		}
	})

	return app
}

// SetProgram sets the tea.Program reference
func (a *App) SetProgram(p *tea.Program) {
	a.program = p
	w, err := watcher.New(config.FileWatcherDebounce, func() {
		if a.program != nil {
			a.program.Send(GitChangedMsg{})
		}
	})
	if err == nil {
		a.watcher = w
		a.watcher.Start()
	}
}

// Cleanup stops the watcher
func (a *App) Cleanup() {
	if a.watcher != nil {
		a.watcher.Stop()
	}
}

// Init initializes the application
func (a *App) Init() tea.Cmd {
	return tea.Batch(
		a.loadBranchInfo(),
		a.loadFiles(),
		a.loadCommits(),
		a.loadDiffStats(),
		a.loadPR(),
		a.schedulePRPoll(),
	)
}

func (a *App) schedulePRPoll() tea.Cmd {
	return tea.Tick(config.PRPollInterval, func(t time.Time) tea.Msg {
		return PRPollTickMsg{}
	})
}

// Update handles messages
func (a *App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		a.width = msg.Width
		a.height = msg.Height
		a.layout.Resize(msg.Width, msg.Height)
		return a, nil

	case tea.KeyMsg:
		if a.state.ActiveModal != "" {
			return a.handleModalKey(msg)
		}
		return a.handleKey(msg)

	// Selection messages
	case FileSelectedMsg:
		a.state.Selection.SelectFile(msg.Path)
		return a, a.loadContent()

	case FolderSelectedMsg:
		a.state.Selection.SelectFolder(msg.Path, msg.Children)
		return a, a.loadContent()

	case CommitSelectedMsg:
		a.state.Selection.SelectCommit(&msg.Commit)
		a.updatePreview()
		return a, nil

	// Data messages
	case FilesLoadedMsg:
		a.state.SetFiles(msg.Files)
		a.fileList.SetFiles(msg.Files)
		return a, nil

	case ContentLoadedMsg:
		a.state.DiffContent = msg.Content
		a.updatePreview()
		return a, nil

	case CommitsLoadedMsg:
		a.commitList.SetCommits(msg.Commits)
		return a, nil

	case BranchInfoMsg:
		a.state.Branch = msg.Branch
		a.state.BaseBranch = msg.BaseBranch
		return a, nil

	case DiffStatsMsg:
		a.state.DiffAdded = msg.Added
		a.state.DiffRemoved = msg.Removed
		return a, nil

	case PRLoadedMsg:
		if msg.Err == nil {
			a.state.PR = msg.PR
		} else {
			a.state.PR = nil
		}
		a.diffView.SetPR(a.state.PR)
		a.fileList.SetPR(a.state.PR)
		return a, nil

	case ErrorMsg:
		a.state.Error = msg.Err.Error()
		return a, nil

	case GitChangedMsg:
		return a, tea.Batch(
			a.loadBranchInfo(),
			a.loadFiles(),
			a.loadCommits(),
			a.loadContent(),
			a.loadDiffStats(),
			a.loadPR(),
		)

	case PRPollTickMsg:
		return a, tea.Batch(a.loadPR(), a.schedulePRPoll())
	}

	return a, nil
}

func (a *App) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, config.DefaultKeyMap.Quit):
		return a, tea.Quit

	case key.Matches(msg, config.DefaultKeyMap.Help):
		a.state.ToggleModal(config.ModalHelp)
		return a, nil

	case key.Matches(msg, config.DefaultKeyMap.Refresh):
		return a, tea.Batch(a.loadFiles(), a.loadContent(), a.loadDiffStats())

	case key.Matches(msg, config.DefaultKeyMap.CycleMode):
		a.state.CycleMode()
		a.applyModeChange()
		return a, a.loadFiles()

	case key.Matches(msg, config.DefaultKeyMap.Mode1):
		a.state.SetMode(ModeChangedWorking)
		a.applyModeChange()
		return a, a.loadFiles()

	case key.Matches(msg, config.DefaultKeyMap.Mode2):
		a.state.SetMode(ModeChangedBranch)
		a.applyModeChange()
		return a, a.loadFiles()

	case key.Matches(msg, config.DefaultKeyMap.Mode3):
		a.state.SetMode(ModeBrowse)
		a.applyModeChange()
		return a, a.loadFiles()

	case key.Matches(msg, config.DefaultKeyMap.Mode4):
		a.state.SetMode(ModeDocs)
		a.applyModeChange()
		return a, a.loadFiles()

	case key.Matches(msg, config.DefaultKeyMap.Tab):
		a.cycleFocus(true)
		return a, nil

	case key.Matches(msg, config.DefaultKeyMap.ShiftTab):
		a.cycleFocus(false)
		return a, nil

	case key.Matches(msg, config.DefaultKeyMap.Yank):
		return a, a.handleYank()

	case key.Matches(msg, config.DefaultKeyMap.OpenEditor):
		return a, a.handleOpenEditor()
	}

	return a.delegateToFocused(msg)
}

func (a *App) handleModalKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if key.Matches(msg, config.DefaultKeyMap.Quit) {
		return a, tea.Quit
	}
	if key.Matches(msg, config.DefaultKeyMap.Help) || key.Matches(msg, config.DefaultKeyMap.Escape) {
		a.state.CloseModal()
		return a, nil
	}
	return a, nil
}

func (a *App) delegateToFocused(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch a.state.FocusedWindow {
	case config.WindowFileList:
		_, cmd = a.fileList.Update(msg)
	case config.WindowCommitList:
		_, cmd = a.commitList.Update(msg)
	case config.WindowDiffView:
		_, cmd = a.diffView.Update(msg)
	case config.WindowFileView:
		_, cmd = a.fileView.Update(msg)
	}
	return a, cmd
}

func (a *App) cycleFocus(reverse bool) {
	previewWindow := a.getPreviewWindow()
	windowOrder := []string{config.WindowFileList, config.WindowCommitList, previewWindow}

	prevWindow := a.state.FocusedWindow
	a.state.CycleWindow(windowOrder, reverse)
	a.updateFocus()

	// Update preview based on focus change
	if a.state.FocusedWindow == config.WindowCommitList && prevWindow != config.WindowCommitList {
		// Switching to commit list - show commit summary
		commit := a.commitList.SelectedCommit()
		if commit != nil {
			a.state.Selection.SelectCommit(commit)
			a.updatePreview()
		}
	} else if prevWindow == config.WindowCommitList && a.state.FocusedWindow != config.WindowCommitList {
		// Leaving commit list - restore file selection if any
		a.state.Selection.Clear()
		a.updatePreview()
	}
}

func (a *App) updateFocus() {
	a.fileList.SetFocus(a.state.FocusedWindow == config.WindowFileList)
	a.commitList.SetFocus(a.state.FocusedWindow == config.WindowCommitList)
	a.diffView.SetFocus(a.state.FocusedWindow == config.WindowDiffView)
	a.fileView.SetFocus(a.state.FocusedWindow == config.WindowFileView)
}

func (a *App) getPreviewWindow() string {
	if a.state.Mode == ModeBrowse {
		return config.WindowFileView
	}
	return config.WindowDiffView
}

func (a *App) applyModeChange() {
	a.fileList.SetViewMode(a.state.Mode.FileViewMode())

	// Update preview window assignment
	previewWindow := a.getPreviewWindow()
	a.assignments["right"] = previewWindow
	a.assignments["bottom"] = previewWindow

	// Reset focus
	a.state.FocusedWindow = config.WindowFileList
	a.updateFocus()
}

func (a *App) updatePreview() {
	preview := a.computePreview()

	if a.state.Mode == ModeBrowse && preview.Type == window.PreviewFileContent {
		a.fileView.SetContent(preview.Content, preview.FilePath)
	} else {
		a.diffView.SetPreview(preview)
	}
}

func (a *App) computePreview() window.PreviewContent {
	switch a.state.Selection.Type {
	case SelectionCommit:
		return window.PreviewContent{
			Type:   window.PreviewCommitSummary,
			Commit: a.state.Selection.Commit,
			PR:     a.state.PR,
		}

	case SelectionFile:
		if a.state.Mode == ModeBrowse {
			return window.PreviewContent{
				Type:     window.PreviewFileContent,
				Content:  a.state.DiffContent,
				FilePath: a.state.Selection.FilePath,
			}
		}
		return window.PreviewContent{
			Type:     window.PreviewFileDiff,
			Content:  a.state.DiffContent,
			FilePath: a.state.Selection.FilePath,
		}

	case SelectionFolder:
		return window.PreviewContent{
			Type:       window.PreviewFolderDiff,
			Content:    a.state.DiffContent,
			FolderPath: a.state.Selection.FolderPath,
		}

	default:
		return window.PreviewContent{Type: window.PreviewEmpty}
	}
}

func (a *App) handleYank() tea.Cmd {
	var toCopy string

	if a.state.FocusedWindow == config.WindowDiffView {
		filePath, lineNum := a.diffView.GetSelectedLocation()
		if filePath != "" && lineNum > 0 {
			toCopy = fmt.Sprintf("%s:%d", filePath, lineNum)
		} else if filePath != "" {
			toCopy = filePath
		}
	} else if a.state.FocusedWindow == config.WindowFileView {
		filePath := a.fileView.GetFilePath()
		lineNum := a.fileView.GetSelectedLine()
		if filePath != "" && lineNum > 0 {
			toCopy = fmt.Sprintf("%s:%d", filePath, lineNum)
		} else if filePath != "" {
			toCopy = filePath
		}
	} else if a.state.Selection.FilePath != "" {
		toCopy = a.state.Selection.FilePath
	}

	if toCopy != "" {
		if err := clipboard.WriteAll(toCopy); err == nil {
			a.statusMessage = fmt.Sprintf("Copied: %s", toCopy)
		}
	}
	return nil
}

func (a *App) handleOpenEditor() tea.Cmd {
	var filePath string
	var lineNum int

	if a.state.FocusedWindow == config.WindowDiffView {
		filePath, lineNum = a.diffView.GetSelectedLocation()
	} else if a.state.FocusedWindow == config.WindowFileView {
		filePath = a.fileView.GetFilePath()
		lineNum = a.fileView.GetSelectedLine()
	} else {
		filePath = a.state.Selection.FilePath
		lineNum = 1
	}

	if filePath == "" {
		return nil
	}

	editor := os.Getenv("EDITOR")
	if editor == "" {
		editor = "vim"
	}

	var c *exec.Cmd
	if lineNum > 1 {
		c = exec.Command(editor, fmt.Sprintf("+%d", lineNum), filePath)
	} else {
		c = exec.Command(editor, filePath)
	}
	return tea.ExecProcess(c, func(err error) tea.Msg {
		return RefreshMsg{}
	})
}

// View renders the application
func (a *App) View() string {
	if a.width == 0 || a.height == 0 {
		return "Loading..."
	}

	if !a.git.IsRepo() {
		return a.renderError("Not a git repository", "Run blocks from within a git repository")
	}

	statusBar := a.renderStatusBar()
	mainView := a.layout.Render(a.windows, a.assignments, statusBar)

	if a.state.ActiveModal == config.ModalHelp {
		mainView = a.renderWithModal(mainView, a.help)
	}

	return mainView
}

func (a *App) renderStatusBar() string {
	branch := a.state.Branch
	if branch == "" {
		branch = "unknown"
	}

	mode := fmt.Sprintf("[%s]", a.state.Mode.String())
	fileCount := fmt.Sprintf("%d files", len(a.state.Files))

	// Diff stats with background
	stats := ""
	if a.state.DiffAdded > 0 || a.state.DiffRemoved > 0 {
		bg := a.styles.StatusBar.GetBackground()
		addedStyle := lipgloss.NewStyle().Foreground(a.styles.DiffAdded.GetForeground()).Background(bg)
		removedStyle := lipgloss.NewStyle().Foreground(a.styles.DiffRemoved.GetForeground()).Background(bg)
		stats = fmt.Sprintf("%s %s",
			addedStyle.Render(fmt.Sprintf("+%d", a.state.DiffAdded)),
			removedStyle.Render(fmt.Sprintf("-%d", a.state.DiffRemoved)),
		)
	}

	// PR info
	prInfo := ""
	if a.state.PR != nil {
		commentCount := len(a.state.PR.Comments) + len(a.state.PR.ReviewComments)
		if commentCount > 0 {
			prInfo = fmt.Sprintf("%d💬", commentCount)
		}
	}

	// Status message
	statusMsg := ""
	if a.statusMessage != "" {
		bg := a.styles.StatusBar.GetBackground()
		mutedStyle := lipgloss.NewStyle().Foreground(a.styles.Muted.GetForeground()).Background(bg)
		statusMsg = mutedStyle.Render(" │ " + a.statusMessage)
		a.statusMessage = ""
	}

	// Build line
	left := fmt.Sprintf("%s  %s  %s", branch, mode, fileCount)
	if stats != "" {
		left += "  " + stats
	}
	if prInfo != "" {
		left += "  " + prInfo
	}
	left += statusMsg

	right := "[?]"

	leftWidth := lipgloss.Width(left)
	rightWidth := lipgloss.Width(right)
	available := a.width - 2
	padding := available - leftWidth - rightWidth
	if padding < 1 {
		padding = 1
	}

	bg := a.styles.StatusBar.GetBackground()
	mutedWithBg := lipgloss.NewStyle().Foreground(a.styles.Muted.GetForeground()).Background(bg)
	line := " " + left + strings.Repeat(" ", padding) + mutedWithBg.Render(right) + " "

	return lipgloss.NewStyle().
		Background(a.styles.StatusBar.GetBackground()).
		Foreground(a.styles.StatusBar.GetForeground()).
		Width(a.width).
		Render(line)
}

func (a *App) renderWithModal(background string, modal window.Window) string {
	modalWidth := min(config.ModalMaxWidth, a.width-config.ModalPadding)
	modalHeight := min(config.ModalMaxHeight, a.height-config.ModalPadding)
	modalContent := modal.View(modalWidth, modalHeight)
	return lipgloss.Place(a.width, a.height, lipgloss.Center, lipgloss.Center, modalContent)
}

func (a *App) renderError(title, hint string) string {
	style := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#f38ba8")).
		Bold(true).
		Padding(2)
	content := fmt.Sprintf("%s\n\n%s", title, a.styles.Muted.Render(hint))
	return lipgloss.Place(a.width, a.height, lipgloss.Center, lipgloss.Center, style.Render(content))
}

// Data loading commands

func (a *App) loadBranchInfo() tea.Cmd {
	return func() tea.Msg {
		branch, _ := a.git.CurrentBranch()
		baseBranch, _ := a.git.BaseBranch()
		return BranchInfoMsg{Branch: branch, BaseBranch: baseBranch}
	}
}

func (a *App) loadFiles() tea.Cmd {
	return func() tea.Msg {
		var files []git.FileStatus
		var err error

		switch a.state.Mode.FileViewMode() {
		case git.FileViewAll:
			files, err = a.git.ListAllFiles()
		case git.FileViewDocs:
			files, err = a.git.ListDocFiles()
		default:
			files, err = a.git.Status(a.state.Mode.DiffMode())
		}

		if err != nil {
			return ErrorMsg{Err: err}
		}
		return FilesLoadedMsg{Files: files}
	}
}

func (a *App) loadCommits() tea.Cmd {
	return func() tea.Msg {
		commits, err := a.git.Log()
		if err != nil {
			return CommitsLoadedMsg{Commits: nil}
		}
		if len(commits) > 8 {
			commits = commits[:8]
		}
		return CommitsLoadedMsg{Commits: commits}
	}
}

func (a *App) loadContent() tea.Cmd {
	return func() tea.Msg {
		switch a.state.Selection.Type {
		case SelectionFile:
			if a.state.Mode == ModeBrowse {
				content, err := a.git.ReadFile(a.state.Selection.FilePath)
				if err != nil {
					return ErrorMsg{Err: err}
				}
				return ContentLoadedMsg{Content: content}
			}
			content, err := a.git.Diff(a.state.Selection.FilePath, a.state.Mode.DiffMode())
			if err != nil {
				return ErrorMsg{Err: err}
			}
			return ContentLoadedMsg{Content: content}

		case SelectionFolder:
			var combined strings.Builder
			for _, path := range a.state.Selection.Children {
				diff, err := a.git.Diff(path, a.state.Mode.DiffMode())
				if err == nil && diff != "" {
					combined.WriteString(diff)
					combined.WriteString("\n")
				}
			}
			return ContentLoadedMsg{Content: combined.String()}

		default:
			return ContentLoadedMsg{Content: ""}
		}
	}
}

func (a *App) loadDiffStats() tea.Cmd {
	return func() tea.Msg {
		added, removed, err := a.git.DiffStats(a.state.Mode.DiffMode())
		if err != nil {
			return DiffStatsMsg{Added: 0, Removed: 0}
		}
		return DiffStatsMsg{Added: added, Removed: removed}
	}
}

func (a *App) loadPR() tea.Cmd {
	return func() tea.Msg {
		pr, err := a.gh.GetPRForBranch()
		return PRLoadedMsg{PR: pr, Err: err}
	}
}
