package session

import (
	"context"
	"strings"

	"github.com/risadams/gitrevenge/internal/branchkind"
	"github.com/risadams/gitrevenge/internal/git"
)

// maxPins is how many Trunk lanes the graph holds at the left (ADR 0009).
const maxPins = 2

// TrunkTips returns the commits of the Trunk branches to pin in the graph, in pattern
// order (see KindTips).
func TrunkTips(ctx context.Context, g *git.Repo, rules branchkind.Rules, remotes []string, show bool) []string {
	pins, _ := KindTips(ctx, g, rules, remotes, show)
	return pins
}

// KindTips reads the branches the kind patterns name, and returns the Trunk tips to pin
// (for each Trunk pattern in order, the local branch if there is one, else a remote's; at
// most maxPins) and the kind of every kinded branch tip, for the graph's lanes. Hidden
// branches are left out unless they're being shown. It reads only the refs the patterns
// name, so it stays fast with thousands of branches.
func KindTips(ctx context.Context, g *git.Repo, rules branchkind.Rules, remotes []string, show bool) (pins []string, kinds map[string]string) {
	var patterns []string
	for _, k := range branchkind.Order {
		patterns = append(patterns, rules.Patterns(k)...)
	}
	if len(patterns) == 0 {
		return nil, nil
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
		return nil, nil
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
	visible := func(name string) bool { return show || !rules.Hidden(name) }

	kinds = map[string]string{}
	for _, list := range [][]ref{remote, local} { // local last, so it wins a shared tip
		for _, r := range list {
			if k := rules.Kind(r.name); k != branchkind.None && visible(r.name) {
				kinds[r.commit] = string(k)
			}
		}
	}

	seen := map[string]bool{}
	for _, p := range rules.Patterns(branchkind.Trunk) {
		pick := ""
		for _, list := range [][]ref{local, remote} {
			for _, r := range list {
				if branchkind.Match(p, r.name) && visible(r.name) {
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
			pins = append(pins, pick)
			if len(pins) == maxPins {
				break
			}
		}
	}
	return pins, kinds
}
