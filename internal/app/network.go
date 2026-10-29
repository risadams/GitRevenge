package app

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/risadams/gitrevenge/internal/askpass"
	"github.com/risadams/gitrevenge/internal/git"
	"github.com/risadams/gitrevenge/internal/session"
)

// network holds the credential bridge and in-flight network operations.
type network struct {
	once    sync.Once
	server  *askpass.Server
	exe     string
	initErr error

	mu      sync.Mutex
	prompts map[string]chan promptAnswer  // pending prompts by ID
	running map[string]context.CancelFunc // network operation per repo path
}

type promptAnswer struct {
	answer string
	ok     bool
}

// Prompt is the payload of the "askpass:prompt" event.
type Prompt struct {
	ID     string `json:"id"`
	Text   string `json:"text"`
	Kind   string `json:"kind"`
	Expiry int64  `json:"expiry"`
}

// NetProgress is the payload of the "net:progress" event.
type NetProgress struct {
	Path    string `json:"path"`
	Op      string `json:"op"`
	Phase   string `json:"phase"`
	Percent int    `json:"percent"`
	Line    string `json:"line"`
}

// askEnv returns the environment that routes credential prompts to this app.
func (a *App) askEnv() ([]string, error) {
	n := &a.net
	n.once.Do(func() {
		n.prompts = map[string]chan promptAnswer{}
		n.running = map[string]context.CancelFunc{}
		// Tests may preset exe to a small helper build.
		if n.exe == "" {
			if n.exe, n.initErr = os.Executable(); n.initErr != nil {
				return
			}
		}
		n.server, n.initErr = askpass.Listen(a.handlePrompt)
	})
	if n.initErr != nil {
		return nil, n.initErr
	}
	return n.server.Env(n.exe), nil
}

// handlePrompt shows a prompt in the UI and waits for AnswerPrompt.
func (a *App) handlePrompt(ctx context.Context, text string, kind askpass.Kind) (string, bool) {
	if a.promptHook != nil {
		return a.promptHook(text, kind)
	}
	if a.ctx == nil {
		return "", false // no UI to ask
	}
	idb := make([]byte, 8)
	_, _ = rand.Read(idb)
	id := hex.EncodeToString(idb)
	ch := make(chan promptAnswer, 1)
	a.net.mu.Lock()
	a.net.prompts[id] = ch
	a.net.mu.Unlock()
	defer func() {
		a.net.mu.Lock()
		delete(a.net.prompts, id)
		a.net.mu.Unlock()
	}()
	deadline, _ := ctx.Deadline()
	a.emit("askpass:prompt", Prompt{ID: id, Text: strings.TrimSpace(text), Kind: string(kind), Expiry: deadline.Unix()})
	select {
	case ans := <-ch:
		return ans.answer, ans.ok
	case <-ctx.Done():
		a.emit("askpass:expired", id)
		return "", false
	}
}

// AnswerPrompt delivers the user's answer to a credential prompt. ok=false cancels it.
func (a *App) AnswerPrompt(id, answer string, ok bool) {
	a.net.mu.Lock()
	ch := a.net.prompts[id]
	a.net.mu.Unlock()
	if ch != nil {
		ch <- promptAnswer{answer, ok}
	}
}

// runNetwork runs fn as the repository's one network operation, with progress events and
// cancellation via CancelNetwork.
func (a *App) runNetwork(path, op string, fn func(ctx context.Context, r *session.Repo, o git.NetOptions) error) error {
	r, err := a.openRepo(path)
	if err != nil {
		return err
	}
	return a.track(r.Path, r.Path, op, "Another fetch, pull or push is already running in this repository.",
		func(ctx context.Context, o git.NetOptions) error { return fn(ctx, r, o) })
}

// track runs fn as the network operation registered under key (busy is the message when one
// is already running there), reporting progress as "net:progress" events for path.
func (a *App) track(key, path, op string, busy userError, fn func(ctx context.Context, o git.NetOptions) error) error {
	env, err := a.askEnv()
	if err != nil {
		return err
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	a.net.mu.Lock()
	if _, running := a.net.running[key]; running {
		a.net.mu.Unlock()
		return busy
	}
	a.net.running[key] = cancel
	a.net.mu.Unlock()
	defer func() {
		a.net.mu.Lock()
		delete(a.net.running, key)
		a.net.mu.Unlock()
	}()

	var last time.Time
	o := git.NetOptions{Env: env, OnProgress: func(p git.Progress) {
		// git redraws meters many times a second; the UI needs a few updates.
		if time.Since(last) < 100*time.Millisecond && p.Percent != 100 {
			return
		}
		last = time.Now()
		a.emit("net:progress", NetProgress{Path: path, Op: op, Phase: p.Phase, Percent: p.Percent, Line: p.Line})
	}}
	err = fn(ctx, o)
	if errors.Is(err, context.Canceled) {
		return userError("Cancelled.")
	}
	return netErr(err)
}

// CancelNetwork stops the repository's running fetch, pull or push.
func (a *App) CancelNetwork(path string) {
	r, err := a.openRepo(path)
	if err != nil {
		return
	}
	a.cancelNetwork(r.Path)
}

func (a *App) cancelNetwork(key string) {
	a.net.mu.Lock()
	cancel := a.net.running[key]
	a.net.mu.Unlock()
	if cancel != nil {
		cancel()
	}
}

// Fetch downloads from a remote, or all remotes when remote is empty.
func (a *App) Fetch(path, remote string) error {
	return a.runNetwork(path, "fetch", func(ctx context.Context, r *session.Repo, o git.NetOptions) error {
		return r.Git.Fetch(ctx, remote, o)
	})
}

// Pull integrates the current branch's upstream. mode is "", "ff-only", "merge" or "rebase".
func (a *App) Pull(path, mode string) error {
	return a.runNetwork(path, "pull", func(ctx context.Context, r *session.Repo, o git.NetOptions) error {
		return a.record(r, "Pull", kindMove, "", func() error { return r.Git.Pull(ctx, git.PullMode(mode), o) })
	})
}

// PushCurrent pushes the checked-out branch: to its upstream if it has one, otherwise to
// the only remote (or "origin") and records that as the upstream.
func (a *App) PushCurrent(path string, force bool) error {
	return a.runNetwork(path, "push", func(ctx context.Context, r *session.Repo, o git.NetOptions) error {
		st, err := r.Git.Status(ctx)
		if err != nil {
			return err
		}
		if st.Head == "" {
			return userError("HEAD is detached. Check out a branch to push it.")
		}
		p := git.PushOptions{Refspec: st.Head}
		if remote, branch, ok := strings.Cut(st.Upstream, "/"); ok {
			p.Remote, p.Refspec = remote, st.Head+":"+branch
		} else {
			remotes, err := r.Git.Remotes(ctx)
			if err != nil {
				return err
			}
			switch {
			case len(remotes) == 0:
				return userError("This repository has no remotes. Add one first.")
			case len(remotes) == 1:
				p.Remote = remotes[0].Name
			default:
				p.Remote = "origin"
				if !hasRemote(remotes, "origin") {
					return userError("This branch has no upstream and there's no \"origin\" remote. Set an upstream first.")
				}
			}
			p.SetUpstream = true
		}
		p.ForceWithLease = force
		return r.Git.Push(ctx, p, o)
	})
}

func hasRemote(rs []git.Remote, name string) bool {
	for _, r := range rs {
		if r.Name == name {
			return true
		}
	}
	return false
}

// PushRef pushes or deletes one ref on a remote: kind is "branch", "tag", "delete-branch"
// or "delete-tag".
func (a *App) PushRef(path, remote, kind, name string) error {
	refspec := map[string]string{
		"branch":        "refs/heads/" + name,
		"tag":           "refs/tags/" + name,
		"delete-branch": ":refs/heads/" + name,
		"delete-tag":    ":refs/tags/" + name,
	}[kind]
	if refspec == "" {
		return userError("Unknown push kind " + kind + ".")
	}
	return a.runNetwork(path, "push", func(ctx context.Context, r *session.Repo, o git.NetOptions) error {
		return r.Git.Push(ctx, git.PushOptions{Remote: remote, Refspec: refspec}, o)
	})
}

// Remotes lists the repository's remotes.
func (a *App) Remotes(path string) ([]git.Remote, error) {
	r, err := a.openRepo(path)
	if err != nil {
		return nil, err
	}
	return r.Git.Remotes(context.Background())
}

// RemoteAction edits remotes: add (name, url), rename (from, to), seturl (name, url),
// remove (name).
func (a *App) RemoteAction(path, action string, args []string) error {
	r, err := a.openRepo(path)
	if err != nil {
		return err
	}
	arg := func(i int) string {
		if i < len(args) {
			return strings.TrimSpace(args[i])
		}
		return ""
	}
	ctx := context.Background()
	switch action {
	case "add":
		err = r.Git.AddRemote(ctx, arg(0), arg(1))
	case "rename":
		err = r.Git.RenameRemote(ctx, arg(0), arg(1))
	case "seturl":
		err = r.Git.SetRemoteURL(ctx, arg(0), arg(1))
	case "remove":
		err = r.Git.RemoveRemote(ctx, arg(0))
	default:
		return userError("Unknown remote action " + action + ".")
	}
	if errors.Is(err, git.ErrInvalidName) {
		return userError("That remote name isn't allowed. Use letters, digits, '-', '_' or '.', and don't start with '-'.")
	}
	return branchErr(err)
}

// netErr explains common network failures.
func netErr(err error) error {
	var gerr *git.Error
	if !errors.As(err, &gerr) {
		return writeErr(err)
	}
	s := gerr.Stderr
	switch {
	case strings.Contains(s, "Authentication failed"), strings.Contains(s, "could not read Username"),
		strings.Contains(s, "Invalid username or password"):
		return userError("Authentication failed. Check your credentials (or Git Credential Manager) and try again.\n\n" + string(gitMessage(s)))
	case strings.Contains(s, "returned error: 403"):
		return userError("The server refused this operation (HTTP 403). Your account may not have permission to push to or read this repository.")
	case strings.Contains(s, "returned error: 401"):
		return userError("Authentication failed (HTTP 401). Check your credentials and try again.")
	case strings.Contains(s, "Permission denied (publickey"):
		return userError("The server refused your SSH key. Make sure your key is loaded in ssh-agent and added to your account.\n\n" + string(gitMessage(s)))
	case strings.Contains(s, "Host key verification failed"):
		return userError("The server's SSH host key wasn't trusted, so the connection was stopped.")
	case strings.Contains(s, "Repository not found"), strings.Contains(s, "' does not exist"),
		strings.Contains(s, "does not appear to be a git repository"),
		strings.Contains(s, "repository '") && strings.Contains(s, "' not found"):
		return userError("No repository was found at that URL. Check the URL, and that your account can access the repository.")
	case strings.Contains(s, "Could not resolve host"), strings.Contains(s, "unable to access"):
		return userError("Couldn't reach the remote. Check the URL and your network connection.\n\n" + string(gitMessage(s)))
	case strings.Contains(s, "[rejected]") || strings.Contains(s, "fetch first") || strings.Contains(s, "non-fast-forward"):
		return userError("The remote has commits you don't have. Pull first, or force push (with lease) to overwrite them.")
	case strings.Contains(s, "stale info"):
		return userError("The remote branch moved since you last fetched, so force push was refused to protect it. Fetch, review, then try again.")
	case strings.Contains(s, "Not possible to fast-forward"), strings.Contains(s, "Diverging branches"):
		return userError("Your branch and its upstream have diverged. Pull with merge or rebase instead of fast-forward only.")
	}
	if strings.TrimSpace(s) != "" {
		return gitMessage(s)
	}
	return err
}
