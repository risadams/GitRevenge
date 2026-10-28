// Package settings persists app state as a local JSON file (ADR-0005).
package settings

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"sync"
	"time"

	"github.com/risadams/gitrevenge/internal/fsutil"
)

// schemaVersion is bumped when the file format changes incompatibly.
const schemaVersion = 1

// maxRecent caps the recently opened repositories list.
const maxRecent = 20

// Settings is everything the app remembers between launches.
type Settings struct {
	Version int `json:"version"`
	// Recent lists recently opened repositories, most recent first.
	Recent []RecentRepo `json:"recent"`
	// OpenTabs holds the repository paths open as tabs, in tab order.
	OpenTabs []string `json:"openTabs"`
	// ActiveTab is the path of the selected tab, or empty.
	ActiveTab string `json:"activeTab"`
}

// RecentRepo is an entry in the recent repositories list.
type RecentRepo struct {
	Path       string    `json:"path"`
	LastOpened time.Time `json:"lastOpened"`
}

func defaults() Settings {
	return Settings{Version: schemaVersion, Recent: []RecentRepo{}, OpenTabs: []string{}}
}

// clone returns a deep copy so callers can't mutate the store's state.
func (s Settings) clone() Settings {
	s.Recent = slices.Clone(s.Recent)
	s.OpenTabs = slices.Clone(s.OpenTabs)
	return s
}

// AddRecent moves path to the front of the recent list, adding it if needed.
func (s *Settings) AddRecent(path string, now time.Time) {
	s.Recent = slices.DeleteFunc(s.Recent, func(r RecentRepo) bool { return fsutil.SamePath(r.Path, path) })
	s.Recent = slices.Insert(s.Recent, 0, RecentRepo{Path: path, LastOpened: now})
	if len(s.Recent) > maxRecent {
		s.Recent = s.Recent[:maxRecent]
	}
}

// RemoveRecent drops path from the recent list.
func (s *Settings) RemoveRecent(path string) {
	s.Recent = slices.DeleteFunc(s.Recent, func(r RecentRepo) bool { return fsutil.SamePath(r.Path, path) })
}

// DefaultPath returns the settings file location: %APPDATA%\GitRevenge\settings.json on
// Windows and ~/Library/Application Support/GitRevenge/settings.json on macOS.
func DefaultPath() (string, error) {
	dir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "GitRevenge", "settings.json"), nil
}

// Store loads and saves Settings. It is safe for concurrent use.
type Store struct {
	path string

	mu  sync.Mutex
	cur Settings
}

// Open loads settings from path. A missing file yields defaults. A file that can't be
// parsed is renamed aside (settings.json.corrupt-<timestamp>) and defaults are used, so a
// bad file never stops the app from starting.
func Open(path string) (*Store, error) {
	s := &Store{path: path, cur: defaults()}
	data, err := os.ReadFile(path)
	switch {
	case errors.Is(err, fs.ErrNotExist):
		return s, nil
	case err != nil:
		return nil, fmt.Errorf("reading settings: %w", err)
	}

	loaded := defaults()
	if err := json.Unmarshal(data, &loaded); err != nil || loaded.Version > schemaVersion {
		aside := fmt.Sprintf("%s.corrupt-%s", path, time.Now().UTC().Format("20060102T150405Z"))
		if rerr := os.Rename(path, aside); rerr != nil {
			return nil, fmt.Errorf("settings file %s is unreadable and could not be moved aside: %w", path, rerr)
		}
		return s, nil
	}
	loaded.Version = schemaVersion
	if loaded.Recent == nil {
		loaded.Recent = []RecentRepo{}
	}
	if loaded.OpenTabs == nil {
		loaded.OpenTabs = []string{}
	}
	s.cur = loaded
	return s, nil
}

// Path returns the settings file location.
func (s *Store) Path() string { return s.path }

// Get returns a copy of the current settings.
func (s *Store) Get() Settings {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.cur.clone()
}

// Update applies fn to a copy of the settings and saves the result. If saving fails, the
// in-memory settings are left unchanged.
func (s *Store) Update(fn func(*Settings)) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	next := s.cur.clone()
	fn(&next)
	if err := writeAtomic(s.path, next); err != nil {
		return err
	}
	s.cur = next
	return nil
}

// writeAtomic writes to a temp file in the same directory, syncs it, then renames it over
// the target, so a crash mid-write never leaves a truncated settings file.
func writeAtomic(path string, v Settings) error {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return err
	}
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return fmt.Errorf("creating settings dir: %w", err)
	}
	tmp, err := os.CreateTemp(dir, filepath.Base(path)+".tmp-*")
	if err != nil {
		return fmt.Errorf("saving settings: %w", err)
	}
	defer func() { _ = os.Remove(tmp.Name()) }() // no-op after a successful rename

	if _, err := tmp.Write(append(data, '\n')); err != nil {
		_ = tmp.Close()
		return fmt.Errorf("saving settings: %w", err)
	}
	if err := tmp.Sync(); err != nil {
		_ = tmp.Close()
		return fmt.Errorf("saving settings: %w", err)
	}
	if err := tmp.Close(); err != nil {
		return fmt.Errorf("saving settings: %w", err)
	}
	if err := os.Rename(tmp.Name(), path); err != nil {
		return fmt.Errorf("saving settings: %w", err)
	}
	return nil
}
