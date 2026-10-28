package app

import (
	"context"

	"github.com/risadams/gitrevenge/internal/git"
)

// Diff modes for FileContents.
const (
	// DiffCommit compares a commit's file with its first parent.
	DiffCommit = "commit"
	// DiffStaged compares HEAD with the index.
	DiffStaged = "staged"
	// DiffUnstaged compares the index with the work tree.
	DiffUnstaged = "unstaged"
)

// FileContents returns both versions of a file for the diff viewer. commit is used in
// commit mode; oldPath is the source of a rename, if any. force loads files over the
// normal size limit.
func (a *App) FileContents(path, mode, commit, file, oldPath string, force bool) (git.FileContents, error) {
	r, err := a.openRepo(path)
	if err != nil {
		return git.FileContents{}, err
	}
	if oldPath == "" {
		oldPath = file
	}
	ctx := context.Background()
	var from, to git.FileRef
	switch mode {
	case DiffCommit:
		c, err := git.ReadCommit(r.Objects, commit)
		if err != nil {
			return git.FileContents{}, err
		}
		if len(c.Parents) > 0 {
			from = git.FileRef{Rev: c.Parents[0], Path: oldPath}
		}
		to = git.FileRef{Rev: c.ID, Path: file}
	case DiffStaged:
		from, to = git.FileRef{Rev: "HEAD", Path: oldPath}, git.IndexVersion(file)
	case DiffUnstaged:
		from, to = git.IndexVersion(file), git.WorkTreeVersion(file)
	default:
		return git.FileContents{}, userError("Unknown diff mode " + mode + ".")
	}
	return r.Git.Contents(ctx, file, from, to, force)
}

// errChanged is returned when a file changed after it was displayed.
const errChanged = userError("The file changed since it was shown. The view has been refreshed; please try again.")

// SetStagedContent replaces a file's staged content with content (in work-tree form). The
// UI computes it from the displayed diff to stage or unstage hunks and lines. expected is
// the staged content the UI was showing; if the index has changed since, nothing is written.
func (a *App) SetStagedContent(path, file, expected, content string) error {
	r, err := a.openRepo(path)
	if err != nil {
		return err
	}
	ctx := context.Background()
	cur, err := r.Git.Contents(ctx, file, git.FileRef{}, git.IndexVersion(file), true)
	if err != nil {
		return err
	}
	// A file not in the index (untracked) is shown as empty.
	if cur.New.Text != expected {
		return errChanged
	}
	return writeErr(r.Git.SetIndexContent(ctx, file, content))
}

// SetWorkTreeContent overwrites a file on disk to discard hunks or lines. expected is the
// content the UI was showing; if the file has changed since, nothing is written, so newer
// edits are never lost.
func (a *App) SetWorkTreeContent(path, file, expected, content string) error {
	r, err := a.openRepo(path)
	if err != nil {
		return err
	}
	cur, err := r.Git.Contents(context.Background(), file, git.FileRef{}, git.WorkTreeVersion(file), true)
	if err != nil {
		return err
	}
	if cur.New.Text != expected {
		return errChanged
	}
	return r.Git.SetWorkTreeContent(file, content)
}
