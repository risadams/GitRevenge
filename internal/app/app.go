// Package app is the API surface bound into the Wails frontend.
package app

import (
	"context"
	"errors"
	"sync"

	"github.com/risadams/gitrevenge/internal/git"
	"github.com/risadams/gitrevenge/internal/version"
)

// App holds application-wide state. Its exported methods are callable from the frontend.
type App struct {
	ctx context.Context

	envOnce sync.Once
	env     Environment
}

// Environment reports the git installation the app will use.
type Environment struct {
	GitPath    string `json:"gitPath"`
	GitVersion string `json:"gitVersion"`
	LFSVersion string `json:"lfsVersion"`
	// Problem is a user-facing explanation when git is missing or too old; empty when usable.
	Problem    string `json:"problem"`
	MinVersion string `json:"minVersion"`
}

// New creates the App.
func New() *App {
	return &App{}
}

// Startup is called by Wails when the app starts. The context is kept for runtime calls.
func (a *App) Startup(ctx context.Context) {
	a.ctx = ctx
}

func discover(ctx context.Context) Environment {
	env := Environment{MinVersion: git.MinVersion.String()}
	in, err := git.Discover(ctx)
	if in != nil {
		env.GitPath = in.Path
		env.GitVersion = in.Version.String()
		env.LFSVersion = in.LFS
	}
	var tooOld *git.TooOldError
	switch {
	case err == nil:
	case errors.Is(err, git.ErrNotFound):
		env.Problem = "Git Revenge needs git, but no git installation was found."
	case errors.As(err, &tooOld):
		env.Problem = tooOld.Error()
	default:
		env.Problem = "Git could not be started: " + err.Error()
	}
	return env
}

// Version returns the application version.
func (a *App) Version() string {
	return version.Version
}

// Environment returns the git installation details. Discovery runs once, on first call,
// so it never races Wails' startup ordering.
func (a *App) Environment() Environment {
	a.envOnce.Do(func() { a.env = discover(context.Background()) })
	return a.env
}
