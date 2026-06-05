package gomark

import (
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestAppPublicFSDefaultsToEmbeddedAssets(t *testing.T) {
	a := App{}
	publicFS, err := a.publicFS()
	if err != nil {
		t.Fatalf("public fs: %v", err)
	}

	entries, err := fs.ReadDir(publicFS, ".")
	if err != nil {
		t.Fatalf("read embedded public dir: %v", err)
	}
	if len(entries) == 0 {
		t.Fatal("expected embedded public assets")
	}

	foundFavicon := false
	for _, entry := range entries {
		if strings.EqualFold(entry.Name(), "favicon.ico") {
			foundFavicon = true
			break
		}
	}
	if !foundFavicon {
		t.Fatal("expected embedded favicon.ico")
	}
}

func TestAppPublicFSOverlaysOnDiskAssets(t *testing.T) {
	dir := t.TempDir()
	// favicon.ico also exists in the embedded tree (override case); og-image.png
	// does not (additive case).
	if err := os.WriteFile(filepath.Join(dir, "favicon.ico"), []byte("custom-icon"), 0o644); err != nil {
		t.Fatalf("write favicon: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "og-image.png"), []byte("custom-og"), 0o644); err != nil {
		t.Fatalf("write og: %v", err)
	}

	a := App{PublicDir: dir}
	publicFS, err := a.publicFS()
	if err != nil {
		t.Fatalf("public fs: %v", err)
	}

	// The on-disk file wins over the bundled one of the same name.
	got, err := fs.ReadFile(publicFS, "favicon.ico")
	if err != nil {
		t.Fatalf("read overlaid favicon: %v", err)
	}
	if string(got) != "custom-icon" {
		t.Fatalf("expected on-disk favicon to win, got %q", got)
	}

	// A file that only exists on disk is served from the overlay.
	got, err = fs.ReadFile(publicFS, "og-image.png")
	if err != nil {
		t.Fatalf("read overlaid og-image: %v", err)
	}
	if string(got) != "custom-og" {
		t.Fatalf("expected on-disk og-image, got %q", got)
	}

	// Bundled assets the overlay does not touch are still available.
	if _, err := fs.Stat(publicFS, "runner.wasm.gz"); err != nil {
		t.Fatalf("expected bundled runner.wasm.gz to remain: %v", err)
	}

	// ReadDir merges both trees, deduplicating by name.
	entries, err := fs.ReadDir(publicFS, ".")
	if err != nil {
		t.Fatalf("read merged dir: %v", err)
	}
	names := map[string]bool{}
	for _, e := range entries {
		if names[e.Name()] {
			t.Fatalf("duplicate entry %q in merged dir", e.Name())
		}
		names[e.Name()] = true
	}
	if !names["og-image.png"] || !names["favicon.ico"] || !names["runner.wasm.gz"] {
		t.Fatalf("merged dir missing expected entries: %v", names)
	}
}

func TestAppPublicFSFallsBackWhenDirMissing(t *testing.T) {
	a := App{PublicDir: filepath.Join(t.TempDir(), "does-not-exist")}
	publicFS, err := a.publicFS()
	if err != nil {
		t.Fatalf("public fs: %v", err)
	}
	// Falls back to embedded assets only.
	if _, err := fs.Stat(publicFS, "favicon.ico"); err != nil {
		t.Fatalf("expected embedded favicon fallback: %v", err)
	}
}

func TestAppPublicFSFallsBackWhenDirIsAFile(t *testing.T) {
	file := filepath.Join(t.TempDir(), "public")
	if err := os.WriteFile(file, []byte("not a dir"), 0o644); err != nil {
		t.Fatalf("write file: %v", err)
	}
	a := App{PublicDir: file}
	publicFS, err := a.publicFS()
	if err != nil {
		t.Fatalf("public fs: %v", err)
	}
	// A non-directory public dir is tolerated: embedded assets are used.
	if _, err := fs.Stat(publicFS, "favicon.ico"); err != nil {
		t.Fatalf("expected embedded favicon fallback: %v", err)
	}
}

func TestStaticFileExistsUsesFSAndBlocksTraversal(t *testing.T) {
	a := App{}
	publicFS, err := a.publicFS()
	if err != nil {
		t.Fatalf("public fs: %v", err)
	}

	exists, err := staticFileExists(publicFS, "/favicon.ico")
	if err != nil {
		t.Fatalf("static file exists: %v", err)
	}
	if !exists {
		t.Fatal("expected favicon.ico to exist")
	}

	exists, err = staticFileExists(publicFS, "/../runtime.go")
	if err != nil {
		t.Fatalf("static file traversal check: %v", err)
	}
	if exists {
		t.Fatal("did not expect path traversal to resolve")
	}
}
