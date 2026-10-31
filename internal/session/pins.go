package session

import (
	"context"
	"strings"

	"github.com/risadams/gitrevenge/internal/branchkind"
	"github.com/risadams/gitrevenge/internal/git"
)

// maxPins is how many Trunk lanes the graph holds at the left (ADR 0009).
const maxPins = 2

// TrunkTips returns the commits of the Trunk branches to pin in the graph, in pattern order:
// for each Trunk pattern, the local branch if there is one, else a remote's. Hidden
// branches aren't pinned unless they're being shown. It reads only the refs the patterns
// name, so it stays fast with thousands of branches.
func TrunkTips(ctx context.Context, g *git.Repo, rules branchkind.Rules, remotes []string, show bool) []string {
	patterns := rules.Patterns(branchkind.Trunk)
	if len(patterns) == 0 {
		return nil
	}
	args := []string{"for-each-ref", "--format=%(refname) %(objectname)"}
	for _, p := range patterns {
		args = append(args, "refs/heads/"+p)
		for _, r := range remotes {
			args = append(args, "refs/remotes/"+r+"/"+p)
		}
	}
	out, err := g.ReadString(ctx, args...)
	if err != nil {
		return nil
	}
	type ref struct{ name, commit string }
	var local, remote []ref
	for line := range strings.Lines(out) {
		name, commit, ok := strings.Cut(strings.TrimSpace(line), " ")
		if !ok {
			continue
		}
		switch {
		case strings.HasPrefix(name, "refs/heads/"):
			local = append(local, ref{strings.TrimPrefix(name, "refs/heads/"), commit})
		case strings.HasPrefix(name, "refs/remotes/"):
			remote = append(remote, ref{branchkind.Strip(strings.TrimPrefix(name, "refs/remotes/"), remotes), commit})
		}
	}
	var tips []string
	seen := map[string]bool{}
	for _, p := range patterns {
		pick := ""
		for _, list := range [][]ref{local, remote} {
			for _, r := range list {
				if branchkind.Match(p, r.name) && (show || !rules.Hidden(r.name)) {
					pick = r.commit
					break
				}
			}
			if pick != "" {
				break
			}
		}
		if pick != "" && !seen[pick] {
			seen[pick] = true
			tips = append(tips, pick)
			if len(tips) == maxPins {
				break
			}
		}
	}
	return tips
}
