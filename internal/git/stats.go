package git

import (
	"bufio"
	"context"
	"errors"
	"io"
	"slices"
	"strconv"
	"strings"
	"time"
)

// StatsOptions picks the history Stats summarises.
type StatsOptions struct {
	// Since and Until bound author dates, as Unix seconds.
	Since, Until int64
	// Rev is one branch or ref to follow, or "" for every branch, remote branch and tag.
	Rev string
}

// Stats summarises a stretch of history for the insights view.
type Stats struct {
	Commits int `json:"commits"`
	Merges  int `json:"merges"`
	// Bucket is "day" or "week": the width of each Activity entry.
	Bucket       string        `json:"bucket"`
	Activity     []Bucket      `json:"activity"`
	Contributors []Contributor `json:"contributors"`
	Hotspots     []Hotspot     `json:"hotspots"`
	// Cycle measures merged branches: from their first commit to the merge.
	Cycle CycleTime `json:"cycle"`
}

// Bucket counts the commits authored in [Start, Start+width).
type Bucket struct {
	Start   int64 `json:"start"`
	Commits int   `json:"commits"`
}

// Contributor is one author's share of the range (identities folded by .mailmap).
type Contributor struct {
	Name    string `json:"name"`
	Email   string `json:"email"`
	Commits int    `json:"commits"`
	Added   int    `json:"added"`
	Deleted int    `json:"deleted"`
}

// Hotspot is a file by how much it changed in the range.
type Hotspot struct {
	Path    string `json:"path"`
	Commits int    `json:"commits"`
	// Changed is lines added plus lines deleted; binary changes count as commits only.
	Changed int `json:"changed"`
}

// CycleTime summarises merge lead times, in hours, over the Sample most recent merges.
type CycleTime struct {
	Sample int     `json:"sample"`
	Median float64 `json:"median"`
	P90    float64 `json:"p90"`
}

// Limits on what Stats returns and how many merges it measures.
const (
	maxContributors = 50
	maxHotspots     = 25
	maxCycleMerges  = 50
	dayBucketsUpTo  = 92 * 24 * time.Hour
)

// Stats reads the history in opts with one streamed git log --numstat and summarises it.
func (r *Repo) Stats(ctx context.Context, opts StatsOptions) (Stats, error) {
	st := Stats{Bucket: "day", Activity: []Bucket{}, Contributors: []Contributor{}, Hotspots: []Hotspot{}}
	width := int64(24 * 60 * 60)
	if time.Duration(opts.Until-opts.Since)*time.Second > dayBucketsUpTo {
		st.Bucket, width = "week", 7*24*60*60
	}
	start := opts.Since - opts.Since%width
	for t := start; t <= opts.Until; t += width {
		st.Activity = append(st.Activity, Bucket{Start: t})
	}

	revs := []string{opts.Rev}
	if opts.Rev == "" {
		revs = []string{"--branches", "--remotes", "--tags"}
	}
	args := append([]string{"log", "--numstat", "--no-renames", "-z",
		"--since=@" + strconv.FormatInt(opts.Since, 10), "--until=@" + strconv.FormatInt(opts.Until, 10),
		"--format=%x1e%H%x1f%P%x1f%aN%x1f%aE%x1f%at"}, revs...)
	args = append(args, "--")

	people := map[string]*Contributor{}
	files := map[string]*Hotspot{}
	type merge struct {
		id, side string
		at       int64
	}
	var merges []merge
	err := r.Stream(ctx, func(rd io.Reader) error {
		br := bufio.NewReaderSize(rd, 64*1024)
		var who *Contributor
		for {
			field, err := br.ReadString(0)
			for _, part := range strings.Split(field, "\x1e") {
				part = strings.Trim(part, "\x00\n")
				if part == "" {
					continue
				}
				if f := strings.Split(part, "\x1f"); len(f) == 5 {
					// A commit header.
					at, _ := strconv.ParseInt(f[4], 10, 64)
					parents := strings.Fields(f[1])
					st.Commits++
					if i := (at - start) / width; at >= start && int(i) < len(st.Activity) {
						st.Activity[i].Commits++
					}
					key := strings.ToLower(f[3])
					if people[key] == nil {
						people[key] = &Contributor{Name: f[2], Email: f[3]}
					}
					who = people[key]
					who.Commits++
					if len(parents) > 1 {
						st.Merges++
						if len(merges) < maxCycleMerges {
							merges = append(merges, merge{id: f[0], side: parents[1], at: at})
						}
					}
					continue
				}
				// A numstat line: "added\tdeleted\tpath".
				cols := strings.SplitN(part, "\t", 3)
				if len(cols) != 3 || who == nil {
					continue
				}
				added, errA := strconv.Atoi(cols[0])
				deleted, errD := strconv.Atoi(cols[1])
				h := files[cols[2]]
				if h == nil {
					h = &Hotspot{Path: cols[2]}
					files[cols[2]] = h
				}
				h.Commits++
				if errA == nil && errD == nil { // "-" for binary files
					h.Changed += added + deleted
					who.Added += added
					who.Deleted += deleted
				}
			}
			if errors.Is(err, io.EOF) {
				return nil
			}
			if err != nil {
				return err
			}
		}
	}, args...)
	if err != nil && !isEmptyRepoLogError(err) {
		return st, err
	}

	for _, c := range people {
		st.Contributors = append(st.Contributors, *c)
	}
	slices.SortFunc(st.Contributors, func(a, b Contributor) int { return b.Commits - a.Commits })
	st.Contributors = st.Contributors[:min(len(st.Contributors), maxContributors)]
	for _, h := range files {
		st.Hotspots = append(st.Hotspots, *h)
	}
	slices.SortFunc(st.Hotspots, func(a, b Hotspot) int {
		if b.Changed != a.Changed {
			return b.Changed - a.Changed
		}
		return b.Commits - a.Commits
	})
	st.Hotspots = st.Hotspots[:min(len(st.Hotspots), maxHotspots)]

	// Cycle time: from the oldest commit a merge brought in to the merge itself.
	var hours []float64
	for _, m := range merges {
		out, err := r.ReadString(ctx, "log", "--format=%at", m.id+"^1.."+m.side, "--")
		if err != nil {
			continue
		}
		first := int64(0)
		for line := range strings.SplitSeq(strings.TrimSpace(out), "\n") {
			if t, err := strconv.ParseInt(line, 10, 64); err == nil && (first == 0 || t < first) {
				first = t
			}
		}
		if first > 0 && m.at >= first {
			hours = append(hours, float64(m.at-first)/3600)
		}
	}
	if len(hours) > 0 {
		slices.Sort(hours)
		st.Cycle = CycleTime{Sample: len(hours), Median: quantile(hours, 0.5), P90: quantile(hours, 0.9)}
	}
	return st, nil
}

// quantile reads the q-quantile of sorted values by nearest rank.
func quantile(sorted []float64, q float64) float64 {
	i := int(q*float64(len(sorted)-1) + 0.5)
	return sorted[min(i, len(sorted)-1)]
}

// BranchAge is a branch and when it last moved.
type BranchAge struct {
	Name string `json:"name"`
	// Kind is "local" or "remote".
	Kind   string `json:"kind"`
	Last   int64  `json:"last"`
	Author string `json:"author"`
}

// BranchAges lists local and remote branches with their last commit's date, oldest first.
func (r *Repo) BranchAges(ctx context.Context) ([]BranchAge, error) {
	out, err := r.ReadString(ctx, "for-each-ref", "--sort=committerdate", "--format=%(refname)%1f%(committerdate:unix)%1f%(authorname)", "refs/heads", "refs/remotes")
	if err != nil {
		return nil, err
	}
	list := []BranchAge{}
	for line := range strings.SplitSeq(strings.TrimSpace(out), "\n") {
		f := strings.Split(line, "\x1f")
		if len(f) != 3 || strings.HasSuffix(f[0], "/HEAD") {
			continue
		}
		b := BranchAge{Author: f[2]}
		b.Last, _ = strconv.ParseInt(f[1], 10, 64)
		if name, ok := strings.CutPrefix(f[0], "refs/heads/"); ok {
			b.Name, b.Kind = name, "local"
		} else {
			b.Name, b.Kind = strings.TrimPrefix(f[0], "refs/remotes/"), "remote"
		}
		list = append(list, b)
	}
	return list, nil
}
