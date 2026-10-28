package git

import (
	"bufio"
	"context"
	"errors"
	"io"
	"os/exec"
	"slices"
	"strings"
	"testing"

	"github.com/risadams/gitrevenge/internal/gittest"
)

func newTestRunner(t *testing.T) *Runner {
	t.Helper()
	bin, err := exec.LookPath("git")
	if err != nil {
		t.Skip("git not installed")
	}
	return NewRunner(bin)
}

func TestRunOutput(t *testing.T) {
	r := newTestRunner(t)
	repo := gittest.New(t)
	sha := repo.Linear(1)[0]

	got, err := r.Output(context.Background(), Cmd{Dir: repo.Dir, Args: []string{"rev-parse", "HEAD"}})
	if err != nil {
		t.Fatal(err)
	}
	if got != sha {
		t.Fatalf("rev-parse HEAD = %q, want %q", got, sha)
	}
}

func TestRunErrorCarriesStderrAndExitCode(t *testing.T) {
	r := newTestRunner(t)
	repo := gittest.New(t)

	_, err := r.Run(context.Background(), Cmd{Dir: repo.Dir, Args: []string{"rev-parse", "does-not-exist"}})
	var gerr *Error
	if !errors.As(err, &gerr) {
		t.Fatalf("err = %v, want *Error", err)
	}
	if gerr.ExitCode == 0 || !strings.Contains(gerr.Stderr, "does-not-exist") {
		t.Fatalf("unexpected error details: exit=%d stderr=%q", gerr.ExitCode, gerr.Stderr)
	}
}

func TestStreamReadsLines(t *testing.T) {
	r := newTestRunner(t)
	repo := gittest.New(t)
	repo.Linear(50)

	var lines int
	err := r.Stream(context.Background(), Cmd{Dir: repo.Dir, Args: []string{"log", "--format=%H"}, ReadOnly: true},
		func(rd io.Reader) error {
			sc := bufio.NewScanner(rd)
			for sc.Scan() {
				lines++
			}
			return sc.Err()
		})
	if err != nil {
		t.Fatal(err)
	}
	if lines != 50 {
		t.Fatalf("got %d lines, want 50", lines)
	}
}

func TestStreamStopsOnConsumerError(t *testing.T) {
	r := newTestRunner(t)
	repo := gittest.New(t)
	repo.Linear(20)

	stop := errors.New("stop")
	err := r.Stream(context.Background(), Cmd{Dir: repo.Dir, Args: []string{"log", "--format=%H"}},
		func(io.Reader) error { return stop })
	if !errors.Is(err, stop) {
		t.Fatalf("err = %v, want consumer error", err)
	}
}

func TestRunCanceled(t *testing.T) {
	r := newTestRunner(t)
	repo := gittest.New(t)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := r.Run(ctx, Cmd{Dir: repo.Dir, Args: []string{"status"}})
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("err = %v, want context.Canceled", err)
	}
}

func TestBaseEnv(t *testing.T) {
	env := baseEnv([]string{"PATH=/bin", "GIT_DIR=/elsewhere", "git_work_tree=/x", "LC_ALL=fr_FR", "HOME=/home/u"})
	for _, banned := range []string{"GIT_DIR=/elsewhere", "git_work_tree=/x", "LC_ALL=fr_FR"} {
		if slices.Contains(env, banned) {
			t.Errorf("env still contains %q", banned)
		}
	}
	for _, want := range []string{"PATH=/bin", "HOME=/home/u", "GIT_TERMINAL_PROMPT=0", "LC_ALL=C"} {
		if !slices.Contains(env, want) {
			t.Errorf("env missing %q", want)
		}
	}
}
