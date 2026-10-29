package app

import (
	"context"
	"errors"
	"strconv"

	"github.com/risadams/gitrevenge/internal/git"
)

// Integrate merges, cherry-picks, rebases, resets or reverts on the current branch:
//
//	merge:       ref, mode ("" / "no-ff" / "ff-only")
//	cherry-pick: commit IDs, oldest first
//	rebase:      onto
//	reset:       target, mode ("soft" / "mixed" / "hard")
//	revert:      commit, mainline parent for merges ("" for ordinary commits)
//
// When the operation stops with conflicts, the in-progress banner takes over.
func (a *App) Integrate(path, action string, args []string) error {
	r, err := a.openRepo(path)
	if err != nil {
		return err
	}
	ctx := context.Background()
	arg := func(i int) string {
		if i < len(args) {
			return args[i]
		}
		return ""
	}
	switch action {
	case "merge":
		err = r.Git.Merge(ctx, arg(0), git.MergeMode(arg(1)))
	case "cherry-pick":
		if len(args) == 0 {
			return userError("Choose at least one commit to cherry-pick.")
		}
		err = r.Git.CherryPick(ctx, args...)
	case "rebase":
		err = r.Git.Rebase(ctx, arg(0))
	case "reset":
		err = r.Git.Reset(ctx, arg(0), git.ResetMode(arg(1)))
	case "revert":
		mainline, _ := strconv.Atoi(arg(1))
		err = r.Git.Revert(ctx, arg(0), mainline)
	default:
		return userError("Unknown action " + action + ".")
	}
	if errors.Is(err, git.ErrConflicts) {
		return userError("There are conflicts to resolve. Fix the conflicted files, stage them, then use Continue in the banner, or Abort to undo.")
	}
	var gerr *git.Error
	if errors.As(err, &gerr) && gerr.Stderr != "" {
		return gitMessage(gerr.Stderr)
	}
	return writeErr(err)
}
