package gomark

import (
	"embed"
	"errors"
	"fmt"
	"io/fs"
	"log"
	"os"
	"path"
	"sort"
	"strings"
)

//go:embed public/*
var embeddedPublicFS embed.FS

// publicFS returns the asset tree that serves static files (favicons, og images,
// vendored JS/CSS, the runner module). It is always backed by the embedded
// public/ tree; when an on-disk PublicDir is configured, its files are overlaid
// on top so a site can override the bundled assets or add its own.
func (a *App) publicFS() (fs.FS, error) {
	base, err := fs.Sub(embeddedPublicFS, "public")
	if err != nil {
		return nil, err
	}

	dir := a.publicDir()
	if dir == "" {
		return base, nil
	}

	// A missing or non-directory public dir is tolerated (warn and use the
	// bundled assets), but a real stat error — permissions, I/O — is surfaced so
	// a misconfiguration doesn't silently ship the wrong assets.
	info, statErr := os.Stat(dir)
	switch {
	case statErr == nil && info.IsDir():
		return overlayFS{over: os.DirFS(dir), base: base}, nil
	case errors.Is(statErr, fs.ErrNotExist):
		log.Printf("public dir %q not found; using embedded assets only", dir)
		return base, nil
	case statErr == nil && !info.IsDir():
		log.Printf("public dir %q is not a directory; using embedded assets only", dir)
		return base, nil
	default:
		return nil, fmt.Errorf("stat public dir %q: %w", dir, statErr)
	}
}

// overlayFS layers an on-disk file system (over) on top of a base file system,
// with over winning for any path present in both. It implements the fs
// interfaces that gomark relies on: Open and Stat (live serving) and ReadDir
// (directory listing and the export walk).
type overlayFS struct {
	over fs.FS // on-disk overlay; takes precedence
	base fs.FS // embedded defaults
}

func (o overlayFS) Open(name string) (fs.File, error) {
	f, err := o.over.Open(name)
	if err == nil {
		return f, nil
	}
	// Fall back to the embedded asset only when the overlay simply lacks the
	// file; a real error (permissions, I/O) is propagated rather than masked by
	// silently serving the bundled version.
	if !errors.Is(err, fs.ErrNotExist) {
		return nil, err
	}
	return o.base.Open(name)
}

func (o overlayFS) Stat(name string) (fs.FileInfo, error) {
	fi, err := fs.Stat(o.over, name)
	if err == nil {
		return fi, nil
	}
	if !errors.Is(err, fs.ErrNotExist) {
		return nil, err
	}
	return fs.Stat(o.base, name)
}

// ReadDir merges both directories, deduplicating by name with the overlay entry
// winning. Entries are returned sorted, matching the embed/os.DirFS contract.
func (o overlayFS) ReadDir(name string) ([]fs.DirEntry, error) {
	merged := map[string]fs.DirEntry{}

	// A directory may live in only one side, so a missing dir (ErrNotExist) on
	// either is expected; any other error (permissions, I/O) is propagated so an
	// overlay's contents are never silently dropped from a listing or export.
	baseEntries, baseErr := fs.ReadDir(o.base, name)
	if baseErr != nil && !errors.Is(baseErr, fs.ErrNotExist) {
		return nil, baseErr
	}
	for _, e := range baseEntries {
		merged[e.Name()] = e
	}

	overEntries, overErr := fs.ReadDir(o.over, name)
	if overErr != nil && !errors.Is(overErr, fs.ErrNotExist) {
		return nil, overErr
	}
	for _, e := range overEntries {
		merged[e.Name()] = e // overlay wins
	}

	// Neither side has the directory at all: report it as missing.
	if errors.Is(baseErr, fs.ErrNotExist) && errors.Is(overErr, fs.ErrNotExist) {
		return nil, baseErr
	}

	names := make([]string, 0, len(merged))
	for n := range merged {
		names = append(names, n)
	}
	sort.Strings(names)

	out := make([]fs.DirEntry, 0, len(names))
	for _, n := range names {
		out = append(out, merged[n])
	}
	return out, nil
}

func staticFileExists(publicFS fs.FS, requestPath string) (bool, error) {
	relReq := strings.TrimPrefix(requestPath, "/")
	relReq = path.Clean(relReq)

	if relReq == "." || relReq == "" || strings.HasPrefix(relReq, "../") || relReq == ".." {
		return false, nil
	}

	info, statErr := fs.Stat(publicFS, relReq)
	if statErr != nil {
		if errors.Is(statErr, fs.ErrNotExist) {
			return false, nil
		}
		return false, statErr
	}

	return !info.IsDir(), nil
}
