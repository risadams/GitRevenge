package app

import (
	"strings"
	"testing"

	"github.com/risadams/gitrevenge/internal/gittest"
)

func TestCommitDetailAndDiff(t *testing.T) {
	a := newTestApp(t)
	fx := gittest.New(t)
	fx.CommitFile("a.txt", "one\n", "first")
	fx.CommitFile("a.txt", "one\ntwo\n", "second\n\nbody")
	tab, err := a.OpenRepo(fx.Dir)
	if err != nil {
		t.Fatal(err)
	}

	info, err := a.CommitDetail(tab.Path, fx.Head())
	if err != nil {
		t.Fatal(err)
	}
	if info.Commit.Subject != "second" || info.Commit.Body != "body" || len(info.Files) != 1 || info.Files[0].Path != "a.txt" {
		t.Fatalf("info = %+v", info)
	}
	d, err := a.CommitFileDiff(tab.Path, fx.Head(), "a.txt", "")
	if err != nil || !strings.Contains(d.Patch, "+two") {
		t.Fatalf("diff = %+v, %v", d, err)
	}
	if _, err := a.CommitDetail(tab.Path, strings.Repeat("0", 40)); err == nil || err.Error() != "That commit no longer exists." {
		t.Fatalf("missing commit err = %v", err)
	}
}
