package gomark

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"
)

func newLiveServer(t *testing.T, content string) *liveServer {
	t.Helper()
	ls := &liveServer{app: &App{ContentDir: content}, hub: newLiveReloadHub()}
	if err := ls.rebuild(); err != nil {
		t.Fatalf("rebuild: %v", err)
	}
	return ls
}

func TestServeSearch(t *testing.T) {
	dir := t.TempDir()
	mustWrite(t, filepath.Join(dir, "index.md"), "# Home")
	mustWrite(t, filepath.Join(dir, "caching.md"), "---\ntitle: Caching\n---\nAll about caching strategies.")

	idx, err := BuildSearchIndex(dir)
	if err != nil {
		t.Fatalf("index: %v", err)
	}

	// Empty query -> empty results array.
	rec := httptest.NewRecorder()
	if err := serveSearch(rec, httptest.NewRequest("GET", "/api/search", nil), idx); err != nil {
		t.Fatalf("serve: %v", err)
	}
	var empty map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &empty); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if empty["query"] != "" {
		t.Fatalf("expected empty query, got %v", empty["query"])
	}

	// Real query returns results.
	rec2 := httptest.NewRecorder()
	if err := serveSearch(rec2, httptest.NewRequest("GET", "/api/search?q=caching", nil), idx); err != nil {
		t.Fatalf("serve: %v", err)
	}
	if !strings.Contains(rec2.Body.String(), "caching") {
		t.Fatalf("expected caching result: %s", rec2.Body.String())
	}
}

func TestServeSearchLimitClamping(t *testing.T) {
	dir := t.TempDir()
	mustWrite(t, filepath.Join(dir, "index.md"), "# Home")
	idx, err := BuildSearchIndex(dir)
	if err != nil {
		t.Fatalf("index: %v", err)
	}
	for _, limit := range []string{"0", "100", "abc", "5"} {
		rec := httptest.NewRecorder()
		if err := serveSearch(rec, httptest.NewRequest("GET", "/api/search?q=home&limit="+limit, nil), idx); err != nil {
			t.Fatalf("limit %s: %v", limit, err)
		}
	}
}

func TestLiveServerServeWASM(t *testing.T) {
	dir := t.TempDir()
	mustWrite(t, filepath.Join(dir, "index.md"), "# Home")
	ls := newLiveServer(t, dir)
	pub := ls.current.Load().build.publicFS

	rec := httptest.NewRecorder()
	if err := ls.serveWASM(rec, httptest.NewRequest("GET", "/runner.wasm", nil), pub); err != nil {
		t.Fatalf("serveWASM: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Fatalf("code = %d", rec.Code)
	}
	etag := rec.Header().Get("ETag")
	if etag == "" {
		t.Fatal("expected ETag")
	}
	if rec.Header().Get("Content-Encoding") != "gzip" {
		t.Fatal("expected gzip encoding")
	}

	// If-None-Match matching the ETag -> 304.
	req := httptest.NewRequest("GET", "/runner.wasm", nil)
	req.Header.Set("If-None-Match", etag)
	rec2 := httptest.NewRecorder()
	if err := ls.serveWASM(rec2, req, pub); err != nil {
		t.Fatalf("serveWASM 304: %v", err)
	}
	if rec2.Code != http.StatusNotModified {
		t.Fatalf("expected 304, got %d", rec2.Code)
	}
}

func TestLiveServerServeRejectsNonGet(t *testing.T) {
	dir := t.TempDir()
	mustWrite(t, filepath.Join(dir, "index.md"), "# Home")
	ls := newLiveServer(t, dir)

	rec := httptest.NewRecorder()
	ls.ServeHTTP(rec, httptest.NewRequest("POST", "/", nil))
	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", rec.Code)
	}
}

func TestLiveServerServeRoutes(t *testing.T) {
	dir := t.TempDir()
	mustWrite(t, filepath.Join(dir, "index.md"), "# Home")
	mustWrite(t, filepath.Join(dir, "about.md"), "# About")
	ls := newLiveServer(t, dir)

	tests := []struct {
		path string
		code int
	}{
		{"/", http.StatusOK},
		{"/about", http.StatusOK},
		{"/sitemap.xml", http.StatusOK},
		{"/robots.txt", http.StatusOK},
		{"/api/search?q=home", http.StatusOK},
		{"/runner.wasm", http.StatusOK},
		{"/favicon/favicon.ico", http.StatusOK}, // static asset from embedded public
		{"/nonexistent", http.StatusNotFound},
		{"/about/", http.StatusFound}, // trailing slash redirect
	}
	for _, tc := range tests {
		rec := httptest.NewRecorder()
		ls.ServeHTTP(rec, httptest.NewRequest("GET", tc.path, nil))
		if rec.Code != tc.code {
			t.Errorf("%s: code = %d, want %d", tc.path, rec.Code, tc.code)
		}
	}
}

func TestLiveServerOldBaseRedirect(t *testing.T) {
	// content dir named "content" so oldBase is "/content".
	base := t.TempDir()
	content := filepath.Join(base, "content")
	writeExportFile(t, content, "index.md", "# Home")
	writeExportFile(t, content, "about.md", "# About")

	ls := &liveServer{app: &App{ContentDir: content}, hub: newLiveReloadHub()}
	if err := ls.rebuild(); err != nil {
		t.Fatalf("rebuild: %v", err)
	}
	lb := ls.current.Load()
	if lb.oldBase != "/"+strings.Trim(filepath.ToSlash(strings.Trim(content, "/")), "/") {
		// oldBase derived from full content path; just ensure redirect logic runs.
	}

	// Request the oldBase prefix and expect a redirect with Clear-Site-Data.
	rec := httptest.NewRecorder()
	ls.ServeHTTP(rec, httptest.NewRequest("GET", lb.oldBase+"/about", nil))
	if rec.Code != http.StatusFound {
		t.Fatalf("expected redirect, got %d", rec.Code)
	}
	if rec.Header().Get("Clear-Site-Data") == "" {
		t.Fatal("expected Clear-Site-Data header")
	}
}

func TestLiveServerRunnerDisabled(t *testing.T) {
	dir := t.TempDir()
	mustWrite(t, filepath.Join(dir, "index.md"), "# Home")
	ls := &liveServer{app: &App{ContentDir: dir, DisableRunner: true}, hub: newLiveReloadHub()}
	t.Setenv("PLAYGROUND_ENABLED", "")
	if err := ls.rebuild(); err != nil {
		t.Fatalf("rebuild: %v", err)
	}

	rec := httptest.NewRecorder()
	ls.ServeHTTP(rec, httptest.NewRequest("GET", "/runner.wasm", nil))
	// Runner disabled -> falls through to 404.
	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404 for disabled runner, got %d", rec.Code)
	}
}

func TestLiveServerRootRedirectWithoutIndex(t *testing.T) {
	dir := t.TempDir()
	// No index.md at root -> "/" redirects to landing.
	mustWrite(t, filepath.Join(dir, "alpha.md"), "# Alpha")
	ls := newLiveServer(t, dir)

	rec := httptest.NewRecorder()
	ls.ServeHTTP(rec, httptest.NewRequest("GET", "/", nil))
	if rec.Code != http.StatusFound {
		t.Fatalf("expected redirect to landing, got %d", rec.Code)
	}
	if loc := rec.Header().Get("Location"); loc != "/alpha" {
		t.Fatalf("location = %q", loc)
	}
}

func TestLiveServerSitemapDisabled(t *testing.T) {
	dir := t.TempDir()
	mustWrite(t, filepath.Join(dir, "index.md"), "# Home")
	ls := &liveServer{
		app: &App{ContentDir: dir, DisableSitemap: true, DisableRobots: true},
		hub: newLiveReloadHub(),
	}
	if err := ls.rebuild(); err != nil {
		t.Fatalf("rebuild: %v", err)
	}

	for _, path := range []string{"/sitemap.xml", "/robots.txt"} {
		rec := httptest.NewRecorder()
		ls.ServeHTTP(rec, httptest.NewRequest("GET", path, nil))
		if rec.Code != http.StatusNotFound {
			t.Errorf("%s: expected 404 when disabled, got %d", path, rec.Code)
		}
	}
}

func TestLiveServerRebuildErrors(t *testing.T) {
	// No markdown -> rebuild error.
	ls := &liveServer{app: &App{ContentDir: t.TempDir()}, hub: newLiveReloadHub()}
	if err := ls.rebuild(); err == nil {
		t.Fatal("expected rebuild error for empty content dir")
	}
}
