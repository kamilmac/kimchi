package watcher

import (
	"os"
	"path/filepath"
	"time"

	"github.com/fsnotify/fsnotify"
)

// GitWatcher watches for git repository changes
type GitWatcher struct {
	watcher  *fsnotify.Watcher
	onChange func()
	debounce time.Duration
	done     chan struct{}
}

// New creates a new git watcher
func New(debounce time.Duration, onChange func()) (*GitWatcher, error) {
	w, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}

	return &GitWatcher{
		watcher:  w,
		onChange: onChange,
		debounce: debounce,
		done:     make(chan struct{}),
	}, nil
}

// Start begins watching for changes
func (g *GitWatcher) Start() error {
	// Find git directory
	gitDir, err := findGitDir()
	if err != nil {
		return err
	}

	// Watch .git/index (staging changes)
	indexPath := filepath.Join(gitDir, "index")
	if _, err := os.Stat(indexPath); err == nil {
		g.watcher.Add(indexPath)
	}

	// Watch .git/HEAD (branch changes, commits)
	headPath := filepath.Join(gitDir, "HEAD")
	if _, err := os.Stat(headPath); err == nil {
		g.watcher.Add(headPath)
	}

	// Watch .git/refs/heads for new commits
	refsPath := filepath.Join(gitDir, "refs", "heads")
	if _, err := os.Stat(refsPath); err == nil {
		g.watcher.Add(refsPath)
	}

	// Watch current directory for file changes
	cwd, _ := os.Getwd()
	g.watchDirRecursive(cwd, gitDir)

	// Start event loop
	go g.eventLoop()

	return nil
}

func (g *GitWatcher) watchDirRecursive(dir, gitDir string) {
	filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}

		// Skip .git directory
		if info.IsDir() && info.Name() == ".git" {
			return filepath.SkipDir
		}

		// Skip hidden directories and common non-essential dirs
		if info.IsDir() && len(info.Name()) > 0 && info.Name()[0] == '.' {
			return filepath.SkipDir
		}
		if info.IsDir() && (info.Name() == "node_modules" || info.Name() == "vendor" || info.Name() == "__pycache__") {
			return filepath.SkipDir
		}

		// Watch directories
		if info.IsDir() {
			g.watcher.Add(path)
		}

		return nil
	})
}

func (g *GitWatcher) eventLoop() {
	var timer *time.Timer
	var timerC <-chan time.Time

	for {
		select {
		case <-g.done:
			if timer != nil {
				timer.Stop()
			}
			return

		case event, ok := <-g.watcher.Events:
			if !ok {
				return
			}

			// Ignore chmod events
			if event.Op == fsnotify.Chmod {
				continue
			}

			// Debounce: reset timer on each event
			if timer != nil {
				timer.Stop()
			}
			timer = time.NewTimer(g.debounce)
			timerC = timer.C

		case <-timerC:
			// Debounce period passed, trigger callback
			if g.onChange != nil {
				g.onChange()
			}
			timerC = nil

		case _, ok := <-g.watcher.Errors:
			if !ok {
				return
			}
			// Ignore errors, just continue
		}
	}
}

// Stop stops the watcher
func (g *GitWatcher) Stop() {
	close(g.done)
	g.watcher.Close()
}

func findGitDir() (string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return "", err
	}

	dir := cwd
	for {
		gitPath := filepath.Join(dir, ".git")
		if info, err := os.Stat(gitPath); err == nil {
			if info.IsDir() {
				return gitPath, nil
			}
			// Handle git worktrees (file pointing to actual git dir)
			// For simplicity, just return the path
			return gitPath, nil
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			return "", os.ErrNotExist
		}
		dir = parent
	}
}
