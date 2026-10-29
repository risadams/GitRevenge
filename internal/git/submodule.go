package git

import (
	"context"
	"strings"
)

// Submodule is one submodule and its state.
type Submodule struct {
	Path string `json:"path"`
	// Commit is the commit currently checked out (or recorded, when not initialized).
	Commit string `json:"commit"`
	// State is "ok", "uninitialized", "modified" (checked-out commit differs from the one
	// recorded in the superproject) or "conflict".
	State string `json:"state"`
	// Describe is git's description of the commit, e.g. "v1.2-3-gabc" or "heads/main".
	Describe string `json:"describe"`
}

// Submodules lists submodules recursively.
func (r *Repo) Submodules(ctx context.Context) ([]Submodule, error) {
	// Raw output: the first column is a status flag, and " " (up to date) is significant,
	// so the output must not be trimmed.
	out, err := r.Read(ctx, "submodule", "status", "--recursive")
	if err != nil {
		return nil, err
	}
	subs := []Submodule{}
	for line := range strings.SplitSeq(strings.TrimRight(string(out), "\r\n"), "\n") {
		line = strings.TrimRight(line, "\r")
		if s, ok := parseSubmoduleLine(line); ok {
			subs = append(subs, s)
		}
	}
	return subs, nil
}

// parseSubmoduleLine reads one line of `git submodule status`:
// "<flag><sha> <path>[ (<describe>)]" where flag is ' ', '-', '+' or 'U'.
func parseSubmoduleLine(line string) (Submodule, bool) {
	if len(line) < 42 {
		return Submodule{}, false
	}
	s := Submodule{State: "ok"}
	switch line[0] {
	case '-':
		s.State = "uninitialized"
	case '+':
		s.State = "modified"
	case 'U':
		s.State = "conflict"
	}
	rest := line[1:]
	sha, rest, ok := strings.Cut(rest, " ")
	if !ok {
		return Submodule{}, false
	}
	s.Commit = sha
	if i := strings.LastIndex(rest, " ("); i >= 0 && strings.HasSuffix(rest, ")") {
		s.Describe = rest[i+2 : len(rest)-1]
		rest = rest[:i]
	}
	s.Path = rest
	return s, true
}

// UpdateSubmodules initializes and checks out the recorded commit of every submodule
// (recursively). It may fetch, so it uses the network options for credentials.
func (r *Repo) UpdateSubmodules(ctx context.Context, o NetOptions) error {
	c := o.cmd("submodule", "update", "--init", "--recursive", "--progress")
	_, err := r.Write(ctx, c)
	return err
}
