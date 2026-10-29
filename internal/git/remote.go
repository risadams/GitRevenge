package git

import (
	"context"
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// Remote is a configured remote repository.
type Remote struct {
	Name     string `json:"name"`
	FetchURL string `json:"fetchUrl"`
	PushURL  string `json:"pushUrl"`
}

// Remotes lists configured remotes.
func (r *Repo) Remotes(ctx context.Context) ([]Remote, error) {
	out, err := r.ReadString(ctx, "remote", "-v")
	if err != nil {
		return nil, err
	}
	byName := map[string]*Remote{}
	var order []string
	for line := range strings.SplitSeq(out, "\n") {
		// "origin\thttps://example.com/x.git (fetch)"
		name, rest, ok := strings.Cut(line, "\t")
		if !ok {
			continue
		}
		url, kind, _ := strings.Cut(rest, " ")
		rm, seen := byName[name]
		if !seen {
			rm = &Remote{Name: name}
			byName[name] = rm
			order = append(order, name)
		}
		if kind == "(push)" {
			rm.PushURL = url
		} else {
			rm.FetchURL = url
		}
	}
	remotes := make([]Remote, 0, len(order))
	for _, n := range order {
		remotes = append(remotes, *byName[n])
	}
	return remotes, nil
}

// checkRemoteName refuses names git could misread as options in later commands (such as
// "git fetch -x") or that can't form remote-tracking ref names.
func (r *Repo) checkRemoteName(ctx context.Context, name string) error {
	if name == "" || strings.HasPrefix(name, "-") {
		return fmt.Errorf("%q: %w", name, ErrInvalidName)
	}
	if _, err := r.Read(ctx, "check-ref-format", "refs/remotes/"+name+"/x"); err != nil {
		return fmt.Errorf("%q: %w", name, ErrInvalidName)
	}
	return nil
}

// AddRemote adds a remote.
func (r *Repo) AddRemote(ctx context.Context, name, url string) error {
	if err := r.checkRemoteName(ctx, name); err != nil {
		return err
	}
	_, err := r.Write(ctx, Cmd{Args: []string{"remote", "add", "--", name, url}})
	return err
}

// RemoveRemote deletes a remote and its remote-tracking branches.
func (r *Repo) RemoveRemote(ctx context.Context, name string) error {
	_, err := r.Write(ctx, Cmd{Args: []string{"remote", "remove", name}})
	return err
}

// RenameRemote renames a remote, updating its remote-tracking branches and upstreams.
func (r *Repo) RenameRemote(ctx context.Context, from, to string) error {
	if err := r.checkRemoteName(ctx, to); err != nil {
		return err
	}
	_, err := r.Write(ctx, Cmd{Args: []string{"remote", "rename", from, to}})
	return err
}

// SetRemoteURL changes a remote's URL (fetch and push).
func (r *Repo) SetRemoteURL(ctx context.Context, name, url string) error {
	_, err := r.Write(ctx, Cmd{Args: []string{"remote", "set-url", "--", name, url}})
	return err
}

// Progress is one update from a network operation.
type Progress struct {
	// Phase is git's label, e.g. "Receiving objects"; Line is the raw text.
	Phase   string `json:"phase"`
	Percent int    `json:"percent"`
	Line    string `json:"line"`
}

var progressRE = regexp.MustCompile(`^(?:remote: )?([A-Za-z ]+):\s+(\d+)%`)

// ParseProgress extracts a phase and percentage from a git progress line, if it has one.
func ParseProgress(line string) Progress {
	p := Progress{Line: line, Percent: -1}
	if m := progressRE.FindStringSubmatch(line); m != nil {
		p.Phase = strings.TrimSpace(m[1])
		p.Percent, _ = strconv.Atoi(m[2])
	}
	return p
}

// NetOptions configures a network command.
type NetOptions struct {
	// Env carries the askpass bridge settings for credential prompts.
	Env []string
	// OnProgress receives progress updates.
	OnProgress func(Progress)
}

func (o NetOptions) cmd(args ...string) Cmd {
	c := Cmd{Args: args, Env: o.Env}
	if o.OnProgress != nil {
		c.Progress = func(line string) { o.OnProgress(ParseProgress(line)) }
	}
	return c
}

// Fetch downloads from remote (or every remote when remote is empty), pruning
// remote-tracking branches that no longer exist. It only updates refs/remotes and tags,
// so it runs alongside other operations rather than waiting in the write queue.
func (r *Repo) Fetch(ctx context.Context, remote string, o NetOptions) error {
	args := []string{"fetch", "--progress", "--prune"}
	if remote == "" {
		args = append(args, "--all")
	} else {
		args = append(args, remote)
	}
	c := o.cmd(args...)
	c.Dir = r.dir
	_, err := r.runner.Run(ctx, c)
	return err
}

// PullMode chooses how pull integrates upstream changes.
type PullMode string

// Pull modes. PullDefault follows the user's pull.rebase / pull.ff config.
const (
	PullDefault PullMode = ""
	PullFFOnly  PullMode = "ff-only"
	PullMerge   PullMode = "merge"
	PullRebase  PullMode = "rebase"
)

// Pull fetches and integrates the current branch's upstream.
func (r *Repo) Pull(ctx context.Context, mode PullMode, o NetOptions) error {
	args := []string{"pull", "--progress"}
	switch mode {
	case PullFFOnly:
		args = append(args, "--ff-only")
	case PullMerge:
		args = append(args, "--no-rebase")
	case PullRebase:
		args = append(args, "--rebase")
	}
	c := o.cmd(args...)
	c.Env = append(c.Env, "GIT_EDITOR=true") // merge commits keep git's message
	_, err := r.Write(ctx, c)
	return err
}

// PushOptions selects what Push sends.
type PushOptions struct {
	Remote string
	// Refspec is what to push, e.g. "main", "refs/tags/v1.0", or ":branch" to delete.
	Refspec string
	// SetUpstream records the pushed branch as the local branch's upstream.
	SetUpstream bool
	// ForceWithLease overwrites the remote branch only if it is where we last saw it.
	ForceWithLease bool
}

// Push sends refs to a remote.
func (r *Repo) Push(ctx context.Context, p PushOptions, o NetOptions) error {
	args := []string{"push", "--progress"}
	if p.SetUpstream {
		args = append(args, "--set-upstream")
	}
	if p.ForceWithLease {
		args = append(args, "--force-with-lease")
	}
	args = append(args, "--", p.Remote, p.Refspec)
	c := o.cmd(args...)
	c.Dir = r.dir
	_, err := r.runner.Run(ctx, c)
	return err
}
