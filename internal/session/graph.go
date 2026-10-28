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
	refs    *RefIndex
	loading bool
	cancel  context.CancelFunc
}

// RefIndex holds a repository's refs, indexed by the commit they point at.
type RefIndex struct {
	All      []git.Ref
	ByCommit map[string][]git.Ref
}

func newRefIndex(refs []git.Ref) *RefIndex {
	idx := &RefIndex{All: refs, ByCommit: make(map[string][]git.Ref, len(refs))}
	for _, r := range refs {
		idx.ByCommit[r.Commit] = append(idx.ByCommit[r.Commit], r)
	}
	return idx
}

// Refs returns the refs matching the displayed graph.
func (l *graphLoader) Refs() *RefIndex {
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.refs == nil {
		return newRefIndex(nil)
	}
	return l.refs
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
		// Refs are quick to read; load them first so labels appear with the first rows.
		refList, refErr := l.repo.Refs(ctx)
		refs := newRefIndex(refList)
		l.mu.Lock()
		if l.refs == nil && ctx.Err() == nil {
			l.refs = refs
		}
		l.mu.Unlock()

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
		if err == nil {
			err = refErr
		}
		l.mu.Lock()
		l.current = g
		l.refs = refs
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
