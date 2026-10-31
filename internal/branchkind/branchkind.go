// Package branchkind classifies branches into the built-in kinds (Trunk, Release,
// Environment, Docs, Bot) and decides which are hidden, by glob patterns kept per
// repository in local settings (ADR 0009).
package branchkind

import (
	"fmt"
	"slices"
	"strings"
	"unicode"
)

// Kind is a branch's kind. The zero value is an ordinary branch.
type Kind string

// The built-in kinds, in the order they are tried: the first whose patterns match wins.
const (
	None        Kind = ""
	Trunk       Kind = "trunk"
	Release     Kind = "release"
	Environment Kind = "environment"
	Docs        Kind = "docs"
	Bot         Kind = "bot"
)

// Order lists the kinds in matching order.
var Order = []Kind{Trunk, Release, Environment, Docs, Bot}

// Defaults are the patterns each kind uses until a repository overrides them. Environment
// has none: staging and production are too often ordinary branches.
var Defaults = map[Kind][]string{
	Trunk:       {"main", "master", "develop", "trunk"},
	Release:     {"release/*", "releases/*", "hotfix/*"},
	Environment: {},
	Docs:        {"gh-pages", "gh-docs", "docs-site"},
	Bot:         {"dependabot/*", "renovate/*"},
}

// Valid reports whether k is one of the built-in kinds.
func Valid(k Kind) bool { return slices.Contains(Order, k) }

// Rules are one repository's kind patterns and hide patterns.
type Rules struct {
	kinds map[Kind][]string
	hide  []string
}

// New returns rules from a repository's overrides: a kind missing from overrides uses its
// defaults, and a kind present with an empty list matches nothing.
func New(overrides map[Kind][]string, hide []string) Rules {
	r := Rules{kinds: map[Kind][]string{}, hide: hide}
	for _, k := range Order {
		if p, ok := overrides[k]; ok {
			r.kinds[k] = p
		} else {
			r.kinds[k] = Defaults[k]
		}
	}
	return r
}

// Patterns returns the patterns kind k uses under these rules.
func (r Rules) Patterns(k Kind) []string { return r.kinds[k] }

// Hide returns the hide patterns.
func (r Rules) Hide() []string { return r.hide }

// Kind classifies a branch by its name without any remote prefix (see Strip).
func (r Rules) Kind(name string) Kind {
	for _, k := range Order {
		for _, p := range r.kinds[k] {
			if Match(p, name) {
				return k
			}
		}
	}
	return None
}

// Hidden reports whether a branch, named without any remote prefix, matches a hide
// pattern.
func (r Rules) Hidden(name string) bool {
	for _, p := range r.hide {
		if Match(p, name) {
			return true
		}
	}
	return false
}

// Strip removes a remote's name from a remote-tracking branch, so origin/release/1.4
// classifies as release/1.4. Names that don't start with one of remotes are returned as
// they are. The longest matching remote wins, for remotes whose names contain a slash.
func Strip(name string, remotes []string) string {
	best := ""
	for _, rm := range remotes {
		if len(rm) > len(best) && strings.HasPrefix(name, rm+"/") {
			best = rm
		}
	}
	if best == "" {
		return name
	}
	return name[len(best)+1:]
}

// Match reports whether name matches pattern. `*` matches any run of characters, slashes
// included, so dependabot/* matches dependabot/npm_and_yarn/vite-7.1; `?` matches one
// character other than a slash. Everything else matches itself.
func Match(pattern, name string) bool {
	p, n := 0, 0
	star, mark := -1, 0
	for n < len(name) {
		switch {
		case p < len(pattern) && pattern[p] == '*':
			star, mark = p, n
			p++
		case p < len(pattern) && (pattern[p] == name[n] || pattern[p] == '?' && name[n] != '/'):
			p++
			n++
		case star >= 0:
			// Let the last star take one more character and try again.
			mark++
			p, n = star+1, mark
		default:
			return false
		}
	}
	for p < len(pattern) && pattern[p] == '*' {
		p++
	}
	return p == len(pattern)
}

// CheckPattern returns an error describing why pattern can't match a branch name, or nil.
func CheckPattern(pattern string) error {
	switch {
	case pattern == "":
		return fmt.Errorf("a pattern is empty")
	case strings.Contains(pattern, ".."), strings.Contains(pattern, "//"), strings.HasPrefix(pattern, "/"), strings.HasSuffix(pattern, "/"):
		return fmt.Errorf("%q can't match a branch name", pattern)
	}
	for _, c := range pattern {
		if unicode.IsSpace(c) || unicode.IsControl(c) || strings.ContainsRune("~^:\\[", c) {
			return fmt.Errorf("%q has %q, which branch names can't contain", pattern, c)
		}
	}
	return nil
}
