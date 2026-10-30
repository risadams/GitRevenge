package gitconfig

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"slices"
	"strings"
	"time"
)

// DefaultKeep is how many backups are kept per config file.
const DefaultKeep = 20

// stamp names backup files. It has no colons, so it is a valid file name on Windows.
const stamp = "20060102T150405.000Z"

const backupExt = ".gitconfig"

// Backups stores copies of config files from before each save, one folder per file:
// Dir/<FileID>/<timestamp>.gitconfig, with the original path in meta.json.
type Backups struct {
	Dir  string
	Keep int
	// now is replaceable in tests.
	now func() time.Time
}

// Backup is one stored copy.
type Backup struct {
	// ID identifies the backup within its file's folder.
	ID   string    `json:"id"`
	Time time.Time `json:"time"`
	Size int64     `json:"size"`
}

type meta struct {
	Path string `json:"path"`
}

// FileID identifies a config file across sessions: a hash of its absolute path with
// symlinks resolved, case-folded where the file system is case-insensitive (Windows and,
// by default, macOS).
func FileID(path string) string {
	sum := sha256.Sum256([]byte(Normalize(path)))
	return hex.EncodeToString(sum[:8])
}

// Normalize returns the form of path used to compare config files.
func Normalize(path string) string {
	if abs, err := filepath.Abs(path); err == nil {
		path = abs
	}
	if r, err := filepath.EvalSymlinks(path); err == nil {
		path = r
	}
	path = filepath.Clean(path)
	if runtime.GOOS == "windows" || runtime.GOOS == "darwin" {
		path = strings.ToLower(path)
	}
	return path
}

func (b *Backups) folder(path string) string { return filepath.Join(b.Dir, FileID(path)) }

// Add stores data as a backup of path and prunes the oldest beyond Keep.
func (b *Backups) Add(path string, data []byte) (Backup, error) {
	dir := b.folder(path)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return Backup{}, err
	}
	if err := b.writeMeta(dir, path); err != nil {
		return Backup{}, err
	}
	now := time.Now
	if b.now != nil {
		now = b.now
	}
	t := now().UTC().Truncate(time.Millisecond)
	for {
		id := t.Format(stamp)
		f, err := os.OpenFile(filepath.Join(dir, id+backupExt), os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o600)
		if errors.Is(err, os.ErrExist) {
			// Two saves in the same millisecond.
			t = t.Add(time.Millisecond)
			continue
		}
		if err != nil {
			return Backup{}, err
		}
		_, werr := f.Write(data)
		if cerr := f.Close(); werr == nil {
			werr = cerr
		}
		if werr != nil {
			_ = os.Remove(f.Name())
			return Backup{}, werr
		}
		b.prune(dir)
		return Backup{ID: id, Time: t, Size: int64(len(data))}, nil
	}
}

func (b *Backups) writeMeta(dir, path string) error {
	p := filepath.Join(dir, "meta.json")
	if _, err := os.Stat(p); err == nil {
		return nil
	}
	abs, err := filepath.Abs(path)
	if err != nil {
		return err
	}
	data, err := json.MarshalIndent(meta{Path: abs}, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(p, append(data, '\n'), 0o600)
}

// List returns path's backups, newest first.
func (b *Backups) List(path string) ([]Backup, error) {
	entries, err := os.ReadDir(b.folder(path))
	if errors.Is(err, os.ErrNotExist) {
		return []Backup{}, nil
	}
	if err != nil {
		return nil, err
	}
	out := []Backup{}
	for _, e := range entries {
		id, ok := strings.CutSuffix(e.Name(), backupExt)
		if !ok || e.IsDir() {
			continue
		}
		t, err := time.Parse(stamp, id)
		if err != nil {
			continue
		}
		info, err := e.Info()
		if err != nil {
			continue
		}
		out = append(out, Backup{ID: id, Time: t, Size: info.Size()})
	}
	slices.SortFunc(out, func(x, y Backup) int { return y.Time.Compare(x.Time) })
	return out, nil
}

// Read returns the content of one backup of path.
func (b *Backups) Read(path, id string) ([]byte, error) {
	if _, err := time.Parse(stamp, id); err != nil {
		return nil, fmt.Errorf("no backup %q", id)
	}
	return os.ReadFile(filepath.Join(b.folder(path), id+backupExt))
}

func (b *Backups) prune(dir string) {
	keep := b.Keep
	if keep <= 0 {
		keep = DefaultKeep
	}
	list, err := os.ReadDir(dir)
	if err != nil {
		return
	}
	var names []string
	for _, e := range list {
		if strings.HasSuffix(e.Name(), backupExt) {
			names = append(names, e.Name())
		}
	}
	// Timestamps sort chronologically as text.
	slices.Sort(names)
	for _, n := range names[:max(0, len(names)-keep)] {
		_ = os.Remove(filepath.Join(dir, n))
	}
}
