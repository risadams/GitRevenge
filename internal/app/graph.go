package app

import (
	"github.com/risadams/gitrevenge/internal/graph"
	"github.com/risadams/gitrevenge/internal/session"
)

// maxGraphPage caps one GraphRows request so a bad caller can't ask for a million rows.
const maxGraphPage = 1000

// GraphPage is a window of commit graph rows.
type GraphPage struct {
	Rows []graph.Row `json:"rows"`
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
		return GraphPage{Rows: []graph.Row{}, Loading: true}, nil
	}
	return GraphPage{Rows: g.Rows(start, min(count, maxGraphPage)), Total: g.Len(), Loading: loading}, nil
}

func (a *App) emitGraphProgress(path string, p session.GraphProgress) {
	ev := GraphProgress{Path: path, Rows: p.Rows, Done: p.Done}
	if p.Err != nil {
		ev.Error = p.Err.Error()
	}
	a.emit("graph:progress", ev)
}
