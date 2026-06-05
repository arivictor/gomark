package gomark

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestLiveReloadMiddlewareInjectsIntoHTML(t *testing.T) {
	handler := liveReloadMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = w.Write([]byte("<html><body><h1>hi</h1></body></html>"))
	}))

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))

	body := rec.Body.String()
	if !strings.Contains(body, liveReloadPath) {
		t.Fatalf("expected reload client injected, got: %s", body)
	}
	if strings.Index(body, liveReloadPath) > strings.Index(body, "</body>") {
		t.Fatalf("reload script should sit just before </body>, got: %s", body)
	}
	if !strings.Contains(body, "<h1>hi</h1>") || !strings.HasSuffix(strings.TrimSpace(body), "</body></html>") {
		t.Fatalf("original markup should be preserved around the injection, got: %s", body)
	}
}

func TestLiveReloadMiddlewareLeavesNonHTMLUntouched(t *testing.T) {
	payload := []byte("\x00asm binary")
	handler := liveReloadMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/wasm")
		_, _ = w.Write(payload)
	}))

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/runner.wasm", nil))

	if rec.Body.String() != string(payload) {
		t.Fatalf("non-HTML response was modified: %q", rec.Body.String())
	}
	if strings.Contains(rec.Body.String(), liveReloadPath) {
		t.Fatalf("reload client should not be injected into non-HTML responses")
	}
}

func TestLiveReloadMiddlewareSkipsReloadEndpoint(t *testing.T) {
	called := false
	handler := liveReloadMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		if _, ok := w.(*bufferedResponseWriter); ok {
			t.Fatalf("SSE endpoint must not be buffered by the reload middleware")
		}
	}))

	handler.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, liveReloadPath, nil))
	if !called {
		t.Fatalf("expected reload endpoint handler to run")
	}
}

func TestIsInjectablePath(t *testing.T) {
	cases := map[string]bool{
		"/":                    true,
		"/getting-started":     true,
		"/docs/alpha":          true,
		"/runner.wasm":         false,
		"/vendor/highlight.js": false,
		"/styles.css":          false,
		"/sitemap.xml":         false,
		"/robots.txt":          false,
		"/search-index.json":   false,
		"/api/search":          false,
		"/__gomark/livereload": false,
		"/favicon.ico":         false,
	}
	for p, want := range cases {
		if got := isInjectablePath(p); got != want {
			t.Errorf("isInjectablePath(%q) = %v, want %v", p, got, want)
		}
	}
}

func TestWatchTreeDetectsChanges(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "page.md")
	if err := os.WriteFile(file, []byte("one"), 0o644); err != nil {
		t.Fatal(err)
	}

	changed := make(chan struct{}, 1)
	stop := make(chan struct{})
	defer close(stop)
	go watchTree(dir, 20*time.Millisecond, func() { changed <- struct{}{} }, stop)

	// Give the watcher time to take its baseline snapshot, then modify the file.
	time.Sleep(60 * time.Millisecond)
	if err := os.WriteFile(file, []byte("two — changed"), 0o644); err != nil {
		t.Fatal(err)
	}

	select {
	case <-changed:
	case <-time.After(2 * time.Second):
		t.Fatal("watchTree did not report the file change")
	}
}
