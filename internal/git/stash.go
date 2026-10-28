package git

import (
	"context"
	"errors"
	"fmt"
	"strings"
)

// ErrNothingToStash means there are no local changes to stash.
var ErrNothingToStash = errors.New("no local changes to stash")

// ErrStashConflict means applying a stash conflicted with the work tree. The changes that
// applied cleanly are in place; conflicted files are marked as such, and (for pop) the stash
// is kept.
var ErrStashConflict = errors.New("stash applied with conflicts")

// ErrStashNotFound means no stash entry has the given commit ID.
var ErrStashNotFound = errors.New("stash not found")

// StashSave stashes local changes with an optional message. With includeUntracked,
// untracked files are stashed too.
func (r *Repo) StashSave(ctx context.Context, message string, includeUntracked bool) error {
	st, err := r.Status(ctx)
	if err != nil {
		return err
	}
	changed := false
	for _, f := range st.Files {
		if !f.Untracked || includeUntracked {
			changed = true
			break
		}
	}
	if !changed {
		return ErrNothingToStash
	}
	args := []string{"stash", "push"}
	if includeUntracked {
		args = append(args, "--include-untracked")
	}
	if strings.TrimSpace(message) != "" {
		args = append(args, "-m", message)
	}
	_, err = r.Write(ctx, Cmd{Args: args})
	return err
}

// StashIDs returns the commit IDs of stash entries, newest (stash@{0}) first.
func (r *Repo) StashIDs(ctx context.Context) ([]string, error) {
	if _, err := r.Read(ctx, "rev-parse", "--verify", "-q", "refs/stash"); err != nil {
		return nil, nil
	}
	out, err := r.ReadString(ctx, "log", "-g", "--format=%H", "refs/stash", "--")
	if err != nil {
		return nil, err
	}
	return strings.Fields(out), nil
}

// stashRef finds the current stash@{n} for a stash commit. Entries renumber as stashes are
// added and dropped, so actions name stashes by commit ID and resolve just before running.
func (r *Repo) stashRef(ctx context.Context, id string) (string, error) {
	ids, err := r.StashIDs(ctx)
	if err != nil {
		return "", err
	}
	for i, s := range ids {
		if s == id {
			return fmt.Sprintf("stash@{%d}", i), nil
		}
	}
	return "", ErrStashNotFound
}

// StashApply applies a stash, keeping it. StashPop applies it and drops it on success.
func (r *Repo) StashApply(ctx context.Context, id string) error {
	return r.stashApply(ctx, id, "apply")
}

// StashPop applies a stash and drops it if it applied without conflicts.
func (r *Repo) StashPop(ctx context.Context, id string) error { return r.stashApply(ctx, id, "pop") }

func (r *Repo) stashApply(ctx context.Context, id, verb string) error {
	ref, err := r.stashRef(ctx, id)
	if err != nil {
		return err
	}
	_, err = r.Write(ctx, Cmd{Args: []string{"stash", verb, ref}})
	var gerr *Error
	if errors.As(err, &gerr) && gerr.ExitCode == 1 {
		// git exits 1 both for conflicts and for refusing to touch local changes; only the
		// former leaves unmerged paths behind.
		if st, serr := r.Status(ctx); serr == nil {
			for _, f := range st.Files {
				if f.Conflicted {
					return ErrStashConflict
				}
			}
		}
	}
	return err
}

// StashDrop deletes a stash entry.
func (r *Repo) StashDrop(ctx context.Context, id string) error {
	ref, err := r.stashRef(ctx, id)
	if err != nil {
		return err
	}
	_, err = r.Write(ctx, Cmd{Args: []string{"stash", "drop", "-q", ref}})
	return err
}
