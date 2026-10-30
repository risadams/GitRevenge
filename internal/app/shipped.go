package app

import (
	"context"
	"errors"
	"time"

	"github.com/risadams/gitrevenge/internal/git"
)

// InsightsInputs are the choices an insights result depends on, besides the repository
// itself (ADR-0008). The same inputs on the same repository state give the same result.
type InsightsInputs struct {
	// Target is the branch to measure against, or "" to detect it (origin/HEAD, main, master).
	Target string `json:"target"`
	// Window is "7d", "14d", "month", "custom" or "releases".
	Window string `json:"window"`
	// Since and Until bound a custom window, Unix seconds, inclusive.
	Since int64 `json:"since"`
	Until int64 `json:"until"`
	// From and To are the tags of a releases window.
	From string `json:"from"`
	To   string `json:"to"`
	// AsOf is the point windows and ages are measured from, Unix seconds, or 0 for the
	// target's tip commit.
	AsOf int64 `json:"asOf"`
}

// InsightsWindow is what inputs resolved to, printed with every result so it can be
// reproduced.
type InsightsWindow struct {
	Target    string `json:"target"`
	TargetSHA string `json:"targetSha"`
	AsOf      int64  `json:"asOf"`
	// Since and Until are the window's absolute bounds, Unix seconds (UTC), inclusive. For
	// a releases window they are the two tags' commit dates.
	Since int64 `json:"since"`
	Until int64 `json:"until"`
	// Range is "from..to" for a releases window, else "".
	Range string `json:"range"`
}

// ShippedResult is the Shipped view's data for one repository.
type ShippedResult struct {
	Window InsightsWindow    `json:"window"`
	Items  []git.ShippedItem `json:"items"`
}

const day = int64(24 * 60 * 60)

// resolveWindow turns inputs into absolute bounds measured from asOf. Day windows cover
// whole UTC days ending with asOf's day; a month runs from the 1st of asOf's UTC month.
func resolveWindow(in InsightsInputs, asOf int64) (since, until int64, err error) {
	startOfDay := asOf - ((asOf%day)+day)%day
	switch in.Window {
	case "7d", "":
		return startOfDay - 6*day, asOf, nil
	case "14d":
		return startOfDay - 13*day, asOf, nil
	case "month":
		t := time.Unix(asOf, 0).UTC()
		return time.Date(t.Year(), t.Month(), 1, 0, 0, 0, 0, time.UTC).Unix(), asOf, nil
	case "custom":
		if in.Until < in.Since {
			return 0, 0, userError("The start of the window must be before its end.")
		}
		return in.Since, in.Until, nil
	}
	return 0, 0, userError("Unknown window " + in.Window + ".")
}

// resolveInputs fixes the target, as-of point and window for one repository.
func resolveInputs(ctx context.Context, g *git.Repo, in InsightsInputs) (InsightsWindow, error) {
	target := in.Target
	if target == "" {
		t, err := g.InsightsTarget(ctx)
		if errors.Is(err, git.ErrNoTarget) {
			return InsightsWindow{}, userError("This repository has no origin/HEAD, main or master branch to measure against. Choose a branch.")
		}
		if err != nil {
			return InsightsWindow{}, err
		}
		target = t
	}
	tip, err := g.Tip(ctx, target)
	if err != nil {
		return InsightsWindow{}, userError("There's no branch named " + target + ".")
	}
	w := InsightsWindow{Target: target, TargetSHA: tip.SHA, AsOf: in.AsOf}
	if w.AsOf == 0 {
		w.AsOf = tip.Time
	}
	if in.Window == "releases" {
		if in.From == "" || in.To == "" {
			return InsightsWindow{}, userError("Choose the two releases to compare.")
		}
		from, errF := g.Tip(ctx, in.From)
		to, errT := g.Tip(ctx, in.To)
		if errF != nil || errT != nil {
			return InsightsWindow{}, userError("One of those tags no longer exists.")
		}
		w.Range, w.Since, w.Until = in.From+".."+in.To, from.Time, to.Time
		if in.AsOf == 0 {
			w.AsOf = to.Time
		}
		return w, nil
	}
	w.Since, w.Until, err = resolveWindow(in, w.AsOf)
	return w, err
}

// InsightsShipped lists what landed on the target branch in the window: its first-parent
// commits, with authors, sizes, ticket references and, for merges, cycle times.
func (a *App) InsightsShipped(path string, in InsightsInputs) (ShippedResult, error) {
	ctx := context.Background()
	g, _, err := a.repoGit(ctx, path)
	if err != nil {
		return ShippedResult{}, err
	}
	w, err := resolveInputs(ctx, g, in)
	if err != nil {
		return ShippedResult{}, err
	}
	items, err := g.Shipped(ctx, git.ShippedQuery{Target: w.Target, Since: w.Since, Until: w.Until, Range: w.Range})
	if err != nil {
		return ShippedResult{}, err
	}
	return ShippedResult{Window: w, Items: items}, nil
}

// InsightsTags lists a repository's tags newest first, for choosing a releases window.
func (a *App) InsightsTags(path string) ([]git.TagInfo, error) {
	ctx := context.Background()
	g, _, err := a.repoGit(ctx, path)
	if err != nil {
		return nil, err
	}
	return g.Tags(ctx)
}
