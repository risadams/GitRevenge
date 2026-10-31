package app

import (
	"testing"
	"time"

	"github.com/risadams/gitrevenge/internal/gittest"
)

// loadedPage waits for the graph of path to finish loading and returns its first rows.
func loadedPage(t *testing.T, a *App, path string) GraphPage {
	t.Helper()
	var page GraphPage
	var err error
	for deadline := time.Now().Add(10 * time.Second); time.Now().Before(deadline); time.Sleep(20 * time.Millisecond) {
		if page, err = a.GraphRows(path, 0, 50); err != nil {
			t.Fatal(err)
		}
		if !page.Loading {
			return page
		}
	}
	t.Fatal("graph didn't finish loading")
	return page
}

func TestHideBranch(t *testing.T) {
	a := newTestApp(t)
	fx := gittest.New(t)
	fx.CommitFile("a.txt", "1", "base")
	fx.Branch("wip/try")
	fx.CommitFile("b.txt", "1", "wip only")
	fx.Checkout("main")
	tab, err := a.OpenRepo(fx.Dir)
	if err != nil {
		t.Fatal(err)
	}
	if page := loadedPage(t, a, tab.Path); page.Total != 2 {
		t.Fatalf("total = %d before hiding", page.Total)
	}

	if err := a.HideBranch(tab.Path, "wip/try", true); err != nil {
		t.Fatal(err)
	}
	// The reload replaces the graph; wait until it shows one commit.
	var page GraphPage
	for deadline := time.Now().Add(10 * time.Second); time.Now().Before(deadline); time.Sleep(20 * time.Millisecond) {
		if page = loadedPage(t, a, tab.Path); page.Total == 1 {
			break
		}
	}
	if page.Total != 1 || page.Rows[0].Subject != "base" {
		t.Fatalf("after hiding: total %d", page.Total)
	}
	refs, _ := a.Refs(tab.Path)
	hidden := 0
	for _, r := range refs {
		if r.Hidden {
			hidden++
			if r.Short != "wip/try" {
				t.Fatalf("%s marked hidden", r.Short)
			}
		}
	}
	if hidden != 1 {
		t.Fatalf("%d refs marked hidden", hidden)
	}

	// Showing hidden branches puts it back, for the session only.
	if err := a.SetShowHidden(tab.Path, true); err != nil {
		t.Fatal(err)
	}
	for deadline := time.Now().Add(10 * time.Second); time.Now().Before(deadline); time.Sleep(20 * time.Millisecond) {
		if page = loadedPage(t, a, tab.Path); page.Total == 2 {
			break
		}
	}
	if page.Total != 2 {
		t.Fatalf("showing hidden: total %d", page.Total)
	}
	if on, _ := a.ShowingHidden(tab.Path); !on {
		t.Fatal("ShowingHidden = false")
	}
	_ = a.SetShowHidden(tab.Path, false)

	// A wider pattern keeps it hidden after unhiding the exact name.
	p, _ := a.RepoPreferences(tab.Path)
	p.Hide = []string{"wip/*", "wip/try"}
	if err := a.SetRepoPreferences(tab.Path, p); err != nil {
		t.Fatal(err)
	}
	if err := a.HideBranch(tab.Path, "wip/try", false); err == nil {
		t.Fatal("unhiding under wip/* should say which pattern still hides it")
	}
	if p, _ := a.RepoPreferences(tab.Path); len(p.Hide) != 1 || p.Hide[0] != "wip/*" {
		t.Fatalf("hide = %v", p.Hide)
	}
}

func TestRefKinds(t *testing.T) {
	a := newTestApp(t)
	fx := gittest.New(t)
	fx.CommitFile("a.txt", "1", "base")
	fx.Branch("release/1.0")
	fx.Branch("gh-pages")
	fx.Branch("feature/x")
	fx.Checkout("main")
	tab, err := a.OpenRepo(fx.Dir)
	if err != nil {
		t.Fatal(err)
	}
	loadedPage(t, a, tab.Path)
	refs, _ := a.Refs(tab.Path)
	want := map[string]string{"main": "trunk", "release/1.0": "release", "gh-pages": "docs", "feature/x": ""}
	for _, r := range refs {
		if w, ok := want[r.Short]; ok && r.BranchKind != w {
			t.Errorf("%s kind = %q, want %q", r.Short, r.BranchKind, w)
		}
	}
}
