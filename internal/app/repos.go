package app

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"time"

	"github.com/wailsapp/wails/v2/pkg/runtime"

	"github.com/risadams/gitrevenge/internal/git"
	"github.com/risadams/gitrevenge/internal/session"
	"github.com/risadams/gitrevenge/internal/settings"
)

// RepoTab describes an open repository for the tab bar.
type RepoTab struct {
	Path string `json:"path"`
	Name string `json:"name"`
	// Head is the current branch, "detached at <sha>", or empty for a repo with no commits.
	Head string `json:"head"`
}

// SessionState is what the UI needs to rebuild its tabs on launch.
type SessionState struct {
	Tabs   []RepoTab    `json:"tabs"`
	Active string       `json:"active"`
	Failed []FailedRepo `json:"failed"`
}

// FailedRepo is a saved tab that could not be reopened.
type FailedRepo struct {
	Path  string `json:"path"`
	Error string `json:"error"`
}

// RecentRepo is an entry in the start page's recent list.
type RecentRepo struct {
	Path       string    `json:"path"`
	Name       string    `json:"name"`
	LastOpened time.Time `json:"lastOpened"`
	// Exists is false when the folder is gone; the UI greys it out.
	Exists bool `json:"exists"`
}

// sessions returns the session manager, creating it on first use once git is known.
func (a *App) sessions() (*session.Manager, error) {
	a.sessOnce.Do(func() {
		env := a.Environment()
		if env.Problem != "" {
			a.sessErr = errors.New(env.Problem)
			return
		}
		path := a.settingsPath
		if path == "" {
			if path, a.sessErr = settings.DefaultPath(); a.sessErr != nil {
				return
			}
		}
		store, err := settings.Open(path)
		if err != nil {
			a.sessErr = err
			return
		}
		a.sess = session.NewManager(git.NewRunner(env.GitPath), store)
	})
	return a.sess, a.sessErr
}

// RestoreSession reopens the tabs from the previous run.
func (a *App) RestoreSession() (SessionState, error) {
	m, err := a.sessions()
	if err != nil {
		return SessionState{}, err
	}
	ctx := context.Background()
	active, failed := m.Restore(ctx)
	st := SessionState{Active: active, Tabs: []RepoTab{}, Failed: []FailedRepo{}}
	for _, r := range m.List() {
		st.Tabs = append(st.Tabs, tabFor(ctx, r))
	}
	for p, err := range failed {
		st.Failed = append(st.Failed, FailedRepo{Path: p, Error: err.Error()})
	}
	return st, nil
}

// ChooseAndOpenRepo shows a folder picker and opens the chosen repository. It returns nil
// if the user cancels.
func (a *App) ChooseAndOpenRepo() (*RepoTab, error) {
	dir, err := runtime.OpenDirectoryDialog(a.ctx, runtime.OpenDialogOptions{Title: "Open repository"})
	if err != nil || dir == "" {
		return nil, err
	}
	tab, err := a.OpenRepo(dir)
	if err != nil {
		return nil, err
	}
	return &tab, nil
}

// OpenRepo opens the repository containing path as a tab (or focuses it if already open).
func (a *App) OpenRepo(path string) (RepoTab, error) {
	m, err := a.sessions()
	if err != nil {
		return RepoTab{}, err
	}
	ctx := context.Background()
	r, err := m.Open(ctx, path)
	if err != nil {
		return RepoTab{}, friendly(err)
	}
	return tabFor(ctx, r), nil
}

// CloseRepo closes the tab for path.
func (a *App) CloseRepo(path string) error {
	m, err := a.sessions()
	if err != nil {
		return err
	}
	return m.Close(path)
}

// SetActiveRepo records the selected tab; an empty path means the start page.
func (a *App) SetActiveRepo(path string) error {
	m, err := a.sessions()
	if err != nil {
		return err
	}
	return m.SetActive(path)
}

// RecentRepos lists recently opened repositories, most recent first.
func (a *App) RecentRepos() ([]RecentRepo, error) {
	m, err := a.sessions()
	if err != nil {
		return nil, err
	}
	recent := m.Settings().Get().Recent
	out := make([]RecentRepo, 0, len(recent))
	for _, r := range recent {
		_, statErr := os.Stat(r.Path)
		out = append(out, RecentRepo{Path: r.Path, Name: filepath.Base(r.Path), LastOpened: r.LastOpened, Exists: statErr == nil})
	}
	return out, nil
}

// RemoveRecentRepo drops path from the recent list.
func (a *App) RemoveRecentRepo(path string) error {
	m, err := a.sessions()
	if err != nil {
		return err
	}
	return m.Settings().Update(func(s *settings.Settings) { s.RemoveRecent(path) })
}

func tabFor(ctx context.Context, r *session.Repo) RepoTab {
	return RepoTab{Path: r.Path, Name: r.Name(), Head: head(ctx, r.Git)}
}

// head describes HEAD for display. Errors yield an empty string rather than failing the tab.
func head(ctx context.Context, r *git.Repo) string {
	if branch, err := r.ReadString(ctx, "symbolic-ref", "--short", "-q", "HEAD"); err == nil {
		return branch
	}
	if sha, err := r.ReadString(ctx, "rev-parse", "--short", "HEAD"); err == nil {
		return "detached at " + sha
	}
	return ""
}

// userError is an error whose message is written to be shown to the user as is.
type userError string

func (e userError) Error() string { return string(e) }

// friendly rewrites well-known errors into messages fit for a dialog.
func friendly(err error) error {
	switch {
	case errors.Is(err, session.ErrNotRepo):
		return userError("That folder isn't a git repository.")
	case errors.Is(err, session.ErrBare):
		return userError("Bare repositories aren't supported, because they have no working files to show.")
	}
	return err
}
