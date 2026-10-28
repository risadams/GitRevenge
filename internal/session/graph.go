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
	// RefsReady is set on the update sent when refs finish loading during a first load,
	// so the UI can re-fetch rows to pick up their labels.
	RefsReady bool
	Err       error
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
		// Read refs alongside the log rather than before it: with thousands of refs they can
		// take longer than the first screen of history.
		type refResult struct {
			idx *RefIndex
			err error
		}
		refsCh := make(chan refResult, 1)
		go func() {
			list, err := l.repo.Refs(ctx)
			idx := newRefIndex(list)
			first := false
			l.mu.Lock()
			if l.refs == nil && ctx.Err() == nil {
				l.refs = idx
				first = true
			}
			l.mu.Unlock()
			if first && l.onProgress != nil {
				l.onProgress(GraphProgress{Rows: g.Len(), RefsReady: true})
			}
			refsCh <- refResult{idx, err}
		}()

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
		refs := <-refsCh
		if ctx.Err() != nil {
			return // superseded or closed; the newer load reports
		}
		if err == nil {
			err = refs.err
		}
		l.mu.Lock()
		l.current = g
		l.refs = refs.idx
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
