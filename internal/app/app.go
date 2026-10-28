// Package app is the API surface bound into the Wails frontend.
package app

import (
	"context"

	"github.com/risadams/gitrevenge/internal/version"
)

// App holds application-wide state. Its exported methods are callable from the frontend.
type App struct {
	ctx context.Context
}

// New creates the App.
func New() *App {
	return &App{}
}

// Startup is called by Wails when the app starts. The context is kept for runtime calls.
func (a *App) Startup(ctx context.Context) {
	a.ctx = ctx
}

// Version returns the application version.
func (a *App) Version() string {
	return version.Version
}
