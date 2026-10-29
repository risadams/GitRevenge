package git

import (
	"context"
	"strings"
)

// Clone clones url into dir, which must not exist or be empty. git creates dir and removes
// it again if the clone fails, but not if it is killed by cancellation.
func Clone(ctx context.Context, runner *Runner, url, dir string, o NetOptions) error {
	_, err := runner.Run(ctx, o.cmd("clone", "--progress", "--", url, dir))
	return err
}

// Init creates an empty repository in dir, creating dir if needed. branch names the initial
// branch; empty uses the user's init.defaultBranch.
func Init(ctx context.Context, runner *Runner, dir, branch string) error {
	args := []string{"init", "-q"}
	if branch != "" {
		args = append(args, "--initial-branch="+branch)
	}
	_, err := runner.Run(ctx, Cmd{Args: append(args, "--", dir)})
	return err
}

// CloneName returns the folder name git clone would choose for url, e.g. "repo" for
// "git@host:owner/repo.git" or "https://host/owner/repo/". It returns "" if url has none.
func CloneName(url string) string {
	s := strings.TrimSpace(url)
	s = strings.TrimRight(s, `/\`)
	s = strings.TrimSuffix(s, "/.git")
	s = strings.TrimRight(s, `/\`)
	if i := strings.LastIndexAny(s, `/\:`); i >= 0 {
		s = s[i+1:]
	}
	s = strings.TrimSuffix(s, ".git")
	if s == "." || s == ".." {
		return ""
	}
	return s
}
