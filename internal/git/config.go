package git

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

// ConfigEntry is one key = value line of git config.
type ConfigEntry struct {
	// Scope is "system", "global", "local", "worktree" or "command"; empty when listing a
	// single file.
	Scope string `json:"scope"`
	// File is the file the entry came from, or empty for command-line and blob config.
	File  string `json:"file"`
	Key   string `json:"key"`
	Value string `json:"value"`
	// Implicit is true for a key with no "= value", which git reads as true.
	Implicit bool `json:"implicit"`
}

// EffectiveConfig lists every config entry git sees from dir, in the order git reads them
// (so for single-valued keys the last one wins). Included files are followed. With an empty
// dir, only system and global config are read.
func (r *Runner) EffectiveConfig(ctx context.Context, dir string) ([]ConfigEntry, error) {
	if dir == "" {
		// Outside any repository, so local config can't leak in from the process's directory.
		dir = os.TempDir()
	}
	out, err := r.Run(ctx, Cmd{
		Dir:      dir,
		Args:     []string{"config", "--list", "--includes", "--show-origin", "--show-scope", "-z"},
		ReadOnly: true,
	})
	if err != nil {
		return nil, configErr(err)
	}
	entries := parseConfigList(out, true)
	for i, e := range entries {
		// Local config is reported relative to the repository.
		if e.File != "" && !filepath.IsAbs(e.File) {
			entries[i].File = filepath.Join(dir, e.File)
		}
	}
	return entries, nil
}

// FileConfig lists the entries written in one config file, without following includes.
// A missing file has no entries.
func (r *Runner) FileConfig(ctx context.Context, file string) ([]ConfigEntry, error) {
	if _, err := os.Stat(file); errors.Is(err, os.ErrNotExist) {
		return nil, nil
	}
	out, err := r.Run(ctx, Cmd{
		Dir:      filepath.Dir(file),
		Args:     []string{"config", "--file", file, "--list", "--no-includes", "-z"},
		ReadOnly: true,
	})
	if err != nil {
		return nil, configErr(err)
	}
	entries := parseConfigList(out, false)
	for i := range entries {
		entries[i].File = file
	}
	return entries, nil
}

// EditConfig runs `git config --file file args...`, such as a set, unset or section rename.
func (r *Runner) EditConfig(ctx context.Context, file string, args ...string) error {
	_, err := r.Run(ctx, Cmd{Dir: filepath.Dir(file), Args: append([]string{"config", "--file", file}, args...)})
	return configErr(err)
}

// ConfigKeys lists every config key the installed git documents, for autocompletion.
func (r *Runner) ConfigKeys(ctx context.Context) ([]string, error) {
	out, err := r.Output(ctx, Cmd{Dir: os.TempDir(), Args: []string{"help", "--config"}, ReadOnly: true})
	if err != nil {
		return nil, err
	}
	var keys []string
	for l := range strings.SplitSeq(out, "\n") {
		// The list ends with a blank line and a note; keys never contain spaces.
		if l = strings.TrimSpace(l); l != "" && !strings.Contains(l, " ") {
			keys = append(keys, l)
		}
	}
	return keys, nil
}

// Commands lists git's built-in commands, so aliases that shadow one can be flagged.
func (r *Runner) Commands(ctx context.Context) ([]string, error) {
	out, err := r.Output(ctx, Cmd{Dir: os.TempDir(), Args: []string{"--list-cmds=builtins"}, ReadOnly: true})
	if err != nil {
		return nil, err
	}
	return strings.Fields(out), nil
}

// ExpandPath expands a leading ~ the way git does for path-typed config, using git's own
// idea of the home directory (HOME, then USERPROFILE on Windows).
func (r *Runner) ExpandPath(ctx context.Context, p string) (string, error) {
	out, err := r.Output(ctx, Cmd{
		Dir:      os.TempDir(),
		Args:     []string{"-c", "gitrevenge.path=" + p, "config", "--type=path", "gitrevenge.path"},
		ReadOnly: true,
	})
	if err != nil {
		return "", err
	}
	return filepath.Clean(filepath.FromSlash(out)), nil
}

// GlobalConfigPath returns the file `git config --global` writes to: GIT_CONFIG_GLOBAL if
// set, else ~/.gitconfig unless only the XDG file ($XDG_CONFIG_HOME/git/config, by default
// ~/.config/git/config) exists. Alt is the other candidate when both exist, since git
// reads both.
func (r *Runner) GlobalConfigPath(ctx context.Context) (path, alt string, err error) {
	if p := r.getenv("GIT_CONFIG_GLOBAL"); p != "" {
		return filepath.Clean(p), "", nil
	}
	home, err := r.ExpandPath(ctx, "~/.gitconfig")
	if err != nil {
		return "", "", err
	}
	xdg := r.getenv("XDG_CONFIG_HOME")
	if xdg == "" {
		if xdg, err = r.ExpandPath(ctx, "~/.config"); err != nil {
			return "", "", err
		}
	}
	xdg = filepath.Join(xdg, "git", "config")
	homeOK, xdgOK := exists(home), exists(xdg)
	switch {
	case homeOK && xdgOK:
		return home, xdg, nil
	case xdgOK:
		return xdg, "", nil
	}
	return home, "", nil
}

// getenv reads a variable from the environment git runs with.
func (r *Runner) getenv(name string) string {
	for _, kv := range r.env {
		if k, v, ok := strings.Cut(kv, "="); ok && k == name {
			return v
		}
	}
	return ""
}

// LocalConfigPath returns the repository's config file. Linked worktrees share the main
// repository's.
func (r *Repo) LocalConfigPath(ctx context.Context) (string, error) {
	dir, err := r.ReadString(ctx, "rev-parse", "--path-format=absolute", "--git-common-dir")
	if err != nil {
		return "", err
	}
	return filepath.Join(filepath.Clean(filepath.FromSlash(dir)), "config"), nil
}

func exists(p string) bool {
	_, err := os.Stat(p)
	return err == nil
}

// parseConfigList parses `git config --list -z` output: entries end in NUL, and a key is
// separated from its value by a newline (absent for an implicit true). With origin, each
// entry is preceded by its scope and origin, each NUL-terminated.
func parseConfigList(out []byte, origin bool) []ConfigEntry {
	fields := strings.Split(string(out), "\x00")
	var entries []ConfigEntry
	for i := 0; i < len(fields); i++ {
		var e ConfigEntry
		if origin {
			if i+2 >= len(fields) {
				break
			}
			e.Scope = fields[i]
			if f, ok := strings.CutPrefix(fields[i+1], "file:"); ok {
				e.File = filepath.Clean(filepath.FromSlash(f))
			}
			i += 2
		}
		kv := fields[i]
		if kv == "" {
			continue
		}
		key, value, ok := strings.Cut(kv, "\n")
		e.Key, e.Value, e.Implicit = key, value, !ok
		entries = append(entries, e)
	}
	return entries
}

// SameConfigKey reports whether two config keys name the same variable. Section and
// variable names ignore case; a subsection (`remote "Origin"`) doesn't.
func SameConfigKey(a, b string) bool {
	split := func(k string) (section, sub, name string) {
		i, j := strings.Index(k, "."), strings.LastIndex(k, ".")
		if i < 0 {
			return strings.ToLower(k), "", ""
		}
		if i == j {
			return strings.ToLower(k[:i]), "", strings.ToLower(k[j+1:])
		}
		return strings.ToLower(k[:i]), k[i+1 : j], strings.ToLower(k[j+1:])
	}
	as, asub, an := split(a)
	bs, bsub, bn := split(b)
	return as == bs && asub == bsub && an == bn
}

// ConfigError is a config file git can't parse, or a rejected key or value.
type ConfigError struct {
	// Line is the 1-based line git complained about, or 0 when not about a line.
	Line    int
	Message string
}

func (e *ConfigError) Error() string { return e.Message }

var badLineRE = regexp.MustCompile(`bad config line (\d+)`)

// configErr turns git config's failures into a *ConfigError with a readable message.
func configErr(err error) error {
	var gerr *Error
	if !errors.As(err, &gerr) {
		return err
	}
	msg := strings.TrimSpace(gerr.Stderr)
	for _, p := range []string{"fatal: ", "error: "} {
		msg = strings.TrimPrefix(msg, p)
	}
	ce := &ConfigError{Message: msg}
	if m := badLineRE.FindStringSubmatch(msg); m != nil {
		ce.Line, _ = strconv.Atoi(m[1])
		ce.Message = "Git can't read line " + m[1] + " of this config."
	}
	switch gerr.ExitCode {
	case 5:
		// Unsetting a key that isn't there; callers decide whether that matters.
		return ErrNoSuchKey
	case 1:
		if ce.Message == "" {
			ce.Message = "That isn't a valid config key or section name."
		}
	}
	if ce.Message == "" {
		return err
	}
	return ce
}

// ErrNoSuchKey is returned when unsetting a key (or value) that isn't in the file.
var ErrNoSuchKey = errors.New("no such config key")
