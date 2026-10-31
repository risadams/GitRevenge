package graph

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand/v2"
	"os/exec"
	"slices"
	"strings"
	"testing"

	"github.com/risadams/gitrevenge/internal/git"
	"github.com/risadams/gitrevenge/internal/gittest"
)

// fakeID makes a deterministic 40-hex-digit ID from n.
func fakeID(n int) string { return fmt.Sprintf("%040x", n+1) }

// build lays out commits given as child→parents in topological order.
func build(t *testing.T, commits [][]int) *Graph {
	t.Helper()
	g := New()
	for i, ps := range commits {
		c := git.LogCommit{ID: fakeID(i), Subject: fmt.Sprint(i), AuthorName: "a"}
		for _, p := range ps {
			c.Parents = append(c.Parents, fakeID(p))
		}
		if err := g.Add(c); err != nil {
			t.Fatal(err)
		}
	}
	return g
}

func cols(rows []Row) []int {
	out := make([]int, len(rows))
	for i, r := range rows {
		out[i] = r.Col
	}
	return out
}

func TestLinearStaysInOneLane(t *testing.T) {
	g := build(t, [][]int{{1}, {2}, {3}, {}})
	rows := g.Rows(0, 10)
	if !slices.Equal(cols(rows), []int{0, 0, 0, 0}) {
		t.Fatalf("cols = %v", cols(rows))
	}
	for _, r := range rows {
		if r.Color != rows[0].Color || r.Width != 1 {
			t.Fatalf("row %d: color %d width %d", r.Index, r.Color, r.Width)
		}
	}
	if len(rows[3].Down) != 0 {
		t.Fatal("root commit should have no lines below it")
	}
}

func TestBranchAndMerge(t *testing.T) {
	// 0: merge of 1 and 2
	// 1: main work      -> 3
	// 2: feature work   -> 3
	// 3: base
	g := build(t, [][]int{{1, 2}, {3}, {3}, {}})
	rows := g.Rows(0, 4)
	if !slices.Equal(cols(rows), []int{0, 0, 1, 0}) {
		t.Fatalf("cols = %v", cols(rows))
	}
	if !slices.Equal(rows[0].Down, []Seg{{0, 0, rows[0].Color}, {0, 1, rows[2].Color}}) {
		t.Fatalf("merge row down = %+v", rows[0].Down)
	}
	if rows[2].Color == rows[0].Color {
		t.Fatal("feature lane should get its own color")
	}
	// The feature commit's parent is already awaited by lane 0, so its line joins that lane
	// right away instead of running a parallel lane down to the base.
	if !slices.Equal(rows[2].Down, []Seg{{1, 0, rows[2].Color}, {0, 0, rows[0].Color}}) {
		t.Fatalf("feature row down = %+v", rows[2].Down)
	}
	if len(rows[3].Up) != 1 || rows[3].Width != 1 {
		t.Fatalf("base row should have a single lane: %+v", rows[3])
	}
}

func TestTwoTipsOpenTwoLanes(t *testing.T) {
	// 0: tip A -> 2, 1: tip B -> 2, 2: root
	g := build(t, [][]int{{2}, {2}, {}})
	rows := g.Rows(0, 3)
	if !slices.Equal(cols(rows), []int{0, 1, 0}) {
		t.Fatalf("cols = %v", cols(rows))
	}
	if len(rows[1].Up) != 1 || rows[1].Up[0].From != 0 {
		t.Fatalf("tip B should have lane 0 passing beside it: %+v", rows[1].Up)
	}
}

func TestFreedColumnIsReused(t *testing.T) {
	// 0 -> 1,2 (merge); 2 -> root 4; 1 -> 3; 3 -> 4; 4 root. Then a new tip later reuses col 1.
	g := build(t, [][]int{{1, 2}, {3}, {4}, {4}, {}})
	if w := g.Rows(4, 1)[0].Width; w > 2 {
		t.Fatalf("width at base = %d", w)
	}
}

// randomDAG produces n commits in a valid topological order with merges and many tips.
func randomDAG(n int, seed uint64) [][]int {
	r := rand.New(rand.NewPCG(seed, seed))
	commits := make([][]int, n)
	for i := range n - 1 {
		// Parents come later in the list.
		p1 := i + 1 + r.IntN(min(8, n-1-i))
		ps := []int{p1}
		if r.IntN(5) == 0 && p1+1 < n {
			p2 := p1 + 1 + r.IntN(min(20, n-1-p1))
			if p2 < n && p2 != p1 {
				ps = append(ps, p2)
			}
		}
		commits[i] = ps
	}
	return commits
}

// Lines must join up across rows: every line entering the top of a row continues an
// identical line (column and color) leaving the row above, and every line leaving a row
// lands on a column that continues in the next row. (A line joining an existing lane keeps
// its own color until it meets that lane, so only columns are compared in that direction.)
func TestRowsConnect(t *testing.T) {
	for seed := uint64(1); seed <= 5; seed++ {
		g := build(t, randomDAG(1000, seed))
		rows := g.Rows(0, g.Len())
		for i := 0; i+1 < len(rows); i++ {
			down := map[[2]int]bool{}
			downCols := map[int]bool{}
			for _, s := range rows[i].Down {
				down[[2]int{s.To, s.Color}] = true
				downCols[s.To] = true
			}
			upCols := map[int]bool{}
			for _, s := range rows[i+1].Up {
				upCols[s.From] = true
				if !down[[2]int{s.From, s.Color}] {
					t.Fatalf("seed %d row %d: line from col %d color %d has no source in row %d", seed, i+1, s.From, s.Color, i)
				}
			}
			for c := range downCols {
				if !upCols[c] {
					t.Fatalf("seed %d row %d: line to col %d does not continue into row %d", seed, i, c, i+1)
				}
			}
		}
		// The last row is a root: nothing leaves it.
		if last := rows[len(rows)-1]; len(last.Down) != 0 {
			t.Fatalf("seed %d: lines below the final root: %+v", seed, last.Down)
		}
	}
}

// A window replayed from a checkpoint must match the same rows from a full pass.
func TestWindowsMatchFullLayout(t *testing.T) {
	g := build(t, randomDAG(1000, 2))
	full := g.Rows(0, g.Len())
	for _, start := range []int{0, 1, 127, 128, 129, 500, 990} {
		win := g.Rows(start, 20)
		for i, r := range win {
			want := full[start+i]
			if r.Col != want.Col || r.Color != want.Color || !slices.Equal(r.Up, want.Up) || !slices.Equal(r.Down, want.Down) {
				t.Fatalf("row %d differs when read from window at %d", start+i, start)
			}
		}
	}
	if got := g.Rows(995, 100); len(got) != 5 {
		t.Fatalf("window past the end returned %d rows", len(got))
	}
}

// A graph wider than wideLanes snapshots less often; windows read anywhere in it still match
// the full layout.
func TestWideGraphSpacesCheckpoints(t *testing.T) {
	const tips = 2*wideLanes + 100 // every tip's lane stays open until its parent arrives
	commits := make([][]int, 0, 2*tips+200)
	for i := range tips {
		commits = append(commits, []int{tips + i})
	}
	for range tips {
		commits = append(commits, []int{2 * tips}) // parents all lead to one root line
	}
	for i := range 199 {
		commits = append(commits, []int{2*tips + i + 1})
	}
	commits = append(commits, []int{})
	g := build(t, commits)

	if width := g.Rows(tips, 1)[0].Width; width <= 2*wideLanes {
		t.Fatalf("width %d, want the fixture wider than %d", width, 2*wideLanes)
	}
	if n, dense := len(g.checkpoints), (g.Len()+checkpointEvery-1)/checkpointEvery; n >= dense {
		t.Fatalf("%d checkpoints, want fewer than the %d a narrow graph takes", n, dense)
	}

	full := g.Rows(0, g.Len())
	for _, start := range []int{0, 127, 128, 400, tips - 1, tips, tips + 300, 2*tips - 1, g.Len() - 5} {
		for i, r := range g.Rows(start, 20) {
			want := full[start+i]
			if r.Col != want.Col || r.Color != want.Color || !slices.Equal(r.Up, want.Up) || !slices.Equal(r.Down, want.Down) {
				t.Fatalf("row %d differs when read from window at %d", start+i, start)
			}
		}
	}
}

func TestRowOf(t *testing.T) {
	g := build(t, [][]int{{1}, {}})
	if r, ok := g.RowOf(fakeID(1)); !ok || r != 1 {
		t.Fatalf("RowOf = %d, %v", r, ok)
	}
	if _, ok := g.RowOf(fakeID(7)); ok {
		t.Fatal("unknown commit found")
	}
}

func TestFromRealRepo(t *testing.T) {
	bin, err := exec.LookPath("git")
	if err != nil {
		t.Skip("git not installed")
	}
	fx := gittest.WithMergeHistory(t)
	repo := git.OpenRepo(git.NewRunner(bin), fx.Dir)
	g := New()
	if err := repo.Log(context.Background(), git.LogOptions{}, g.Add); err != nil {
		t.Fatal(err)
	}
	rows := g.Rows(0, 100)
	if len(rows) != 6 || rows[0].ID != fx.Git("rev-parse", rows[0].ID) {
		t.Fatalf("rows = %d", len(rows))
	}
	if rows[0].Author != "Test User" || rows[0].AuthorTime == 0 {
		t.Fatalf("row 0 = %+v", rows[0])
	}
}

func BenchmarkAdd(b *testing.B) {
	commits := randomDAG(100_000, 3)
	logs := make([]git.LogCommit, len(commits))
	for i, ps := range commits {
		logs[i] = git.LogCommit{ID: fakeID(i), AuthorName: "a", Subject: "s"}
		for _, p := range ps {
			logs[i].Parents = append(logs[i].Parents, fakeID(p))
		}
	}
	b.ResetTimer()
	for b.Loop() {
		g := New()
		for _, c := range logs {
			_ = g.Add(c)
		}
	}
}

func BenchmarkRowsWindow(b *testing.B) {
	g := New()
	for i, ps := range randomDAG(100_000, 4) {
		c := git.LogCommit{ID: fakeID(i), AuthorName: "a", Subject: "s"}
		for _, p := range ps {
			c.Parents = append(c.Parents, fakeID(p))
		}
		_ = g.Add(c)
	}
	b.ResetTimer()
	for i := 0; b.Loop(); i++ {
		g.Rows((i*7919)%99_000, 100)
	}
}

func TestRowsJSONHasNoNullSegments(t *testing.T) {
	g := build(t, [][]int{{1}, {}})
	b, err := json.Marshal(g.Rows(0, 2))
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(b), "null") {
		t.Fatalf("rows JSON contains null: %s", b)
	}
}
