package git

import (
	"bufio"
	"bytes"
	"context"
	"io"
	"regexp"
	"strconv"
	"strings"
)

// RefKind classifies a ref.
type RefKind string

// Ref kinds.
const (
	LocalBranch  RefKind = "local"
	RemoteBranch RefKind = "remote"
	Tag          RefKind = "tag"
	Stash        RefKind = "stash"
)

// Ref is a branch, remote-tracking branch, tag or stash entry.
type Ref struct {
	// Name is the full ref name, e.g. refs/heads/main, or stash@{0} for stashes.
	Name string `json:"name"`
	// Short is the display name, e.g. main, origin/main, v1.0.
	Short string  `json:"short"`
	Kind  RefKind `json:"kind"`
	// Commit is the commit the ref points at (annotated tags are peeled).
	Commit string `json:"commit"`
	// Head marks the checked-out branch.
	Head bool `json:"head"`
	// Upstream is the tracked branch's short name, for local branches.
	Upstream string `json:"upstream"`
	Ahead    int    `json:"ahead"`
	Behind   int    `json:"behind"`
	// Gone means the upstream is configured but no longer exists.
	Gone bool `json:"gone"`
	// Subject is the stash message, for stashes.
	Subject string `json:"subject,omitempty"`
}

const refFormat = "%(refname)%00%(refname:short)%00%(objectname)%00%(*objectname)%00%(HEAD)%00%(upstream:short)%00%(upstream:track)"

var trackRE = regexp.MustCompile(`(ahead|behind) (\d+)`)

// Refs lists branches, remote-tracking branches, tags and stashes. Symbolic refs such as
// refs/remotes/origin/HEAD are skipped.
func (r *Repo) Refs(ctx context.Context) ([]Ref, error) {
	var refs []Ref
	err := r.Stream(ctx, func(rd io.Reader) error {
		sc := bufio.NewScanner(rd)
		sc.Buffer(make([]byte, 64*1024), 1024*1024)
		for sc.Scan() {
			f := bytes.Split(sc.Bytes(), []byte{0})
			if len(f) != 7 {
				continue
			}
			ref := Ref{
				Name:     string(f[0]),
				Short:    string(f[1]),
				Commit:   string(f[2]),
				Head:     string(f[4]) == "*",
				Upstream: string(f[5]),
			}
			if len(f[3]) > 0 {
				ref.Commit = string(f[3]) // annotated tag: use the commit it points to
			}
			switch {
			case strings.HasPrefix(ref.Name, "refs/heads/"):
				ref.Kind = LocalBranch
			case strings.HasPrefix(ref.Name, "refs/remotes/"):
				if strings.HasSuffix(ref.Name, "/HEAD") {
					continue
				}
				ref.Kind = RemoteBranch
			case strings.HasPrefix(ref.Name, "refs/tags/"):
				ref.Kind = Tag
			default:
				continue
			}
			track := string(f[6])
			ref.Gone = track == "[gone]"
			for _, m := range trackRE.FindAllStringSubmatch(track, -1) {
				n, _ := strconv.Atoi(m[2])
				if m[1] == "ahead" {
					ref.Ahead = n
				} else {
					ref.Behind = n
				}
			}
			refs = append(refs, ref)
		}
		return sc.Err()
	}, "for-each-ref", "--format="+refFormat, "refs/heads", "refs/remotes", "refs/tags")
	if err != nil {
		return nil, err
	}

	stashes, err := r.stashes(ctx)
	if err != nil {
		return nil, err
	}
	return append(refs, stashes...), nil
}

func (r *Repo) stashes(ctx context.Context) ([]Ref, error) {
	if _, err := r.Read(ctx, "rev-parse", "--verify", "-q", "refs/stash"); err != nil {
		return nil, nil // no stashes
	}
	out, err := r.Read(ctx, "log", "-g", "-z", "--format=%H%x1f%gd%x1f%s", "refs/stash", "--")
	if err != nil {
		return nil, err
	}
	var refs []Ref
	for rec := range strings.SplitSeq(string(out), "\x00") {
		f := strings.SplitN(strings.TrimLeft(rec, "\n"), "\x1f", 3)
		if len(f) != 3 {
			continue
		}
		refs = append(refs, Ref{Name: f[1], Short: f[1], Kind: Stash, Commit: f[0], Subject: f[2]})
	}
	return refs, nil
}
