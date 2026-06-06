package gomark

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"testing/fstest"
)

func TestExportBuildSiteError(t *testing.T) {
	// A non-existent content dir makes buildSite fail before any output.
	s := NewSite(WithSiteContentDir(filepath.Join(t.TempDir(), "missing")))
	if err := s.Export(t.TempDir()); err == nil {
		t.Fatal("expected build error for missing content dir")
	}
}

func TestExportMkdirError(t *testing.T) {
	dir := t.TempDir()
	writeExportFile(t, dir, "index.md", "# Home")
	// Output dir whose parent is a regular file -> MkdirAll fails (ENOTDIR).
	blocker := filepath.Join(t.TempDir(), "file")
	if err := os.WriteFile(blocker, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	out := filepath.Join(blocker, "sub")
	if err := NewSite(WithSiteContentDir(dir)).Export(out); err == nil {
		t.Fatal("expected mkdir error for output under a file")
	}
}

func TestExportEmptyOutputDir(t *testing.T) {
	s := NewSite(WithSiteContentDir(t.TempDir()))
	if err := s.Export("   "); err == nil {
		t.Fatal("expected error for empty output dir")
	}
}

func TestExportNoMarkdown(t *testing.T) {
	dir := t.TempDir()
	// A non-markdown file only -> build succeeds but no pages exported.
	if err := os.WriteFile(filepath.Join(dir, "note.txt"), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	s := NewSite(WithSiteContentDir(dir))
	if err := s.Export(t.TempDir()); err == nil {
		t.Fatal("expected error when no markdown pages found")
	}
}

func TestExportRunnerDisabledSkipsRunnerWasm(t *testing.T) {
	dir := t.TempDir()
	writeExportFile(t, dir, "index.md", "# Home")
	out := t.TempDir()
	t.Setenv("PLAYGROUND_ENABLED", "")
	s := NewSite(WithSiteContentDir(dir), WithSiteRunnerEnabled(false))
	if err := s.Export(out); err != nil {
		t.Fatalf("export: %v", err)
	}
	if _, err := os.Stat(filepath.Join(out, "runner.wasm")); !os.IsNotExist(err) {
		t.Fatalf("expected no runner.wasm when runner disabled, err=%v", err)
	}
	// 404 page is always written.
	if _, err := os.Stat(filepath.Join(out, "404.html")); err != nil {
		t.Fatalf("expected 404.html: %v", err)
	}
}

func TestExportSitemapAndRobotsDisabled(t *testing.T) {
	dir := t.TempDir()
	writeExportFile(t, dir, "index.md", "# Home")
	out := t.TempDir()
	s := NewSite(
		WithSiteContentDir(dir),
		WithSiteSitemapEnabled(false),
		WithSiteRobotsEnabled(false),
	)
	if err := s.Export(out); err != nil {
		t.Fatalf("export: %v", err)
	}
	if _, err := os.Stat(filepath.Join(out, "sitemap.xml")); !os.IsNotExist(err) {
		t.Fatal("expected no sitemap.xml when disabled")
	}
	if _, err := os.Stat(filepath.Join(out, "robots.txt")); !os.IsNotExist(err) {
		t.Fatal("expected no robots.txt when disabled")
	}
}

func TestWritePageFileMkdirError(t *testing.T) {
	renderer, err := NewFileTemplateRenderer("", "")
	if err != nil {
		t.Fatalf("renderer: %v", err)
	}
	// Create a file, then try to write a page underneath it (parent is a file).
	base := t.TempDir()
	blocker := filepath.Join(base, "blocker")
	if err := os.WriteFile(blocker, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	target := filepath.Join(blocker, "sub", "index.html")
	if err := writePageFile(renderer, target, PageData{Title: "x"}); err == nil {
		t.Fatal("expected mkdir error when parent path is a file")
	}
}

func TestCopyFSFileOpenError(t *testing.T) {
	a := &App{}
	pub, err := a.publicFS()
	if err != nil {
		t.Fatalf("publicFS: %v", err)
	}
	// A srcPath that does not exist surfaces an open error.
	if err := copyFSFile(pub, "definitely-not-a-real-file.xyz", filepath.Join(t.TempDir(), "out")); err == nil {
		t.Fatal("expected open error for missing source file")
	}
}

func TestCopyFSFileSucceeds(t *testing.T) {
	a := &App{}
	pub, err := a.publicFS()
	if err != nil {
		t.Fatalf("publicFS: %v", err)
	}
	target := filepath.Join(t.TempDir(), "nested", "favicon.ico")
	if err := copyFSFile(pub, "favicon.ico", target); err != nil {
		t.Fatalf("copy: %v", err)
	}
	if _, err := os.Stat(target); err != nil {
		t.Fatalf("expected copied file: %v", err)
	}
}

func TestWriteNotFoundPage(t *testing.T) {
	dir := t.TempDir()
	writeExportFile(t, dir, "index.md", "# Home")
	b, err := (&App{ContentDir: dir}).buildSite(true)
	if err != nil {
		t.Fatalf("build: %v", err)
	}
	out := t.TempDir()
	if err := writeNotFoundPage(b, out); err != nil {
		t.Fatalf("write 404: %v", err)
	}
	data, err := os.ReadFile(filepath.Join(out, "404.html"))
	if err != nil {
		t.Fatalf("read 404: %v", err)
	}
	if !strings.Contains(string(data), "Page not found") {
		t.Fatalf("expected 404 content: %s", string(data))
	}
}

func TestWriteNotFoundPageCreateError(t *testing.T) {
	dir := t.TempDir()
	writeExportFile(t, dir, "index.md", "# Home")
	b, err := (&App{ContentDir: dir}).buildSite(true)
	if err != nil {
		t.Fatalf("build: %v", err)
	}
	// Output dir is actually a file -> os.Create(.../404.html) fails (ENOTDIR).
	blocker := filepath.Join(t.TempDir(), "file")
	if err := os.WriteFile(blocker, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := writeNotFoundPage(b, blocker); err == nil {
		t.Fatal("expected create error writing 404 under a file")
	}
}

func TestCopyFSToFileDestinationFails(t *testing.T) {
	a := &App{}
	pub, err := a.publicFS()
	if err != nil {
		t.Fatalf("publicFS: %v", err)
	}
	// Destination root is a regular file -> copying any file underneath fails.
	blocker := filepath.Join(t.TempDir(), "file")
	if err := os.WriteFile(blocker, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := copyFS(blocker, pub); err == nil {
		t.Fatal("expected copyFS error when destination is a file")
	}
}

func TestWriteDecompressedRunnerMissingSource(t *testing.T) {
	// A source FS without runner.wasm.gz surfaces a read error.
	if err := writeDecompressedRunner(t.TempDir(), fstest.MapFS{}); err == nil {
		t.Fatal("expected error when runner.wasm.gz is absent")
	}
}

func TestWriteDecompressedRunner(t *testing.T) {
	a := &App{}
	pub, err := a.publicFS()
	if err != nil {
		t.Fatalf("publicFS: %v", err)
	}
	out := t.TempDir()
	if err := writeDecompressedRunner(out, pub); err != nil {
		t.Fatalf("decompress: %v", err)
	}
	info, err := os.Stat(filepath.Join(out, "runner.wasm"))
	if err != nil {
		t.Fatalf("stat: %v", err)
	}
	if info.Size() == 0 {
		t.Fatal("expected non-empty decompressed runner")
	}
}
