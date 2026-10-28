package app

import (
	"errors"
	"os"
	"testing"

	"github.com/risadams/gitrevenge/internal/gittest"
)

func TestFileContentsModes(t *testing.T) {
	a := newTestApp(t)
	fx := gittest.New(t)
	fx.CommitFile("a.txt", "v1\n", "first")
	fx.CommitFile("a.txt", "v2\n", "second")
	tab, err := a.OpenRepo(fx.Dir)
	if err != nil {
		t.Fatal(err)
	}
	fx.WriteFile("a.txt", "v3\n")
	fx.Git("add", "a.txt")
	fx.WriteFile("a.txt", "v4\n")

	cases := []struct {
		mode, commit, old, new string
	}{
		{DiffCommit, fx.Head(), "v1\n", "v2\n"},
		{DiffStaged, "", "v2\n", "v3\n"},
		{DiffUnstaged, "", "v3\n", "v4\n"},
	}
	for _, c := range cases {
		fc, err := a.FileContents(tab.Path, c.mode, c.commit, "a.txt", "", false)
		if err != nil || fc.Old.Text != c.old || fc.New.Text != c.new {
			t.Errorf("%s: old %q new %q err %v", c.mode, fc.Old.Text, fc.New.Text, err)
		}
	}
	root := fx.Git("rev-list", "--max-parents=0", "HEAD")
	if fc, _ := a.FileContents(tab.Path, DiffCommit, root, "a.txt", "", false); fc.Old.Exists || fc.New.Text != "v1\n" {
		t.Fatalf("root commit = %+v", fc)
	}
}

func TestPartialStagingGuards(t *testing.T) {
	a := newTestApp(t)
	fx := gittest.New(t)
	fx.CommitFile("a.txt", "one\ntwo\n", "base")
	tab, _ := a.OpenRepo(fx.Dir)
	fx.WriteFile("a.txt", "ONE\nTWO\n")

	// Stage the first line only, based on what the UI saw.
	if err := a.SetStagedContent(tab.Path, "a.txt", "one\ntwo\n", "ONE\ntwo\n"); err != nil {
		t.Fatal(err)
	}
	if got := fx.Git("show", ":a.txt"); got != "ONE\ntwo" {
		t.Fatalf("index = %q", got)
	}
	// A stale view (index changed since) is refused.
	if err := a.SetStagedContent(tab.Path, "a.txt", "one\ntwo\n", "x\n"); !errors.Is(err, errChanged) {
		t.Fatalf("stale stage err = %v", err)
	}

	// Discard the second line's change in the work tree.
	if err := a.SetWorkTreeContent(tab.Path, "a.txt", "ONE\nTWO\n", "ONE\ntwo\n"); err != nil {
		t.Fatal(err)
	}
	// The file changed on disk since the view: refuse rather than clobber.
	fx.WriteFile("a.txt", "edited elsewhere\n")
	if err := a.SetWorkTreeContent(tab.Path, "a.txt", "ONE\ntwo\n", "x\n"); !errors.Is(err, errChanged) {
		t.Fatalf("stale discard err = %v", err)
	}
	if b, _ := os.ReadFile(fx.Path("a.txt")); string(b) != "edited elsewhere\n" {
		t.Fatal("newer edits must not be overwritten")
	}
}
