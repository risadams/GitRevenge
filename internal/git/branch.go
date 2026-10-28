package git

import (
	"context"
	"errors"
	"fmt"
	"strings"
)

// ErrInvalidName means a branch or tag name isn't allowed by git.
var ErrInvalidName = errors.New("invalid name")

// ErrNotMerged means a branch has commits not merged anywhere; deleting it needs force.
var ErrNotMerged = errors.New("branch is not fully merged")

// ErrCurrentBranch means the operation can't apply to the checked-out branch.
var ErrCurrentBranch = errors.New("branch is checked out")

// CheckBranchName validates a branch name with git's own rules.
func (r *Repo) CheckBranchName(ctx context.Context, name string) error {
	if strings.TrimSpace(name) == "" || strings.HasPrefix(name, "-") {
		return fmt.Errorf("%q: %w", name, ErrInvalidName)
	}
	if _, err := r.Read(ctx, "check-ref-format", "--branch", name); err != nil {
		return fmt.Errorf("%q: %w", name, ErrInvalidName)
	}
	return nil
}

// CreateBranch creates a branch at start (a commit, branch or tag; empty means HEAD) and
// optionally checks it out.
func (r *Repo) CreateBranch(ctx context.Context, name, start string, checkout bool) error {
	if err := r.CheckBranchName(ctx, name); err != nil {
		return err
	}
	if start == "" {
		start = "HEAD"
	}
	args := []string{"branch", "--no-track", name, start}
	if checkout {
		args = []string{"switch", "--no-track", "-c", name, start}
	}
	_, err := r.Write(ctx, Cmd{Args: args})
	return err
}

// RenameBranch renames a local branch, keeping its upstream and reflog.
func (r *Repo) RenameBranch(ctx context.Context, from, to string) error {
	if err := r.CheckBranchName(ctx, to); err != nil {
		return err
	}
	_, err := r.Write(ctx, Cmd{Args: []string{"branch", "-m", from, to}})
	return err
}

// DeleteBranch deletes a local branch. Without force, git refuses branches whose commits
// would be lost, reported as ErrNotMerged.
func (r *Repo) DeleteBranch(ctx context.Context, name string, force bool) error {
	flag := "-d"
	if force {
		flag = "-D"
	}
	_, err := r.Write(ctx, Cmd{Args: []string{"branch", flag, name}})
	var gerr *Error
	if errors.As(err, &gerr) {
		switch {
		case strings.Contains(gerr.Stderr, "not fully merged"):
			return ErrNotMerged
		case strings.Contains(gerr.Stderr, "checked out"), strings.Contains(gerr.Stderr, "used by worktree"):
			return ErrCurrentBranch
		}
	}
	return err
}

// Checkout switches to ref. A local branch is checked out directly. A remote-tracking
// branch (e.g. "origin/feature") checks out the local branch of the same name, creating it
// to track the remote if needed. Anything else (a commit or tag) detaches HEAD.
func (r *Repo) Checkout(ctx context.Context, ref string) error {
	if r.refExists(ctx, "refs/heads/"+ref) {
		_, err := r.Write(ctx, Cmd{Args: []string{"switch", ref}})
		return err
	}
	if r.refExists(ctx, "refs/remotes/"+ref) {
		_, local, ok := strings.Cut(ref, "/")
		if !ok {
			return fmt.Errorf("%q: %w", ref, ErrInvalidName)
		}
		if r.refExists(ctx, "refs/heads/"+local) {
			_, err := r.Write(ctx, Cmd{Args: []string{"switch", local}})
			return err
		}
		_, err := r.Write(ctx, Cmd{Args: []string{"switch", "-c", local, "--track", ref}})
		return err
	}
	_, err := r.Write(ctx, Cmd{Args: []string{"switch", "--detach", ref}})
	return err
}

func (r *Repo) refExists(ctx context.Context, full string) bool {
	_, err := r.Read(ctx, "show-ref", "--verify", "--quiet", full)
	return err == nil
}

// SetUpstream makes branch track upstream (e.g. "origin/main"). An empty upstream unsets it.
func (r *Repo) SetUpstream(ctx context.Context, branch, upstream string) error {
	args := []string{"branch", "--set-upstream-to=" + upstream, branch}
	if upstream == "" {
		args = []string{"branch", "--unset-upstream", branch}
	}
	_, err := r.Write(ctx, Cmd{Args: args})
	return err
}

// CreateTag creates a tag at target (empty means HEAD). With a message the tag is annotated
// (and signed if tag.gpgSign is set); without one it is lightweight.
func (r *Repo) CreateTag(ctx context.Context, name, target, message string) error {
	if strings.TrimSpace(name) == "" || strings.HasPrefix(name, "-") {
		return fmt.Errorf("%q: %w", name, ErrInvalidName)
	}
	if _, err := r.Read(ctx, "check-ref-format", "refs/tags/"+name); err != nil {
		return fmt.Errorf("%q: %w", name, ErrInvalidName)
	}
	if target == "" {
		target = "HEAD"
	}
	args := []string{"tag", name, target}
	var stdin *strings.Reader
	if strings.TrimSpace(message) != "" {
		args = []string{"tag", "-a", "-F", "-", "--cleanup=whitespace", name, target}
		stdin = strings.NewReader(message)
	}
	c := Cmd{Args: args, Env: []string{"GIT_EDITOR=true"}}
	if stdin != nil {
		c.Stdin = stdin
	}
	_, err := r.Write(ctx, c)
	return err
}

// DeleteTag deletes a local tag.
func (r *Repo) DeleteTag(ctx context.Context, name string) error {
	_, err := r.Write(ctx, Cmd{Args: []string{"tag", "-d", name}})
	return err
}
