package app

import (
	"slices"
	"testing"

	"github.com/risadams/gitrevenge/internal/gittest"
)

// A repository with release branches, a gh-pages site, a bot's merged branch and a commit by
// a GitHub App, for the branch-kind side of Insights (ADR 0009).
func kindsFixture(t *testing.T) *gittest.Repo {
	t.Helper()
	fx := gittest.New(t)
	fx.CommitFile("a.txt", "1\n", "Start")
	fx.Git("branch", "release/1.0")
	fx.CommitFile("a.txt", "2\n", "Feature one")
	fx.CommitFile("a.txt", "3\n", "Feature two")
	fx.Git("branch", "release/1.1")
	fx.Git("checkout", "-q", "-b", "dependabot/npm/vite-7")
	fx.CommitFile("deps.txt", "vite 7\n", "Bump vite to 7")
	fx.Checkout("main")
	fx.Git("merge", "-q", "--no-ff", "-m", "Merge branch 'dependabot/npm/vite-7'", "dependabot/npm/vite-7")
	fx.WriteFile("deps.txt", "vite 7.1\n")
	fx.Git("-c", "user.name=renovate[bot]", "commit", "-qam", "Update vite to 7.1")
	fx.CommitFile("a.txt", "4\n", "Feature three")
	fx.Git("checkout", "-q", "--orphan", "gh-pages")
	fx.Git("rm", "-rq", "--cached", ".")
	fx.CommitFile("index.html", "site\n", "Publish the site")
	fx.Checkout("main")
	return fx
}

func TestInsightsShippedMarksBots(t *testing.T) {
	fx := kindsFixture(t)
	a := newTestApp(t)
	res, err := a.InsightsShipped(fx.Dir, InsightsInputs{Window: "14d"})
	if err != nil {
		t.Fatal(err)
	}
	bots := map[string]bool{}
	for _, it := range res.Items {
		bots[it.Title] = it.Bot
	}
	want := map[string]bool{"Merge branch 'dependabot/npm/vite-7'": true, "Update vite to 7.1": true, "Feature three": false, "Feature one": false}
	for title, bot := range want {
		if got, ok := bots[title]; !ok || got != bot {
			t.Errorf("%q bot = %v (listed %v), want %v", title, got, ok, bot)
		}
	}
	if _, ok := bots["Publish the site"]; ok {
		t.Error("a gh-pages commit is listed as shipped")
	}
	if !slices.Contains(res.Window.Kinds["docs"], "gh-pages") || !slices.Contains(res.Window.Kinds["bot"], "dependabot/*") {
		t.Errorf("window kinds = %v, want the patterns in use", res.Window.Kinds)
	}
}

func TestInsightsBranchesLeaveOutDocs(t *testing.T) {
	fx := kindsFixture(t)
	a := newTestApp(t)
	res, err := a.InsightsBranches(fx.Dir, InsightsInputs{Window: "14d"})
	if err != nil {
		t.Fatal(err)
	}
	for _, b := range res.Branches {
		if b.Name == "gh-pages" {
			t.Fatal("the Docs branch is listed among open branches")
		}
	}
	// Marking gh-pages ordinary brings it back, and the cache doesn't hide the change.
	if err := a.MarkBranchKind(fx.Dir, "gh-pages", ""); err != nil {
		t.Fatal(err)
	}
	again, _ := a.InsightsBranches(fx.Dir, InsightsInputs{Window: "14d"})
	found := false
	for _, b := range again.Branches {
		found = found || b.Name == "gh-pages"
	}
	if !found {
		t.Fatal("as an ordinary branch, gh-pages should be an open branch again")
	}
}

func TestInsightsReleaseBranchWindow(t *testing.T) {
	fx := kindsFixture(t)
	a := newTestApp(t)
	rbs, err := a.InsightsReleaseBranches(fx.Dir, "")
	if err != nil {
		t.Fatal(err)
	}
	if len(rbs) != 2 || rbs[0].Name != "release/1.1" || rbs[1].Name != "release/1.0" {
		t.Fatalf("release branches = %+v, want release/1.1 then release/1.0", rbs)
	}
	res, err := a.InsightsShipped(fx.Dir, InsightsInputs{Window: "releases", From: "release/1.0", To: "release/1.1"})
	if err != nil {
		t.Fatal(err)
	}
	var titles []string
	for _, it := range res.Items {
		titles = append(titles, it.Title)
	}
	if !slices.Equal(titles, []string{"Feature two", "Feature one"}) {
		t.Fatalf("between the release branches = %v", titles)
	}
	if res.Window.Range != "release/1.0..release/1.1" || res.Window.Points == "" {
		t.Errorf("window = %+v", res.Window)
	}
	if _, err := a.InsightsShipped(fx.Dir, InsightsInputs{Window: "releases", From: "release/1.0", To: "main"}); err == nil {
		t.Error("main isn't a release, but was accepted as one")
	}
}
