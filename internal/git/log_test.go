package git

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/risadams/gitrevenge/internal/gittest"
)

func readLog(t *testing.T, fx *gittest.Repo, opts LogOptions) []LogCommit {
	t.Helper()
	repo := OpenRepo(newTestRunner(t), fx.Dir)
	var out []LogCommit
	if err := repo.Log(context.Background(), opts, func(c LogCommit) error {
		out = append(out, c)
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	return out
}

func TestLogLinear(t *testing.T) {
	fx := gittest.New(t)
	shas := fx.Linear(3)

	got := readLog(t, fx, LogOptions{})
	if len(got) != 3 {
		t.Fatalf("got %d commits", len(got))
	}
	// Children first.
	for i, c := range got {
		if c.ID != shas[2-i] {
			t.Fatalf("commit %d = %s, want %s", i, c.ID, shas[2-i])
		}
	}
	if len(got[2].Parents) != 0 || got[0].Parents[0] != shas[1] {
		t.Fatalf("parents wrong: %v / %v", got[0].Parents, got[2].Parents)
	}
	if got[0].Subject != "commit 2" || got[0].AuthorName != "Test User" || got[0].AuthorEmail != "test@example.com" || got[0].AuthorTime == 0 {
		t.Fatalf("fields wrong: %+v", got[0])
	}
}

func TestLogIncludesAllBranchesAndMerges(t *testing.T) {
	fx := gittest.WithMergeHistory(t)
	got := readLog(t, fx, LogOptions{})
	if len(got) != 6 {
		t.Fatalf("got %d commits, want 6 (including unmerged topic)", len(got))
	}
	if len(got[0].Parents) != 2 && got[0].Subject != "topic 1" {
		t.Fatalf("expected the merge or topic tip first, got %+v", got[0])
	}
	// Topological order: every commit appears before its parents.
	pos := map[string]int{}
	for i, c := range got {
		pos[c.ID] = i
	}
	for _, c := range got {
		for _, p := range c.Parents {
			if pos[p] <= pos[c.ID] {
				t.Fatalf("parent %s listed before child %s", p, c.ID)
			}
		}
	}
}

func TestLogSubjectWithSpecialCharacters(t *testing.T) {
	fx := gittest.New(t)
	subject := "fix: handle ünïcødé, tabs\tand | pipes"
	fx.CommitFile("a.txt", "a", subject+"\n\nbody line\nmore body")
	got := readLog(t, fx, LogOptions{})
	if got[0].Subject != subject {
		t.Fatalf("subject = %q", got[0].Subject)
	}
}

func TestLogEmptyRepo(t *testing.T) {
	if got := readLog(t, gittest.New(t), LogOptions{}); len(got) != 0 {
		t.Fatalf("empty repo returned %d commits", len(got))
	}
}

func TestLogDetachedHeadIncluded(t *testing.T) {
	fx := gittest.New(t)
	fx.Linear(2)
	fx.Git("checkout", "-q", "--detach")
	detached := fx.CommitFile("d.txt", "d", "detached work")
	got := readLog(t, fx, LogOptions{})
	if got[0].ID != detached {
		t.Fatalf("detached commit missing from log; first = %s", got[0].Subject)
	}
}

func TestLogMaxCountAndStop(t *testing.T) {
	fx := gittest.New(t)
	fx.Linear(10)
	if got := readLog(t, fx, LogOptions{MaxCount: 4}); len(got) != 4 {
		t.Fatalf("MaxCount: got %d", len(got))
	}

	repo := OpenRepo(newTestRunner(t), fx.Dir)
	stop := errors.New("enough")
	n := 0
	err := repo.Log(context.Background(), LogOptions{}, func(LogCommit) error {
		n++
		if n == 3 {
			return stop
		}
		return nil
	})
	if !errors.Is(err, stop) || n != 3 {
		t.Fatalf("err = %v after %d commits", err, n)
	}
}

func TestParseLogRecordRejectsGarbage(t *testing.T) {
	if _, err := parseLogRecord([]byte("not a record")); err == nil || !strings.Contains(err.Error(), "unexpected") {
		t.Fatalf("err = %v", err)
	}
}

func TestLogOrphanBranchKeepsOtherHistory(t *testing.T) {
	fx := gittest.New(t)
	fx.Linear(2)
	fx.Git("checkout", "-q", "--orphan", "fresh")
	if got := readLog(t, fx, LogOptions{}); len(got) != 2 {
		t.Fatalf("got %d commits, want main's 2 while HEAD is unborn", len(got))
	}
}

func TestLogVeryLongSubject(t *testing.T) {
	fx := gittest.New(t)
	fx.Linear(1)
	long := strings.Repeat("x", 200*1024) // larger than the 64KB read buffer
	// Via a file: a 200KB -m argument exceeds the Windows command-line limit.
	msg := filepath.Join(t.TempDir(), "msg")
	if err := os.WriteFile(msg, []byte(long), 0o644); err != nil {
		t.Fatal(err)
	}
	fx.Git("commit", "--allow-empty", "-q", "-F", msg)
	fx.Linear(1)
	got := readLog(t, fx, LogOptions{})
	if len(got) != 3 || got[1].Subject != long {
		t.Fatalf("got %d commits; long subject intact = %v", len(got), len(got) > 1 && got[1].Subject == long)
	}
}

func TestCountCommitsMatchesLog(t *testing.T) {
	fx := gittest.New(t)
	repo := OpenRepo(newTestRunner(t), fx.Dir)
	if n, err := repo.CountCommits(context.Background(), LogOptions{}); err != nil || n != 0 {
		t.Fatalf("empty repo count = %d, %v", n, err)
	}
	fx.Linear(3)
	fx.Branch("side")
	fx.Checkout("side")
	fx.CommitFile("side.txt", "x", "side")
	fx.Checkout("main")
	n, err := repo.CountCommits(context.Background(), LogOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if want := len(readLog(t, fx, LogOptions{})); n != want {
		t.Fatalf("count = %d, log has %d", n, want)
	}
}

func TestLogIncludesTagsOutsideBranches(t *testing.T) {
	fx := gittest.New(t)
	fx.Linear(2)
	fx.Tag("v1", "")
	fx.Branch("side")
	fx.Checkout("side")
	lone := fx.CommitFile("side.txt", "x", "only reachable from a tag")
	fx.Tag("lone", "annotated")
	fx.Checkout("main")
	fx.Git("branch", "-D", "side")

	ids := map[string]bool{}
	for _, c := range readLog(t, fx, LogOptions{}) {
		ids[c.ID] = true
	}
	if !ids[lone] || len(ids) != 3 {
		t.Fatalf("log has %d commits (tag-only commit included: %v), want 3 with it", len(ids), ids[lone])
	}
	repo := OpenRepo(newTestRunner(t), fx.Dir)
	if n, err := repo.CountCommits(context.Background(), LogOptions{}); err != nil || n != 3 {
		t.Fatalf("count = %d, %v; want 3", n, err)
	}
}
