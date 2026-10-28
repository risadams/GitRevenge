// Package git runs the system git CLI (ADR-0002).
package git

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"time"
)

// baseArgs are prepended to every invocation so output is stable and machine-readable.
var baseArgs = []string{
	"-c", "core.quotepath=off",
	"-c", "color.ui=false",
	"-c", "core.pager=cat",
}

// Runner executes git commands with a fixed binary and a safe environment.
type Runner struct {
	bin string
	env []string
}

// NewRunner returns a Runner for the git binary at bin.
func NewRunner(bin string) *Runner {
	return &Runner{bin: bin, env: baseEnv(os.Environ())}
}

// baseEnv strips variables that would redirect git to another repo and adds the
// settings every call needs. Git must never block on a terminal prompt (ADR-0004).
func baseEnv(environ []string) []string {
	env := make([]string, 0, len(environ)+3)
	for _, kv := range environ {
		name, _, _ := strings.Cut(kv, "=")
		switch strings.ToUpper(name) {
		case "GIT_DIR", "GIT_WORK_TREE", "GIT_INDEX_FILE", "GIT_OBJECT_DIRECTORY",
			"GIT_NAMESPACE", "GIT_CEILING_DIRECTORIES", "LC_ALL", "LANG":
			continue
		}
		env = append(env, kv)
	}
	return append(env, "GIT_TERMINAL_PROMPT=0", "LC_ALL=C", "LANG=C")
}

// Bin returns the path of the git binary.
func (r *Runner) Bin() string { return r.bin }

// Cmd describes a single git invocation.
type Cmd struct {
	// Dir is the working directory; usually the repository root.
	Dir string
	// Args are the git arguments, without the binary name.
	Args []string
	// Env holds extra KEY=VALUE pairs, applied after the runner's environment.
	Env []string
	// Stdin, if set, is fed to the process.
	Stdin io.Reader
	// ReadOnly sets GIT_OPTIONAL_LOCKS=0 so reads such as status never take the index
	// lock and race with the user's own git commands.
	ReadOnly bool
}

// Run executes c and returns its stdout.
func (r *Runner) Run(ctx context.Context, c Cmd) ([]byte, error) {
	var stdout bytes.Buffer
	if err := r.run(ctx, c, &stdout); err != nil {
		return stdout.Bytes(), err
	}
	return stdout.Bytes(), nil
}

// Output executes c and returns stdout with surrounding whitespace trimmed.
func (r *Runner) Output(ctx context.Context, c Cmd) (string, error) {
	out, err := r.Run(ctx, c)
	return strings.TrimSpace(string(out)), err
}

// Stream executes c and passes its stdout to consume as it is produced. If consume returns
// an error, the process is killed and that error is returned.
func (r *Runner) Stream(ctx context.Context, c Cmd, consume func(io.Reader) error) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	pr, pw := io.Pipe()
	done := make(chan error, 1)
	go func() {
		err := r.run(ctx, c, pw)
		pw.CloseWithError(err)
		done <- err
	}()

	consumeErr := consume(bufio.NewReaderSize(pr, 64*1024))
	if consumeErr != nil {
		cancel()
	}
	// Drain so the process can exit if consume stopped early without error.
	_, _ = io.Copy(io.Discard, pr)
	runErr := <-done
	if consumeErr != nil {
		return consumeErr
	}
	return runErr
}

func (r *Runner) run(ctx context.Context, c Cmd, stdout io.Writer) error {
	args := append(append(make([]string, 0, len(baseArgs)+len(c.Args)), baseArgs...), c.Args...)
	cmd := exec.CommandContext(ctx, r.bin, args...)
	cmd.Dir = c.Dir
	cmd.Env = append(append(make([]string, 0, len(r.env)+len(c.Env)+1), r.env...), c.Env...)
	if c.ReadOnly {
		cmd.Env = append(cmd.Env, "GIT_OPTIONAL_LOCKS=0")
	}
	cmd.Stdin = c.Stdin
	cmd.Stdout = stdout
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	hideWindow(cmd)
	// After cancellation, don't wait forever on pipes held open by git's own children
	// (e.g. git-remote-https).
	cmd.WaitDelay = 2 * time.Second

	err := cmd.Run()
	if err == nil {
		return nil
	}
	if ctx.Err() != nil {
		return ctx.Err()
	}
	gerr := &Error{Args: c.Args, Stderr: strings.TrimSpace(stderr.String()), Err: err}
	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) {
		gerr.ExitCode = exitErr.ExitCode()
	}
	return gerr
}

// Error is returned when git exits unsuccessfully.
type Error struct {
	Args     []string
	ExitCode int
	Stderr   string
	Err      error
}

func (e *Error) Error() string {
	msg := e.Stderr
	if msg == "" {
		msg = e.Err.Error()
	}
	return fmt.Sprintf("git %s: %s", strings.Join(e.Args, " "), msg)
}

func (e *Error) Unwrap() error { return e.Err }
