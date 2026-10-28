// Package perf measures Git Revenge against a large real repository. It only runs when
// GITREVENGE_BENCH_REPO points at one, e.g.:
//
//	GITREVENGE_BENCH_REPO=/src/linux go test ./internal/perf -v -count=1 -timeout 30m
//
// The targets come from docs/PLAN.md. Run before tagging each milestone.
package perf

import (
	"context"
	"errors"
	"math/rand/v2"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"github.com/risadams/gitrevenge/internal/git"
	"github.com/risadams/gitrevenge/internal/graph"
)

const (
	firstPageRows    = 100
	targetFirstPage  = time.Second
	targetHeapBytes  = 500 << 20
	windowReads      = 1000
	targetWindowRead = 16 * time.Millisecond // one frame at 60fps
)

func benchRepo(t *testing.T) *git.Repo {
	t.Helper()
	dir := os.Getenv("GITREVENGE_BENCH_REPO")
	if dir == "" {
		t.Skip("set GITREVENGE_BENCH_REPO to a large repository to run performance checks")
	}
	bin, err := exec.LookPath("git")
	if err != nil {
		t.Skip("git not installed")
	}
	r := git.OpenRepo(git.NewRunner(bin), dir)
	gitDir, err := r.ReadString(context.Background(), "rev-parse", "--absolute-git-dir")
	if err != nil {
		t.Fatalf("%s is not a git repository: %v", dir, err)
	}
	if _, err := os.Stat(filepath.Join(gitDir, "objects", "info", "commit-graph")); err != nil {
		if _, err := os.Stat(filepath.Join(gitDir, "objects", "info", "commit-graphs")); err != nil {
			t.Logf("WARNING: no commit-graph file; git must walk all history before the first commit. " +
				"Run `git commit-graph write --reachable` for representative numbers.")
		}
	}
	return r
}

var errFirstPage = errors.New("first page reached")

func TestTargets(t *testing.T) {
	r := benchRepo(t)
	ctx := context.Background()

	// Time to first screen: from starting git log until enough rows are laid out to draw.
	start := time.Now()
	g := graph.New()
	err := r.Log(ctx, git.LogOptions{}, func(c git.LogCommit) error {
		if err := g.Add(c); err != nil {
			return err
		}
		if g.Len() == firstPageRows {
			return errFirstPage
		}
		return nil
	})
	if err != nil && !errors.Is(err, errFirstPage) {
		t.Fatal(err)
	}
	firstPage := time.Since(start)

	// Full load.
	start = time.Now()
	g = graph.New()
	if err := r.Log(ctx, git.LogOptions{}, g.Add); err != nil {
		t.Fatal(err)
	}
	fullLoad := time.Since(start)

	runtime.GC()
	var ms runtime.MemStats
	runtime.ReadMemStats(&ms)

	// Random window reads, as when scrolling or jumping around.
	n := g.Len()
	rnd := rand.New(rand.NewPCG(1, 1))
	start = time.Now()
	var worst time.Duration
	for range windowReads {
		s := time.Now()
		g.Rows(rnd.IntN(max(1, n)), firstPageRows)
		worst = max(worst, time.Since(s))
	}
	avgWindow := time.Since(start) / windowReads

	start = time.Now()
	res := g.Search("fix", 10000)
	search := time.Since(start)

	start = time.Now()
	refs, err := r.Refs(ctx)
	if err != nil {
		t.Fatal(err)
	}
	refsTime := time.Since(start)

	t.Logf("repository:           %s", os.Getenv("GITREVENGE_BENCH_REPO"))
	t.Logf("commits:              %d", n)
	t.Logf("refs:                 %d (read in %v)", len(refs), refsTime.Round(time.Millisecond))
	t.Logf("first %d rows:       %v (target < %v)", firstPageRows, firstPage.Round(time.Millisecond), targetFirstPage)
	t.Logf("full load + layout:   %v", fullLoad.Round(time.Millisecond))
	t.Logf("heap in use:          %d MB (target < %d MB)", ms.HeapInuse>>20, targetHeapBytes>>20)
	t.Logf("100-row window:       avg %v, worst %v (target < %v)", avgWindow.Round(time.Microsecond), worst.Round(time.Microsecond), targetWindowRead)
	t.Logf("search \"fix\":         %v (%d matches)", search.Round(time.Millisecond), res.Total)

	if firstPage > targetFirstPage {
		t.Errorf("first page took %v, target %v", firstPage, targetFirstPage)
	}
	if ms.HeapInuse > targetHeapBytes {
		t.Errorf("heap in use %d MB, target %d MB", ms.HeapInuse>>20, targetHeapBytes>>20)
	}
	if worst > targetWindowRead {
		t.Errorf("worst window read %v, target %v", worst, targetWindowRead)
	}
}

func BenchmarkFullLoad(b *testing.B) {
	r := benchRepo(&testing.T{})
	for b.Loop() {
		g := graph.New()
		if err := r.Log(context.Background(), git.LogOptions{}, g.Add); err != nil {
			b.Fatal(err)
		}
	}
}
