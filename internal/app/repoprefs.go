package app

import (
	"slices"
	"strings"

	"github.com/risadams/gitrevenge/internal/branchkind"
	"github.com/risadams/gitrevenge/internal/settings"
)

// KindPatterns are one branch kind's patterns in a repository.
type KindPatterns struct {
	Kind string `json:"kind"`
	// Patterns are the patterns in use: the repository's own, or the defaults.
	Patterns []string `json:"patterns"`
	// Custom is true when the repository overrides the defaults.
	Custom bool `json:"custom"`
	// Defaults are the built-in patterns, for showing and restoring.
	Defaults []string `json:"defaults"`
}

// RepoPreferences are the settings that belong to one repository.
type RepoPreferences struct {
	// Kinds lists every branch kind, in matching order.
	Kinds []KindPatterns `json:"kinds"`
	// Hide lists patterns of branches to leave out of the sidebar and the graph.
	Hide          []string `json:"hide"`
	CleanupOnPull bool     `json:"cleanupOnPull"`
	SyncNotes     bool     `json:"syncNotes"`
}

// RepoPreferences returns the settings for the repository at path.
func (a *App) RepoPreferences(path string) (RepoPreferences, error) {
	m, err := a.sessions()
	if err != nil {
		return RepoPreferences{}, err
	}
	rs := m.Settings().Get().Repo(path)
	out := RepoPreferences{Hide: nonNil(rs.Hide), CleanupOnPull: rs.CleanupOnPull, SyncNotes: rs.SyncNotes}
	for _, k := range branchkind.Order {
		kp := KindPatterns{Kind: string(k), Defaults: nonNil(branchkind.Defaults[k])}
		if p, ok := rs.Kinds[string(k)]; ok {
			kp.Patterns, kp.Custom = nonNil(p), true
		} else {
			kp.Patterns = kp.Defaults
		}
		out.Kinds = append(out.Kinds, kp)
	}
	return out, nil
}

// SetRepoPreferences checks and saves the settings for the repository at path. A kind
// whose patterns equal its defaults is stored as using the defaults, so later changes to
// the defaults reach it.
func (a *App) SetRepoPreferences(path string, p RepoPreferences) error {
	if strings.TrimSpace(path) == "" {
		return userError("No repository is open.")
	}
	rs := settings.RepoSettings{Path: path, CleanupOnPull: p.CleanupOnPull, SyncNotes: p.SyncNotes}
	for _, kp := range p.Kinds {
		k := branchkind.Kind(kp.Kind)
		if !branchkind.Valid(k) {
			return userError("Unknown branch kind " + kp.Kind + ".")
		}
		pats, err := cleanPatterns(kp.Patterns)
		if err != nil {
			return err
		}
		if !kp.Custom || slices.Equal(pats, branchkind.Defaults[k]) {
			continue
		}
		if rs.Kinds == nil {
			rs.Kinds = map[string][]string{}
		}
		rs.Kinds[string(k)] = pats
	}
	hide, err := cleanPatterns(p.Hide)
	if err != nil {
		return err
	}
	if len(hide) > 0 {
		rs.Hide = hide
	}
	m, err := a.sessions()
	if err != nil {
		return err
	}
	return m.Settings().Update(func(s *settings.Settings) { s.SetRepo(rs) })
}

// branchRules returns the kind and hide rules for the repository at path.
func (a *App) branchRules(path string) branchkind.Rules {
	m, err := a.sessions()
	if err != nil {
		return branchkind.New(nil, nil)
	}
	rs := m.Settings().Get().Repo(path)
	overrides := map[branchkind.Kind][]string{}
	for k, p := range rs.Kinds {
		overrides[branchkind.Kind(k)] = p
	}
	return branchkind.New(overrides, rs.Hide)
}

// cleanPatterns trims patterns, drops blanks and duplicates, and checks each one.
func cleanPatterns(in []string) ([]string, error) {
	out := []string{}
	for _, p := range in {
		p = strings.TrimSpace(p)
		if p == "" || slices.Contains(out, p) {
			continue
		}
		if err := branchkind.CheckPattern(p); err != nil {
			return nil, userError("Branch pattern: " + err.Error() + ".")
		}
		out = append(out, p)
	}
	return out, nil
}

func nonNil(s []string) []string {
	if s == nil {
		return []string{}
	}
	return slices.Clone(s)
}
