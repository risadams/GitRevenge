// Package gitconfig edits git config files through a working copy, with validated saves
// and backups (ADR-0007). Reading and writing the INI format is left to git itself.
package gitconfig

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// ErrChangedOnDisk means the file changed outside the editor since it was opened or last
// saved, so saving would overwrite someone else's edit.
var ErrChangedOnDisk = errors.New("the config file changed on disk")

// ErrLocked means git (or another editor) holds the file's .lock.
var ErrLocked = errors.New("the config file is locked by another git process")

// Doc is an open config file. Edits go to a working copy; the file itself is only written
// by Save.
type Doc struct {
	// Path is the config file as the user knows it (not symlink-resolved).
	Path string
	work string
	// base is the file's content when opened or last saved, nil if it didn't exist.
	base []byte
	// eol is the file's line ending, kept on save. The working copy always uses "\n".
	eol string
}

// Open copies path into a new working copy in workDir. A missing file opens empty and is
// created on save.
func Open(path, workDir string) (*Doc, error) {
	d := &Doc{Path: filepath.Clean(path)}
	f, err := os.CreateTemp(workDir, "config-*")
	if err != nil {
		return nil, err
	}
	_ = f.Close()
	d.work = f.Name()
	if err := d.Reload(); err != nil {
		_ = os.Remove(d.work)
		return nil, err
	}
	return d, nil
}

// WorkPath is the working copy, for `git config --file`.
func (d *Doc) WorkPath() string { return d.work }

// Reload discards edits and copies the file again.
func (d *Doc) Reload() error {
	disk, err := readFile(d.Path)
	if err != nil {
		return err
	}
	d.base, d.eol = disk, lineEnding(disk)
	return os.WriteFile(d.work, toLF(disk), 0o600)
}

// Text returns the working copy, with "\n" line endings.
func (d *Doc) Text() (string, error) {
	b, err := os.ReadFile(d.work)
	return string(b), err
}

// SetText replaces the working copy, as the raw editor does.
func (d *Doc) SetText(s string) error {
	return os.WriteFile(d.work, toLF([]byte(s)), 0o600)
}

// Exists reports whether the file existed when opened or last saved.
func (d *Doc) Exists() bool { return d.base != nil }

// Dirty reports whether the working copy differs from the file as opened or last saved.
func (d *Doc) Dirty() (bool, error) {
	work, err := os.ReadFile(d.work)
	if err != nil {
		return false, err
	}
	return !bytes.Equal(work, toLF(d.base)), nil
}

// DiskText returns the file as it is on disk now, or nil if it doesn't exist.
func (d *Doc) DiskText() ([]byte, error) { return readFile(d.Path) }

// ChangedOnDisk reports whether the file differs from the content last opened or saved.
func (d *Doc) ChangedOnDisk() (bool, error) {
	disk, err := readFile(d.Path)
	if err != nil {
		return false, err
	}
	return !sameContent(disk, d.base), nil
}

// Save writes the working copy to the file. validate checks the working copy first (git
// must be able to parse it). Unless force is set, a file changed on disk since it was
// opened is not overwritten. The current file is backed up before it is replaced.
func (d *Doc) Save(validate func(work string) error, backups *Backups, force bool) error {
	if err := validate(d.work); err != nil {
		return err
	}
	disk, err := readFile(d.Path)
	if err != nil {
		return err
	}
	if !force && !sameContent(disk, d.base) {
		return ErrChangedOnDisk
	}
	work, err := os.ReadFile(d.work)
	if err != nil {
		return err
	}
	if disk != nil && bytes.Equal(toLF(disk), work) {
		d.base, d.eol = disk, lineEnding(disk)
		return nil
	}
	data := fromLF(work, d.eol)
	if err := Write(d.Path, data, disk, backups); err != nil {
		return err
	}
	d.base = data
	return nil
}

// Close removes the working copy.
func (d *Doc) Close() { _ = os.Remove(d.work) }

// Write replaces the config file at path with data. If current (its present content) is
// not nil, it is backed up first. A symlinked file (common for dotfiles) is written through
// the link. The write follows git's own locking: data goes to path.lock, which is renamed
// over the file, so git and the app never interleave writes.
func Write(path string, data, current []byte, backups *Backups) error {
	target := path
	if t, err := filepath.EvalSymlinks(path); err == nil {
		target = t
	}
	if current != nil && backups != nil {
		if _, err := backups.Add(path, current); err != nil {
			return fmt.Errorf("backing up %s: %w", path, err)
		}
	}
	if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
		return err
	}
	mode := os.FileMode(0o644)
	if info, err := os.Stat(target); err == nil {
		mode = info.Mode().Perm()
	}
	lock := target + ".lock"
	f, err := os.OpenFile(lock, os.O_WRONLY|os.O_CREATE|os.O_EXCL, mode)
	if errors.Is(err, os.ErrExist) {
		return fmt.Errorf("%w: %s exists", ErrLocked, lock)
	}
	if err != nil {
		return err
	}
	ok := false
	defer func() {
		if !ok {
			_ = f.Close()
			_ = os.Remove(lock)
		}
	}()
	if _, err := f.Write(data); err != nil {
		return err
	}
	if err := f.Sync(); err != nil {
		return err
	}
	if err := f.Close(); err != nil {
		return err
	}
	if err := os.Rename(lock, target); err != nil {
		return err
	}
	ok = true
	return nil
}

// readFile reads path, returning nil (not an error) if it doesn't exist.
func readFile(path string) ([]byte, error) {
	b, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return nil, nil
	}
	if b == nil && err == nil {
		b = []byte{}
	}
	return b, err
}

// sameContent compares file contents where nil means missing, so a missing file and an
// empty one differ.
func sameContent(a, b []byte) bool {
	return (a == nil) == (b == nil) && bytes.Equal(a, b)
}

// lineEnding returns "\r\n" if most of b's lines end that way, else "\n".
func lineEnding(b []byte) string {
	crlf := bytes.Count(b, []byte("\r\n"))
	if crlf > 0 && crlf*2 >= bytes.Count(b, []byte("\n")) {
		return "\r\n"
	}
	return "\n"
}

func toLF(b []byte) []byte { return bytes.ReplaceAll(b, []byte("\r\n"), []byte("\n")) }

func fromLF(b []byte, eol string) []byte {
	if eol == "\n" {
		return b
	}
	return []byte(strings.ReplaceAll(string(b), "\n", eol))
}
