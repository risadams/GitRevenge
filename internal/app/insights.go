package app

import (
	"context"
	"encoding/base64"
	"os"

	"github.com/risadams/gitrevenge/internal/git"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// RepoStats summarises the history authored between since and until (Unix seconds) on rev,
// or on every branch when rev is "", for the insights view.
func (a *App) RepoStats(path string, since, until int64, rev string) (git.Stats, error) {
	r, err := a.openRepo(path)
	if err != nil {
		return git.Stats{}, err
	}
	if until <= since {
		return git.Stats{}, userError("The start of the range must be before its end.")
	}
	return r.Git.Stats(context.Background(), git.StatsOptions{Since: since, Until: until, Rev: rev})
}

// BranchAges lists local and remote branches by when they last moved, oldest first.
func (a *App) BranchAges(path string) ([]git.BranchAge, error) {
	r, err := a.openRepo(path)
	if err != nil {
		return nil, err
	}
	return r.Git.BranchAges(context.Background())
}

// SaveFile asks where to save a file (an export) and writes the base64 content there. It
// returns the chosen path, or "" if the user cancelled.
func (a *App) SaveFile(title, defaultName, contentBase64 string) (string, error) {
	data, err := base64.StdEncoding.DecodeString(contentBase64)
	if err != nil {
		return "", err
	}
	path, err := runtime.SaveFileDialog(a.ctx, runtime.SaveDialogOptions{Title: title, DefaultFilename: defaultName})
	if err != nil || path == "" {
		return "", err
	}
	if err := os.WriteFile(path, data, 0o644); err != nil {
		return "", userError("Couldn't save " + path + ": " + err.Error())
	}
	return path, nil
}
