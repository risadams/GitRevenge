package git

import (
	"context"
	"testing"
	"time"

	"github.com/risadams/gitrevenge/internal/gittest"
)

func TestStats(t *testing.T) {
	fx := gittest.New(t)
	fx.CommitFile("a.txt", "1\n2\n3\n", "add a")
	fx.CommitFile("a.txt", "1\n2\n3\n4\n", "grow a")
	fx.Branch("feature")
	fx.CommitFile("b.txt", "x\n", "add b on feature")
	fx.Checkout("main")
	fx.CommitFile("c.txt", "y\n", "add c on main")
	if err := fx.Merge("feature"); err != nil {
		t.Fatal(err)
	}
	repo := OpenRepo(newTestRunner(t), fx.Dir)
	// Fixture commits are dated from 2026-01-01, a minute apart.
	now := time.Date(2026, 1, 2, 0, 0, 0, 0, time.UTC).Unix()
	st, err := repo.Stats(context.Background(), StatsOptions{Since: now - 30*24*3600, Until: now + 3600})
	if err != nil {
		t.Fatal(err)
	}
	if st.Commits != 5 || st.Merges != 1 || st.Bucket != "day" {
		t.Fatalf("commits %d, merges %d, bucket %s", st.Commits, st.Merges, st.Bucket)
	}
	total := 0
	for _, b := range st.Activity {
		total += b.Commits
	}
	if total != 5 {
		t.Fatalf("activity sums to %d, want 5", total)
	}
	if len(st.Contributors) != 1 || st.Contributors[0].Commits != 5 || st.Contributors[0].Added != 6 {
		t.Fatalf("contributors = %+v", st.Contributors)
	}
	if st.Hotspots[0].Path != "a.txt" || st.Hotspots[0].Commits != 2 || st.Hotspots[0].Changed != 4 {
		t.Fatalf("hotspots = %+v", st.Hotspots)
	}
	if st.Cycle.Sample != 1 || st.Cycle.Median < 0 {
		t.Fatalf("cycle = %+v", st.Cycle)
	}

	// A long range counts by week; one branch narrows the history.
	st, err = repo.Stats(context.Background(), StatsOptions{Since: now - 400*24*3600, Until: now + 3600, Rev: "feature"})
	if err != nil || st.Bucket != "week" || st.Commits != 3 || st.Merges != 0 {
		t.Fatalf("feature over a year = %+v, %v", st, err)
	}
}

func TestBranchAges(t *testing.T) {
	fx := gittest.New(t)
	fx.Linear(1)
	fx.Branch("side")
	ages, err := OpenRepo(newTestRunner(t), fx.Dir).BranchAges(context.Background())
	if err != nil || len(ages) != 2 || ages[0].Kind != "local" || ages[0].Last == 0 {
		t.Fatalf("ages = %+v, %v", ages, err)
	}
}
