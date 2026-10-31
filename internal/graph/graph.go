// Package graph lays out commit history as rows and lanes for the commit graph view.
//
// Commits must be added children before parents, as `git log --date-order` produces them.
// Layout is incremental: each Add places one more row.
//
// Only each row's node column and color are stored, plus a snapshot of the lane state every
// checkpointEvery rows, or further apart where the graph is wide (see checkpointGap). Rows
// replays from the nearest snapshot to produce the full edge data for a window, so memory
// grows with the number of commits, not commits × graph width.
package graph

import (
	"encoding/hex"
	"fmt"
	"slices"
	"sort"
	"sync"

	"github.com/risadams/gitrevenge/internal/git"
)

const checkpointEvery = 128

// wideLanes is the width past which snapshots spread out: a graph twice as wide snapshots
// half as often. Snapshot memory is rows/gap × width, so it stops growing with width, and a
// window read replays at most about checkpointEvery × wideLanes lane steps. Date order puts
// the Linux kernel's parallel lines side by side, thousands of lanes wide; snapshotting
// that every 128 rows took the full-history heap from 467 MB to 641 MB.
const wideLanes = 1024

// checkpointGap is how many rows to lay out after a snapshot taken at the given width.
func checkpointGap(width int) int {
	return checkpointEvery * max(1, (width+wideLanes-1)/wideLanes)
}

// oid is a binary object ID. 32 bytes fits both SHA-1 (20) and SHA-256 (32) repositories.
type oid [32]byte

func parseOID(s string) (oid, error) {
	var o oid
	if len(s) > 64 || len(s)%2 != 0 {
		return o, fmt.Errorf("bad object id %q", s)
	}
	if _, err := hex.Decode(o[:], []byte(s)); err != nil {
		return o, fmt.Errorf("bad object id %q: %w", s, err)
	}
	return o, nil
}

// Seg is one line drawn in a row, between lane columns. Up segments run from the top of
// the row to its middle; Down segments from the middle to the bottom.
type Seg struct {
	From  int `json:"from"`
	To    int `json:"to"`
	Color int `json:"color"`
}

// Row is everything needed to draw one commit row.
type Row struct {
	Index int `json:"index"`
	// Col and Color place the commit's node.
	Col   int   `json:"col"`
	Color int   `json:"color"`
	Up    []Seg `json:"up"`
	Down  []Seg `json:"down"`
	// Width is the number of lane columns used by this row.
	Width int `json:"width"`

	ID         string `json:"id"`
	Parents    int    `json:"parents"`
	Author     string `json:"author"`
	AuthorTime int64  `json:"authorTime"`
	Subject    string `json:"subject"`
}

type lane struct {
	want oid
	// pos is where want sits in Graph.parents, so a checkpoint can store the lane as a
	// position instead of a 32-byte ID.
	pos   int32
	color int
	used  bool
}

type state struct {
	lanes     []lane
	nextColor int
}

// checkpoint is a state stored compactly: each lane as the position of the parent it waits
// for (-1 for a free lane) and its color. On a wide history (the Linux kernel runs to
// about 800 lanes) this is a sixth of the size of the state itself.
type checkpoint struct {
	row       int // the state before this row
	lanes     []cpLane
	nextColor int
}

type cpLane struct {
	pos   int32
	color int32
}

func (st state) pack(row int) checkpoint {
	cp := checkpoint{row: row, lanes: make([]cpLane, len(st.lanes)), nextColor: st.nextColor}
	for i, l := range st.lanes {
		cp.lanes[i] = cpLane{pos: -1}
		if l.used {
			cp.lanes[i] = cpLane{pos: l.pos, color: int32(l.color)}
		}
	}
	return cp
}

// unpack restores a checkpoint, reading each lane's commit ID back from parents.
func (cp checkpoint) unpack(parents []oid) state {
	st := state{lanes: make([]lane, len(cp.lanes)), nextColor: cp.nextColor}
	for i, l := range cp.lanes {
		if l.pos >= 0 {
			st.lanes[i] = lane{want: parents[l.pos], pos: l.pos, color: int(l.color), used: true}
		}
	}
	return st
}

// Graph holds commits and their layout. It is safe for one writer and many readers.
type Graph struct {
	mu sync.RWMutex

	hashLen int // bytes per object ID in this repository (20 or 32)

	ids          []oid
	parentStart  []int32 // parents of commit i are parents[parentStart[i]:parentStart[i+1]]
	parents      []oid
	index        map[oid]int32
	authorIdx    []int32
	authors      []string
	authorLookup map[string]int32
	times        []int64
	subjects     []string

	nodeCol   []uint16
	nodeColor []uint16

	cur         state
	pins        *pinning     // Trunk lanes held in the leftmost columns, or nil
	checkpoints []checkpoint // in row order; the first is before row 0
	nextCP      int          // the row before which the next checkpoint is taken
}

// New returns an empty Graph.
func New() *Graph {
	return &Graph{
		parentStart:  []int32{0},
		index:        map[oid]int32{},
		authorLookup: map[string]int32{},
	}
}

// pinning holds Trunk lanes in the leftmost columns (ADR 0009). Columns below n only ever
// hold a pinned lane, so whether a lane is pinned follows from its column and checkpoints
// don't need to store it.
type pinning struct {
	tips map[oid]int // a Trunk branch's tip → its column
	n    int
}

// Pin holds the lanes of the given commits, Trunk branch tips in order, in columns 0, 1, …
// from their tip down their first parents: main reads as a straight spine at the left. It
// must be called before the first Add; duplicates and unknown IDs are ignored.
func (g *Graph) Pin(ids []string) {
	g.mu.Lock()
	defer g.mu.Unlock()
	if len(g.ids) > 0 {
		return
	}
	p := &pinning{tips: map[oid]int{}}
	for _, id := range ids {
		o, err := parseOID(id)
		if err != nil {
			continue
		}
		if _, dup := p.tips[o]; !dup {
			p.tips[o] = p.n
			p.n++
		}
	}
	if p.n > 0 {
		g.pins = p
		// Colour 0 is the first Trunk's; other lanes start from the next one.
		g.cur.nextColor = 1
	}
}

// reserved is how many leftmost columns pinned lanes own.
func (p *pinning) reserved() int {
	if p == nil {
		return 0
	}
	return p.n
}

// Add appends the next commit (children before parents) and lays out its row.
func (g *Graph) Add(c git.LogCommit) error {
	id, err := parseOID(c.ID)
	if err != nil {
		return err
	}
	ps := make([]oid, len(c.Parents))
	for i, p := range c.Parents {
		if ps[i], err = parseOID(p); err != nil {
			return err
		}
	}

	g.mu.Lock()
	defer g.mu.Unlock()
	if g.hashLen == 0 {
		g.hashLen = len(c.ID) / 2
	}
	row := len(g.ids)
	if row == g.nextCP {
		g.checkpoints = append(g.checkpoints, g.cur.pack(row))
		g.nextCP = row + checkpointGap(len(g.cur.lanes))
	}
	g.index[id] = int32(row)
	g.ids = append(g.ids, id)
	base := int32(len(g.parents))
	g.parents = append(g.parents, ps...)
	g.parentStart = append(g.parentStart, int32(len(g.parents)))
	a, ok := g.authorLookup[c.AuthorName]
	if !ok {
		a = int32(len(g.authors))
		g.authors = append(g.authors, c.AuthorName)
		g.authorLookup[c.AuthorName] = a
	}
	g.authorIdx = append(g.authorIdx, a)
	g.times = append(g.times, c.AuthorTime)
	g.subjects = append(g.subjects, c.Subject)

	col, color := step(&g.cur, g.pins, id, ps, base, nil)
	g.nodeCol = append(g.nodeCol, uint16(min(col, 0xffff)))
	g.nodeColor = append(g.nodeColor, uint16(color%0x10000))
	return nil
}

// Len returns the number of rows laid out so far.
func (g *Graph) Len() int {
	g.mu.RLock()
	defer g.mu.RUnlock()
	return len(g.ids)
}

// RowOf returns the row of the commit with the given full ID.
func (g *Graph) RowOf(id string) (int, bool) {
	o, err := parseOID(id)
	if err != nil {
		return 0, false
	}
	g.mu.RLock()
	defer g.mu.RUnlock()
	r, ok := g.index[o]
	return int(r), ok
}

// Rows returns up to count rows starting at start, with full edge data.
func (g *Graph) Rows(start, count int) []Row {
	g.mu.RLock()
	defer g.mu.RUnlock()
	n := len(g.ids)
	if start < 0 {
		start = 0
	}
	end := min(start+count, n)
	if start >= end {
		return []Row{}
	}

	// The last checkpoint at or before start.
	k := sort.Search(len(g.checkpoints), func(k int) bool { return g.checkpoints[k].row > start }) - 1
	cp := g.checkpoints[k]
	st := cp.unpack(g.parents)
	out := make([]Row, 0, end-start)
	for i := cp.row; i < end; i++ {
		ps := g.parents[g.parentStart[i]:g.parentStart[i+1]]
		if i < start {
			step(&st, g.pins, g.ids[i], ps, g.parentStart[i], nil)
			continue
		}
		// Empty, not nil: the frontend iterates these, and nil marshals to JSON null.
		r := Row{Index: i, Up: []Seg{}, Down: []Seg{}}
		r.Col, r.Color = step(&st, g.pins, g.ids[i], ps, g.parentStart[i], &r)
		r.ID = hex.EncodeToString(g.ids[i][:g.hashLen])
		r.Parents = len(ps)
		r.Author = g.authors[g.authorIdx[i]]
		r.AuthorTime = g.times[i]
		r.Subject = g.subjects[i]
		out = append(out, r)
	}
	return out
}

// step places the commit id with parents ps, advancing st by one row. base is where ps
// starts in Graph.parents. If row is non-nil,
// the segments drawn in this row are recorded in it. It returns the node column and color.
//
// Lanes keep their column while active, so lines run straight down; a freed column is
// reused by the next new lane. With pins, the leftmost columns belong to Trunk lanes: a
// Trunk tip takes its column even if a branch above reached it first, its first parent
// always continues there, and other lanes start to the right.
func step(st *state, pins *pinning, id oid, ps []oid, base int32, row *Row) (col, color int) {
	reserved := pins.reserved()
	col = -1
	for i, l := range st.lanes {
		if l.used && l.want == id {
			col = i
			break
		}
	}
	pinCol, isTip := -1, false
	if pins != nil {
		pinCol, isTip = pins.tips[id]
	}
	switch {
	case isTip && col != pinCol:
		// A Trunk tip: its own column, with any lanes waiting for it converging there.
		for len(st.lanes) <= pinCol {
			st.lanes = append(st.lanes, lane{})
		}
		col = pinCol
		color = pinColor(st, pinCol)
	case col < 0:
		// A branch tip nothing above points to: start a new lane.
		col = st.freeSlot(reserved)
		color = st.newColor()
	default:
		color = st.lanes[col].color
	}

	// Lines entering from above: lanes waiting for this commit converge on the node, the
	// rest pass straight through. Then clear every lane that was waiting for this commit.
	for i, l := range st.lanes {
		if !l.used {
			continue
		}
		if l.want == id {
			if row != nil {
				row.Up = append(row.Up, Seg{From: i, To: col, Color: l.color})
			}
			st.lanes[i] = lane{}
		} else if row != nil {
			row.Up = append(row.Up, Seg{From: i, To: i, Color: l.color})
		}
	}

	// Lines leaving downward: one per parent. A parent another lane already waits for joins
	// that lane; otherwise the first parent continues in this column and others get new lanes.
	var created []int // lanes started by this commit; tiny, so a slice beats a map
	for pi, p := range ps {
		target, tcolor := -1, 0
		if pi == 0 && col < reserved {
			// A Trunk lane holds its column: its first parent continues here even if another
			// lane already waits for it (that lane converges here when the parent arrives).
			target, tcolor = col, color
			// The Up pass cleared this column, so the lane starts again here for the parent.
			if l := st.lanes[col]; !l.used || l.want != p {
				st.lanes[target] = lane{want: p, pos: base + int32(pi), color: tcolor, used: true}
				created = append(created, target)
			}
			if row != nil {
				row.Down = append(row.Down, Seg{From: col, To: target, Color: color})
			}
			continue
		}
		for i, l := range st.lanes {
			if l.used && l.want == p {
				target, tcolor = i, l.color
				break
			}
		}
		if target < 0 {
			if pi == 0 {
				target, tcolor = col, color
			} else {
				target, tcolor = st.freeSlot(max(col+1, reserved)), st.newColor()
			}
			st.lanes[target] = lane{want: p, pos: base + int32(pi), color: tcolor, used: true}
			created = append(created, target)
		}
		if row != nil {
			segColor := tcolor
			if pi == 0 {
				segColor = color
			}
			row.Down = append(row.Down, Seg{From: col, To: target, Color: segColor})
		}
	}
	if row != nil {
		for i, l := range st.lanes {
			if l.used && !slices.Contains(created, i) {
				row.Down = append(row.Down, Seg{From: i, To: i, Color: l.color})
			}
		}
		row.Width = len(st.lanes)
		if col+1 > row.Width {
			row.Width = col + 1
		}
	}
	// Drop trailing free lanes so the graph narrows again.
	for len(st.lanes) > 0 && !st.lanes[len(st.lanes)-1].used {
		st.lanes = st.lanes[:len(st.lanes)-1]
	}
	return col, color
}

// pinColor is the colour of a pinned lane: the first Trunk wears colour 0 (the theme's
// rose --lane-0), the others the next colours in turn.
func pinColor(st *state, pinCol int) int {
	if pinCol == 0 {
		return 0
	}
	return st.newColor()
}

// freeSlot returns the first unused lane at or after from, growing the lanes if needed.
// The returned slot is reserved only once the caller assigns it.
func (st *state) freeSlot(from int) int {
	for i := from; i < len(st.lanes); i++ {
		if !st.lanes[i].used {
			return i
		}
	}
	for len(st.lanes) < from {
		st.lanes = append(st.lanes, lane{})
	}
	st.lanes = append(st.lanes, lane{})
	return len(st.lanes) - 1
}

func (st *state) newColor() int {
	c := st.nextColor
	st.nextColor++
	return c
}
