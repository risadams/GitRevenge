package app

import (
	"github.com/risadams/gitrevenge/internal/git"
	"github.com/risadams/gitrevenge/internal/graph"
	"github.com/risadams/gitrevenge/internal/session"
)

// maxGraphPage caps one GraphRows request so a bad caller can't ask for a million rows.
const maxGraphPage = 1000

// GraphPage is a window of commit graph rows.
type GraphPage struct {
	Rows []graph.Row `json:"rows"`
	// Refs maps commit IDs in Rows to the refs pointing at them.
	Refs map[string][]git.Ref `json:"refs"`
	// Total is the number of rows laid out so far.
	Total   int  `json:"total"`
	Loading bool `json:"loading"`
}

// GraphProgress is the payload of the "graph:progress" event.
type GraphProgress struct {
	Path  string `json:"path"`
	Rows  int    `json:"rows"`
	Done  bool   `json:"done"`
	Error string `json:"error"`
}

// GraphRows returns up to count rows of the commit graph starting at start.
func (a *App) GraphRows(path string, start, count int) (GraphPage, error) {
	m, err := a.sessions()
	if err != nil {
		return GraphPage{}, err
	}
	r := m.Get(path)
	if r == nil {
		return GraphPage{}, userError("That repository isn't open.")
	}
	g, loading := r.Graph()
	if g == nil {
		return GraphPage{Rows: []graph.Row{}, Refs: map[string][]git.Ref{}, Loading: true}, nil
	}
	rows := g.Rows(start, min(count, maxGraphPage))
	idx := r.Refs()
	refs := map[string][]git.Ref{}
	for _, row := range rows {
		if rs := idx.ByCommit[row.ID]; len(rs) > 0 {
			refs[row.ID] = rs
		}
	}
	return GraphPage{Rows: rows, Refs: refs, Total: g.Len(), Loading: loading}, nil
}

// Refs lists the branches, remote branches, tags and stashes shown in the graph.
func (a *App) Refs(path string) ([]git.Ref, error) {
	r, err := a.openRepo(path)
	if err != nil {
		return nil, err
	}
	refs := r.Refs().All
	if refs == nil {
		refs = []git.Ref{}
	}
	return refs, nil
}

// CommitRow returns the graph row of a commit, or -1 if it isn't in the loaded graph.
func (a *App) CommitRow(path, id string) (int, error) {
	r, err := a.openRepo(path)
	if err != nil {
		return -1, err
	}
	g, _ := r.Graph()
	if g == nil {
		return -1, nil
	}
	if row, ok := g.RowOf(id); ok {
		return row, nil
	}
	return -1, nil
}

// openRepo returns the open repository at path.
func (a *App) openRepo(path string) (*session.Repo, error) {
	m, err := a.sessions()
	if err != nil {
		return nil, err
	}
	r := m.Get(path)
	if r == nil {
		return nil, userError("That repository isn't open.")
	}
	return r, nil
}

func (a *App) emitGraphProgress(path string, p session.GraphProgress) {
	ev := GraphProgress{Path: path, Rows: p.Rows, Done: p.Done}
	if p.Err != nil {
		ev.Error = p.Err.Error()
	}
	a.emit("graph:progress", ev)
}
