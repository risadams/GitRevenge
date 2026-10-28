package session

import (
	"context"
	"sync"
	"time"

	"github.com/risadams/gitrevenge/internal/git"
	"github.com/risadams/gitrevenge/internal/graph"
)

// GraphProgress reports loading progress for a repository's commit graph.
type GraphProgress struct {
	Rows int
	Done bool
	Err  error
}

// graphLoader owns a repository's commit graph and reloads it on demand.
type graphLoader struct {
	repo       *git.Repo
	onProgress func(GraphProgress)

	mu      sync.Mutex
	current *graph.Graph
	loading bool
	cancel  context.CancelFunc
}

// Graph returns the graph to display and whether a load is in progress. The graph may be
// partially loaded on first open; it is nil only before the first load starts.
func (l *graphLoader) Graph() (*graph.Graph, bool) {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.current, l.loading
}

// reload starts loading the graph, cancelling any load in flight. The first load streams
// into the visible graph so the first screen appears quickly; later loads build a fresh
// graph off-screen and swap it in when complete, so a refresh never blanks the view.
func (l *graphLoader) reload() {
	l.mu.Lock()
	if l.cancel != nil {
		l.cancel()
	}
	ctx, cancel := context.WithCancel(context.Background())
	l.cancel = cancel
	g := graph.New()
	if l.current == nil {
		l.current = g
	}
	l.loading = true
	l.mu.Unlock()

	go func() {
		const every = 100 * time.Millisecond
		last := time.Now()
		err := l.repo.Log(ctx, git.LogOptions{}, func(c git.LogCommit) error {
			if err := g.Add(c); err != nil {
				return err
			}
			if l.onProgress != nil && time.Since(last) >= every {
				last = time.Now()
				l.onProgress(GraphProgress{Rows: g.Len()})
			}
			return nil
		})
		if ctx.Err() != nil {
			return // superseded or closed; the newer load reports
		}
		l.mu.Lock()
		l.current = g
		l.loading = false
		l.cancel = nil
		l.mu.Unlock()
		if l.onProgress != nil {
			l.onProgress(GraphProgress{Rows: g.Len(), Done: true, Err: err})
		}
	}()
}

func (l *graphLoader) stop() {
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.cancel != nil {
		l.cancel()
		l.cancel = nil
	}
}
