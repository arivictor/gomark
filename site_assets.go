package gomark

import (
	"embed"
	"errors"
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

	info, statErr := os.Stat(dir)
	if statErr != nil || !info.IsDir() {
		// A configured-but-missing public dir is a likely misconfiguration; warn
		// and fall back to the embedded assets rather than failing the build.
		log.Printf("public dir %q not found or not a directory; using embedded assets only", dir)
		return base, nil
	}

	return overlayFS{over: os.DirFS(dir), base: base}, nil
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
	if f, err := o.over.Open(name); err == nil {
		return f, nil
	}
	return o.base.Open(name)
}

func (o overlayFS) Stat(name string) (fs.FileInfo, error) {
	if fi, err := fs.Stat(o.over, name); err == nil {
		return fi, nil
	}
	return fs.Stat(o.base, name)
}

// ReadDir merges both directories, deduplicating by name with the overlay entry
// winning. Entries are returned sorted, matching the embed/os.DirFS contract.
func (o overlayFS) ReadDir(name string) ([]fs.DirEntry, error) {
	merged := map[string]fs.DirEntry{}

	baseEntries, baseErr := fs.ReadDir(o.base, name)
	for _, e := range baseEntries {
		merged[e.Name()] = e
	}

	overEntries, overErr := fs.ReadDir(o.over, name)
	for _, e := range overEntries {
		merged[e.Name()] = e // overlay wins
	}

	// Surface an error only when neither side could read the directory.
	if baseErr != nil && overErr != nil {
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
