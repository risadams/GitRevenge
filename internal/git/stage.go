package git

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
)

// pathspecCmd builds a mutating command that reads its paths from stdin, NUL-separated.
// That avoids command-line length limits with many files, and GIT_LITERAL_PATHSPECS stops
// names containing *, ? or [ from being treated as patterns.
func pathspecCmd(paths []string, args ...string) Cmd {
	return Cmd{
		Args:  append(args, "--pathspec-from-file=-", "--pathspec-file-nul"),
		Stdin: strings.NewReader(strings.Join(paths, "\x00") + "\x00"),
		Env:   []string{"GIT_LITERAL_PATHSPECS=1"},
	}
}

// hasHead reports whether HEAD points at a commit (false in a repository with no commits).
func (r *Repo) hasHead(ctx context.Context) bool {
	_, err := r.Read(ctx, "rev-parse", "--verify", "-q", "HEAD")
	return err == nil
}

// Stage adds the given paths' changes, including deletions, to the index.
func (r *Repo) Stage(ctx context.Context, paths []string) error {
	if len(paths) == 0 {
		return nil
	}
	_, err := r.Write(ctx, pathspecCmd(paths, "add", "-A"))
	return err
}

// StageAll adds every change in the work tree to the index.
func (r *Repo) StageAll(ctx context.Context) error {
	_, err := r.Write(ctx, Cmd{Args: []string{"add", "-A"}})
	return err
}

// Unstage removes the given paths' changes from the index, keeping the work tree.
func (r *Repo) Unstage(ctx context.Context, paths []string) error {
	if len(paths) == 0 {
		return nil
	}
	if !r.hasHead(ctx) {
		// Before the first commit there is nothing to restore from; drop the index entries.
		_, err := r.Write(ctx, pathspecCmd(paths, "rm", "--cached", "-r", "-q", "--ignore-unmatch"))
		return err
	}
	_, err := r.Write(ctx, pathspecCmd(paths, "restore", "--staged"))
	return err
}

// UnstageAll empties the staging area, keeping the work tree.
func (r *Repo) UnstageAll(ctx context.Context) error {
	if !r.hasHead(ctx) {
		_, err := r.Write(ctx, Cmd{Args: []string{"rm", "--cached", "-r", "-q", "--ignore-unmatch", "."}})
		return err
	}
	_, err := r.Write(ctx, Cmd{Args: []string{"reset", "-q"}})
	return err
}

// ErrConflicted means a conflicted file was asked to be discarded; it must be resolved.
var ErrConflicted = errors.New("file has merge conflicts")

// Discard throws away unstaged changes: tracked files are restored from the index (staged
// changes are kept) and untracked files are deleted. This cannot be undone.
func (r *Repo) Discard(ctx context.Context, files []FileStatus) error {
	var tracked []string
	for _, f := range files {
		switch {
		case f.Conflicted:
			return fmt.Errorf("%s: %w", f.Path, ErrConflicted)
		case f.Untracked:
			if err := r.removeUntracked(f.Path); err != nil {
				return err
			}
		case f.WorkTree != ".":
			tracked = append(tracked, f.Path)
		}
	}
	if len(tracked) == 0 {
		return nil
	}
	_, err := r.Write(ctx, pathspecCmd(tracked, "restore", "--worktree"))
	return err
}

// removeUntracked deletes an untracked file, refusing anything outside the work tree.
func (r *Repo) removeUntracked(rel string) error {
	p, err := r.inside(rel)
	if err != nil {
		return fmt.Errorf("refusing to delete: %w", err)
	}
	if err := os.Remove(p); err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}
	return nil
}

// ErrNothingStaged means a commit was requested with an empty staging area.
var ErrNothingStaged = errors.New("nothing staged to commit")

// Commit records the staged changes with msg and returns the new commit ID. With amend,
// it replaces the last commit instead. Hooks run, and commit signing follows git config.
func (r *Repo) Commit(ctx context.Context, msg string, amend bool) (string, error) {
	if strings.TrimSpace(msg) == "" {
		return "", errors.New("commit message is empty")
	}
	// git reports "nothing to commit" on stdout with an empty stderr; check up front instead.
	// Exit status 1 from diff --quiet means something is staged.
	if !amend {
		if _, err := r.Read(ctx, "diff", "--cached", "--quiet"); err == nil {
			return "", ErrNothingStaged
		}
	}
	// --cleanup=whitespace keeps lines starting with '#', which are common in messages
	// ("#123 fix ..."); the default would strip them as comments.
	args := []string{"commit", "-F", "-", "--cleanup=whitespace"}
	if amend {
		args = append(args, "--amend")
	}
	_, err := r.Write(ctx, Cmd{
		Args:  args,
		Stdin: strings.NewReader(msg),
		// Never open an editor; the message comes from the app.
		Env: []string{"GIT_EDITOR=true"},
	})
	if err != nil {
		return "", err
	}
	return r.ReadString(ctx, "rev-parse", "HEAD")
}

// LastCommitMessage returns HEAD's full message, for amending.
func (r *Repo) LastCommitMessage(ctx context.Context) (string, error) {
	return r.ReadString(ctx, "log", "-1", "--format=%B")
}

// Signing describes whether new commits will be signed, per git config.
type Signing struct {
	Enabled bool `json:"enabled"`
	// Format is "openpgp", "ssh" or "x509".
	Format string `json:"format"`
	Key    string `json:"key"`
}

// CommitSigning reads commit.gpgsign, gpg.format and user.signingkey.
func (r *Repo) CommitSigning(ctx context.Context) Signing {
	get := func(key string) string {
		v, _ := r.ReadString(ctx, "config", "--get", key)
		return v
	}
	s := Signing{Format: get("gpg.format"), Key: get("user.signingkey")}
	s.Enabled, _ = parseBool(get("commit.gpgsign"))
	if s.Format == "" {
		s.Format = "openpgp"
	}
	return s
}

// Identity is who new commits are recorded as, per git config.
type Identity struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}

// CommitIdentity reads user.name and user.email.
func (r *Repo) CommitIdentity(ctx context.Context) Identity {
	get := func(key string) string {
		v, _ := r.ReadString(ctx, "config", "--get", key)
		return v
	}
	return Identity{Name: get("user.name"), Email: get("user.email")}
}

func parseBool(v string) (bool, error) {
	switch strings.ToLower(v) {
	case "true", "yes", "on", "1":
		return true, nil
	case "", "false", "no", "off", "0":
		return false, nil
	}
	return false, fmt.Errorf("not a boolean: %q", v)
}
