package session

import (
	"context"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/risadams/gitrevenge/internal/fsutil"
	"github.com/risadams/gitrevenge/internal/git"
	"github.com/risadams/gitrevenge/internal/gittest"
	"github.com/risadams/gitrevenge/internal/settings"
)

func newManager(t *testing.T) (*Manager, *settings.Store) {
	t.Helper()
	bin, err := exec.LookPath("git")
	if err != nil {
		t.Skip("git not installed")
	}
	store, err := settings.Open(filepath.Join(t.TempDir(), "settings.json"))
	if err != nil {
		t.Fatal(err)
	}
	m := NewManager(git.NewRunner(bin), store)
	t.Cleanup(m.CloseAll)
	return m, store
}

// realPath resolves symlinks (macOS temp dirs live under a /var -> /private/var symlink,
// and git reports the resolved path).
func realPath(t *testing.T, p string) string {
	t.Helper()
	r, err := filepath.EvalSymlinks(p)
	if err != nil {
		t.Fatal(err)
	}
	return r
}

func TestOpenFromSubdirectoryResolvesRoot(t *testing.T) {
	m, store := newManager(t)
	fx := gittest.New(t)
	fx.CommitFile("src/deep/file.txt", "x", "add")

	r, err := m.Open(context.Background(), fx.Path("src/deep"))
	if err != nil {
		t.Fatal(err)
	}
	if !fsutil.SamePath(r.Path, realPath(t, fx.Dir)) {
		t.Fatalf("Path = %s, want %s", r.Path, fx.Dir)
	}
	s := store.Get()
	if len(s.OpenTabs) != 1 || s.ActiveTab != r.Path || len(s.Recent) != 1 {
		t.Fatalf("settings not updated: %+v", s)
	}
}

func TestOpenTwiceReturnsSameRepo(t *testing.T) {
	m, store := newManager(t)
	fx := gittest.New(t)
	a, err := m.Open(context.Background(), fx.Dir)
	if err != nil {
		t.Fatal(err)
	}
	b, err := m.Open(context.Background(), fx.Dir)
	if err != nil {
		t.Fatal(err)
	}
	if a != b || len(m.List()) != 1 || len(store.Get().OpenTabs) != 1 {
		t.Fatal("opening the same repo twice must reuse the tab")
	}
}

func TestOpenRejectsNonRepoAndBare(t *testing.T) {
	m, _ := newManager(t)
	plain := t.TempDir()
	if _, err := m.Open(context.Background(), plain); !errors.Is(err, ErrNotRepo) {
		t.Fatalf("plain dir: err = %v, want ErrNotRepo", err)
	}
	bare := gittest.NewBare(t)
	if _, err := m.Open(context.Background(), bare.Dir); !errors.Is(err, ErrBare) {
		t.Fatalf("bare repo: err = %v, want ErrBare", err)
	}
	if _, err := m.Open(context.Background(), filepath.Join(plain, "missing")); err == nil {
		t.Fatal("missing folder should fail")
	}
}

func TestCloseMovesActiveToNeighbour(t *testing.T) {
	m, store := newManager(t)
	var paths []string
	for range 3 {
		r, err := m.Open(context.Background(), gittest.New(t).Dir)
		if err != nil {
			t.Fatal(err)
		}
		paths = append(paths, r.Path)
	}
	if err := m.SetActive(paths[1]); err != nil {
		t.Fatal(err)
	}
	if err := m.Close(paths[1]); err != nil {
		t.Fatal(err)
	}
	s := store.Get()
	if len(s.OpenTabs) != 2 || s.ActiveTab != paths[2] {
		t.Fatalf("after close: tabs=%v active=%s, want active %s", s.OpenTabs, s.ActiveTab, paths[2])
	}
	if len(s.Recent) != 3 {
		t.Fatal("closing a tab must not remove it from recents")
	}
}

func TestRestoreDropsMissingRepos(t *testing.T) {
	m, store := newManager(t)
	keep := gittest.New(t)
	gone := gittest.New(t)
	k, _ := m.Open(context.Background(), keep.Dir)
	g, _ := m.Open(context.Background(), gone.Dir)
	m.CloseAll()
	if err := os.RemoveAll(gone.Dir); err != nil {
		t.Fatal(err)
	}

	active, failed := m.Restore(context.Background())
	if len(failed) != 1 || failed[g.Path] == nil {
		t.Fatalf("failed = %v, want only %s", failed, g.Path)
	}
	if active != k.Path {
		t.Fatalf("active = %s, want %s (saved active tab was removed)", active, k.Path)
	}
	if s := store.Get(); len(s.OpenTabs) != 1 || s.OpenTabs[0] != k.Path {
		t.Fatalf("tabs after restore = %v", s.OpenTabs)
	}
}

func TestObjectsReaderIsWarm(t *testing.T) {
	m, _ := newManager(t)
	fx := gittest.New(t)
	sha := fx.Linear(1)[0]
	r, err := m.Open(context.Background(), fx.Dir)
	if err != nil {
		t.Fatal(err)
	}
	info, err := r.Objects.Info("HEAD")
	if err != nil || info.ID != sha {
		t.Fatalf("Objects.Info(HEAD) = %+v, %v", info, err)
	}
}

func TestRestoreKeepsStartPage(t *testing.T) {
	m, store := newManager(t)
	if _, err := m.Open(context.Background(), gittest.New(t).Dir); err != nil {
		t.Fatal(err)
	}
	if err := m.SetActive(""); err != nil {
		t.Fatal(err)
	}
	m.CloseAll()

	active, failed := m.Restore(context.Background())
	if active != "" || len(failed) != 0 || store.Get().ActiveTab != "" {
		t.Fatalf("active = %q, failed = %v; the start page should stay selected", active, failed)
	}
}
