package gomark

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

func TestLiveServerStructuralHotReload(t *testing.T) {
	dir := t.TempDir()
	mustWrite(t, filepath.Join(dir, "index.md"), "# Home")
	if err := os.MkdirAll(filepath.Join(dir, "docs"), 0o755); err != nil {
		t.Fatal(err)
	}
	mustWrite(t, filepath.Join(dir, "docs", "alpha.md"), "# Alpha")

	ls := &liveServer{app: &App{ContentDir: dir}, hub: newLiveReloadHub()}
	if err := ls.rebuild(); err != nil {
		t.Fatalf("initial rebuild: %v", err)
	}

	if code := getStatus(ls, "/docs/alpha"); code != http.StatusOK {
		t.Fatalf("alpha before = %d, want 200", code)
	}
	if code := getStatus(ls, "/docs/beta"); code != http.StatusNotFound {
		t.Fatalf("beta before = %d, want 404", code)
	}

	// Adding a file should register its route after a rebuild — no restart.
	mustWrite(t, filepath.Join(dir, "docs", "beta.md"), "# Beta")
	if err := ls.rebuild(); err != nil {
		t.Fatalf("rebuild after add: %v", err)
	}
	if code := getStatus(ls, "/docs/beta"); code != http.StatusOK {
		t.Fatalf("beta after add = %d, want 200", code)
	}
	if code := getStatus(ls, "/sitemap.xml"); code != http.StatusOK {
		t.Fatalf("sitemap = %d, want 200", code)
	}

	// Deleting a file should drop its route after a rebuild.
	if err := os.Remove(filepath.Join(dir, "docs", "alpha.md")); err != nil {
		t.Fatal(err)
	}
	if err := ls.rebuild(); err != nil {
		t.Fatalf("rebuild after delete: %v", err)
	}
	if code := getStatus(ls, "/docs/alpha"); code != http.StatusNotFound {
		t.Fatalf("alpha after delete = %d, want 404", code)
	}
}

func mustWrite(t *testing.T, path, body string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
}

func getStatus(h http.Handler, path string) int {
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, path, nil))
	return rec.Code
}
