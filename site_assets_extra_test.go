package gomark

import (
	"fmt"
	"io/fs"
	"testing"
	"testing/fstest"
)

// erroringFS returns a non-ErrNotExist error for every operation, exercising the
// overlay's error-propagation branches (which must not be masked as "missing").
type erroringFS struct{}

func (erroringFS) Open(name string) (fs.File, error) {
	return nil, &fs.PathError{Op: "open", Path: name, Err: fmt.Errorf("permission denied")}
}

func TestOverlayOpenPropagatesError(t *testing.T) {
	base := fstest.MapFS{"a.txt": &fstest.MapFile{Data: []byte("base")}}
	o := overlayFS{over: erroringFS{}, base: base}
	if _, err := o.Open("a.txt"); err == nil {
		t.Fatal("expected overlay open error to propagate")
	}
}

func TestOverlayStatPropagatesError(t *testing.T) {
	base := fstest.MapFS{"a.txt": &fstest.MapFile{Data: []byte("base")}}
	o := overlayFS{over: erroringFS{}, base: base}
	if _, err := o.Stat("a.txt"); err == nil {
		t.Fatal("expected overlay stat error to propagate")
	}
}

func TestOverlayReadDirPropagatesOverlayError(t *testing.T) {
	base := fstest.MapFS{"dir/a.txt": &fstest.MapFile{Data: []byte("x")}}
	o := overlayFS{over: erroringFS{}, base: base}
	// over.ReadDir errors with a non-notexist error -> propagated.
	if _, err := o.ReadDir("dir"); err == nil {
		t.Fatal("expected overlay ReadDir error to propagate")
	}
}

func TestOverlayReadDirPropagatesBaseError(t *testing.T) {
	over := fstest.MapFS{"dir/a.txt": &fstest.MapFile{Data: []byte("x")}}
	o := overlayFS{over: over, base: erroringFS{}}
	if _, err := o.ReadDir("dir"); err == nil {
		t.Fatal("expected base ReadDir error to propagate")
	}
}

func TestOverlayReadDirMissingBoth(t *testing.T) {
	o := overlayFS{over: fstest.MapFS{}, base: fstest.MapFS{}}
	if _, err := o.ReadDir("nonexistent"); err == nil {
		t.Fatal("expected error when directory is missing on both sides")
	}
}

func TestOverlayReadDirMergesAndSorts(t *testing.T) {
	base := fstest.MapFS{
		"dir/a.txt": &fstest.MapFile{Data: []byte("a")},
		"dir/b.txt": &fstest.MapFile{Data: []byte("b")},
	}
	over := fstest.MapFS{
		"dir/b.txt": &fstest.MapFile{Data: []byte("over-b")},
		"dir/c.txt": &fstest.MapFile{Data: []byte("c")},
	}
	o := overlayFS{over: over, base: base}
	entries, err := o.ReadDir("dir")
	if err != nil {
		t.Fatalf("readdir: %v", err)
	}
	var names []string
	for _, e := range entries {
		names = append(names, e.Name())
	}
	if len(names) != 3 || names[0] != "a.txt" || names[1] != "b.txt" || names[2] != "c.txt" {
		t.Fatalf("expected merged sorted [a,b,c], got %v", names)
	}
}

func TestOverlayOpenFallsBackToBase(t *testing.T) {
	base := fstest.MapFS{"only-base.txt": &fstest.MapFile{Data: []byte("base")}}
	over := fstest.MapFS{}
	o := overlayFS{over: over, base: base}
	f, err := o.Open("only-base.txt")
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	f.Close()
}
