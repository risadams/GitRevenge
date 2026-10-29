package git

import (
	"context"
	"errors"
	"strings"
)

// Note is a git note attached to a commit, from one notes ref.
type Note struct {
	// Ref is the full notes ref, such as refs/notes/commits.
	Ref  string `json:"ref"`
	Text string `json:"text"`
}

// DefaultNotesRef is where git notes add writes unless told otherwise.
const DefaultNotesRef = "refs/notes/commits"

// Notes returns the commit's notes from every notes ref, the default ref first.
func (r *Repo) Notes(ctx context.Context, commit string) ([]Note, error) {
	out, err := r.ReadString(ctx, "for-each-ref", "--format=%(refname)", "refs/notes/")
	if err != nil {
		return nil, err
	}
	refs := strings.Fields(out)
	notes := []Note{}
	for _, ref := range refs {
		text, err := r.ReadString(ctx, "notes", "--ref="+ref, "show", commit)
		var gerr *Error
		if errors.As(err, &gerr) && strings.Contains(strings.ToLower(gerr.Stderr), "no note found") {
			continue
		}
		if err != nil {
			return nil, err
		}
		n := Note{Ref: ref, Text: strings.TrimRight(text, "\n")}
		if ref == DefaultNotesRef {
			notes = append([]Note{n}, notes...)
		} else {
			notes = append(notes, n)
		}
	}
	return notes, nil
}

// SetNote adds or replaces the commit's note in ref, or removes it when text is empty.
func (r *Repo) SetNote(ctx context.Context, ref, commit, text string) error {
	if !strings.HasPrefix(ref, "refs/notes/") {
		ref = "refs/notes/" + ref
	}
	if strings.TrimSpace(text) == "" {
		_, err := r.Write(ctx, Cmd{Args: []string{"notes", "--ref=" + ref, "remove", "--ignore-missing", commit}})
		return err
	}
	_, err := r.Write(ctx, Cmd{Args: []string{"notes", "--ref=" + ref, "add", "-f", "-F", "-", commit}, Stdin: strings.NewReader(text)})
	return err
}
