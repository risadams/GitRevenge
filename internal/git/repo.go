package git

import (
	"context"
	"errors"
	"fmt"
	"io"
	"regexp"
	"time"
)

// Repo runs git commands against one repository. Mutating commands are serialized so the
// app never races itself; read commands run concurrently and never take optional locks.
type Repo struct {
	runner *Runner
	dir    string
	// writes is a one-slot semaphore. A channel rather than a mutex so waiting respects ctx.
	writes chan struct{}
}

// OpenRepo returns a Repo rooted at dir. It does not validate that dir is a repository.
func OpenRepo(r *Runner, dir string) *Repo {
	return &Repo{runner: r, dir: dir, writes: make(chan struct{}, 1)}
}

// Dir returns the repository's working directory.
func (r *Repo) Dir() string { return r.dir }

// Read runs a read-only git command and returns its stdout.
func (r *Repo) Read(ctx context.Context, args ...string) ([]byte, error) {
	return r.runner.Run(ctx, Cmd{Dir: r.dir, Args: args, ReadOnly: true})
}

// ReadString runs a read-only git command and returns trimmed stdout.
func (r *Repo) ReadString(ctx context.Context, args ...string) (string, error) {
	return r.runner.Output(ctx, Cmd{Dir: r.dir, Args: args, ReadOnly: true})
}

// Stream runs a read-only git command and passes its stdout to consume as it arrives.
func (r *Repo) Stream(ctx context.Context, consume func(io.Reader) error, args ...string) error {
	return r.runner.Stream(ctx, Cmd{Dir: r.dir, Args: args, ReadOnly: true}, consume)
}

// lockRetryDelays is how long Write waits before each retry while another process holds a
// git lock file. The total is about 1.5s, long enough for a typical CLI command to finish.
var lockRetryDelays = []time.Duration{
	100 * time.Millisecond,
	200 * time.Millisecond,
	400 * time.Millisecond,
	800 * time.Millisecond,
}

// Write runs a mutating git command. It waits for any earlier Write on this Repo to finish,
// and retries briefly if another process holds a git lock. If the lock persists it returns
// a *LockedError.
func (r *Repo) Write(ctx context.Context, c Cmd) ([]byte, error) {
	select {
	case r.writes <- struct{}{}:
	case <-ctx.Done():
		return nil, ctx.Err()
	}
	defer func() { <-r.writes }()

	c.Dir = r.dir
	c.ReadOnly = false
	for attempt := 0; ; attempt++ {
		out, err := r.runner.Run(ctx, c)
		lockPath, locked := lockFileFrom(err)
		if !locked {
			return out, err
		}
		if attempt == len(lockRetryDelays) {
			return out, &LockedError{Path: lockPath, Err: err}
		}
		select {
		case <-time.After(lockRetryDelays[attempt]):
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}
}

// ErrLocked matches any *LockedError via errors.Is.
var ErrLocked = errors.New("another git process is running")

// LockedError means a git lock file (usually index.lock) stayed held by another process.
type LockedError struct {
	Path string
	Err  error
}

func (e *LockedError) Error() string {
	return fmt.Sprintf("%s: %s exists. If no other git process is running, delete that file and try again", ErrLocked, e.Path)
}

func (e *LockedError) Unwrap() []error { return []error{ErrLocked, e.Err} }

// lockRE matches git's message when it cannot create a lock file, e.g.
// "fatal: Unable to create '/repo/.git/index.lock': File exists."
var lockRE = regexp.MustCompile(`Unable to create '([^']+\.lock)': File exists`)

func lockFileFrom(err error) (string, bool) {
	var gerr *Error
	if !errors.As(err, &gerr) {
		return "", false
	}
	m := lockRE.FindStringSubmatch(gerr.Stderr)
	if m == nil {
		return "", false
	}
	return m[1], true
}
