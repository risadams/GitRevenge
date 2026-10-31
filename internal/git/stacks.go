package git

import (
	"context"
	"slices"
	"strings"
)

// MaxStackCandidates caps how many branches are checked as possible stack parents: each
// costs one `for-each-ref --contains`.
const MaxStackCandidates = 100

// StackParents infers which local branch each local branch is stacked on. B is stacked on
// A when A isn't merged into trunk, A's tip is in B's history (and isn't B's tip too), and
// no other such branch sits between them: A is the nearest. skip leaves out branches that
// can't be parents, such as Trunk branches. trunk "" treats every local branch as unmerged.
func (r *Repo) StackParents(ctx context.Context, trunk string, skip func(name string) bool) (map[string]string, error) {
	args := []string{"for-each-ref", "--format=%(refname:short) %(objectname)"}
	if trunk != "" {
		args = append(args, "--no-merged="+trunk)
	}
	out, err := r.ReadString(ctx, append(args, "refs/heads")...)
	if err != nil {
		return nil, err
	}
	tips := map[string]string{}
	var cands []string
	for line := range strings.Lines(out) {
		name, sha, ok := strings.Cut(strings.TrimSpace(line), " ")
		if !ok {
			continue
		}
		tips[name] = sha
		if skip == nil || !skip(name) {
			cands = append(cands, name)
		}
	}
	slices.Sort(cands)
	cands = cands[:min(len(cands), MaxStackCandidates)]

	// containers[A] is every local branch whose history holds A's tip, other than branches
	// sitting exactly on it.
	containers := map[string]map[string]bool{}
	for _, a := range cands {
		out, err := r.ReadString(ctx, "for-each-ref", "--contains="+tips[a], "--format=%(refname:short) %(objectname)", "refs/heads")
		if err != nil {
			return nil, err
		}
		set := map[string]bool{}
		for line := range strings.Lines(out) {
			name, sha, ok := strings.Cut(strings.TrimSpace(line), " ")
			if ok && name != a && sha != tips[a] {
				set[name] = true
			}
		}
		containers[a] = set
	}

	parents := map[string]string{}
	for b := range tips {
		var under []string
		for _, a := range cands {
			if containers[a][b] {
				under = append(under, a)
			}
		}
		// The nearest: a candidate that no other candidate below b contains.
		for _, a := range under {
			nearest := true
			for _, c := range under {
				if c != a && containers[a][c] {
					nearest = false
					break
				}
			}
			if nearest {
				parents[b] = a
				break
			}
		}
	}
	return parents, nil
}
