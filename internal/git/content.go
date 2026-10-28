package git

import (
	"bytes"
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// MaxTextBytes is the size above which a file is not shown as text unless forced.
const MaxTextBytes = 2 << 20

// MaxForcedBytes caps a file shown after the user chose "load anyway".
const MaxForcedBytes = 32 << 20

// imageTypes maps extensions to MIME types the UI can display.
var imageTypes = map[string]string{
	".png": "image/png", ".jpg": "image/jpeg", ".jpeg": "image/jpeg", ".gif": "image/gif",
	".webp": "image/webp", ".bmp": "image/bmp", ".ico": "image/x-icon", ".svg": "image/svg+xml",
	".avif": "image/avif",
}

// Side is one version of a file being compared.
type Side struct {
	// Exists is false when the file is absent on this side (added or deleted).
	Exists bool `json:"exists"`
	// Text holds the contents for text files.
	Text string `json:"text"`
	// Image holds base64 contents for image files.
	Image string `json:"image,omitempty"`
	Size  int64  `json:"size"`
}

// FileContents is both versions of a file, ready for the diff viewer.
type FileContents struct {
	Path string `json:"path"`
	Old  Side   `json:"old"`
	New  Side   `json:"new"`
	// Kind is "text", "image", "binary" or "large".
	Kind string `json:"kind"`
	// Mime is set for images.
	Mime string `json:"mime,omitempty"`
}

// FileRef names one version of a file being compared.
type FileRef struct {
	// Rev is a commit, "" for the work tree, or ":" for the index.
	Rev  string
	Path string
}

// IndexVersion names the staged version of a file.
func IndexVersion(path string) FileRef { return FileRef{Rev: ":", Path: path} }

// WorkTreeVersion names the on-disk version of a file.
func WorkTreeVersion(path string) FileRef { return FileRef{Path: path} }

// Contents reads two versions of a file (from → to) for display. Text is returned in work-tree form
// (smudge filters such as line-ending conversion applied), so both sides compare cleanly
// with the file on disk. Binary and oversized files are reported by Kind without text;
// force raises the size limit to MaxForcedBytes.
func (r *Repo) Contents(ctx context.Context, display string, from, to FileRef, force bool) (FileContents, error) {
	fc := FileContents{Path: display, Kind: "text"}
	mime, isImage := imageTypes[strings.ToLower(filepath.Ext(display))]
	limit := int64(MaxTextBytes)
	if force || isImage {
		limit = MaxForcedBytes
	}
	var olds, news []byte
	var err error
	if fc.Old, olds, err = r.readSide(ctx, from, limit); err != nil {
		return fc, err
	}
	if fc.New, news, err = r.readSide(ctx, to, limit); err != nil {
		return fc, err
	}
	switch {
	case fc.Old.Size > limit || fc.New.Size > limit:
		fc.Kind = "large"
	case isImage:
		fc.Kind, fc.Mime = "image", mime
		fc.Old.Image = base64.StdEncoding.EncodeToString(olds)
		fc.New.Image = base64.StdEncoding.EncodeToString(news)
	case isBinary(olds) || isBinary(news):
		fc.Kind = "binary"
	default:
		fc.Old.Text, fc.New.Text = string(olds), string(news)
	}
	return fc, nil
}

// readSide returns a version's metadata and, when within limit, its bytes.
func (r *Repo) readSide(ctx context.Context, v FileRef, limit int64) (Side, []byte, error) {
	if v.Path == "" {
		return Side{}, nil, nil
	}
	if v.Rev == "" {
		p, err := r.inside(v.Path)
		if err != nil {
			return Side{}, nil, err
		}
		info, err := os.Stat(p)
		if errors.Is(err, os.ErrNotExist) {
			return Side{}, nil, nil
		}
		if err != nil {
			return Side{}, nil, err
		}
		s := Side{Exists: true, Size: info.Size()}
		if s.Size > limit {
			return s, nil, nil
		}
		b, err := os.ReadFile(p)
		return s, b, err
	}

	spec := v.Rev + v.Path
	if v.Rev != ":" {
		spec = v.Rev + ":" + v.Path
	}
	sizeOut, err := r.Read(ctx, "cat-file", "-s", spec)
	if err != nil {
		var gerr *Error
		if errors.As(err, &gerr) && gerr.ExitCode == 128 {
			return Side{}, nil, nil // not present in that version
		}
		return Side{}, nil, err
	}
	var size int64
	_, _ = fmt.Sscan(strings.TrimSpace(string(sizeOut)), &size)
	s := Side{Exists: true, Size: size}
	if size > limit {
		return s, nil, nil
	}
	// --filters applies smudge filters (e.g. autocrlf, LFS) to give work-tree form.
	b, err := r.Read(ctx, "cat-file", "--filters", spec)
	return s, b, err
}

// isBinary uses git's heuristic: a NUL byte in the first 8000 bytes.
func isBinary(b []byte) bool {
	return bytes.IndexByte(b[:min(len(b), 8000)], 0) >= 0
}

// inside resolves a work-tree relative path, refusing anything outside the repository.
func (r *Repo) inside(rel string) (string, error) {
	root, err := filepath.Abs(r.dir)
	if err != nil {
		return "", err
	}
	p := filepath.Join(root, filepath.FromSlash(rel))
	if back, err := filepath.Rel(root, p); err != nil || back == ".." || strings.HasPrefix(back, ".."+string(filepath.Separator)) {
		return "", fmt.Errorf("%q is outside the repository", rel)
	}
	return p, nil
}

// SetIndexContent replaces a file's staged content. content is in work-tree form; git's
// clean filters for that path (line endings, LFS) are applied before it's stored. This is
// how hunks and individual lines are staged or unstaged.
func (r *Repo) SetIndexContent(ctx context.Context, path string, content string) error {
	sha, err := r.runner.Output(ctx, Cmd{
		Dir:   r.dir,
		Args:  []string{"hash-object", "-w", "--stdin", "--path=" + path},
		Stdin: strings.NewReader(content),
	})
	if err != nil {
		return err
	}
	mode, err := r.indexMode(ctx, path)
	if err != nil {
		return err
	}
	_, err = r.Write(ctx, Cmd{Args: []string{"update-index", "--add", "--cacheinfo", mode + "," + sha + "," + path}})
	return err
}

// indexMode returns the file's mode in the index, or one derived from the work tree for a
// file not yet tracked.
func (r *Repo) indexMode(ctx context.Context, path string) (string, error) {
	out, err := r.runner.Output(ctx, Cmd{Dir: r.dir, Args: []string{"ls-files", "-s", "--", path}, ReadOnly: true, Env: []string{"GIT_LITERAL_PATHSPECS=1"}})
	if err != nil {
		return "", err
	}
	if mode, _, ok := strings.Cut(out, " "); ok && mode != "" {
		return mode, nil
	}
	p, err := r.inside(path)
	if err != nil {
		return "", err
	}
	if info, err := os.Stat(p); err == nil && info.Mode()&0o111 != 0 {
		return "100755", nil
	}
	return "100644", nil
}

// SetWorkTreeContent overwrites a file on disk, keeping its permissions. Used to discard
// selected hunks or lines. This cannot be undone.
func (r *Repo) SetWorkTreeContent(path string, content string) error {
	p, err := r.inside(path)
	if err != nil {
		return err
	}
	mode := os.FileMode(0o644)
	if info, err := os.Stat(p); err == nil {
		mode = info.Mode().Perm()
	}
	f, err := os.OpenFile(p, os.O_WRONLY|os.O_TRUNC|os.O_CREATE, mode)
	if err != nil {
		return err
	}
	if _, err := io.WriteString(f, content); err != nil {
		_ = f.Close()
		return err
	}
	return f.Close()
}
