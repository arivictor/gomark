package gomark

import (
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"
)

func TestLandingRoute(t *testing.T) {
	// Prefers root "/" when present.
	if got := landingRoute(map[string]string{"/": "x", "/about": "y"}); got != "/" {
		t.Fatalf("expected /, got %q", got)
	}
	// Otherwise lexically first.
	if got := landingRoute(map[string]string{"/zeta": "z", "/alpha": "a", "/mid": "m"}); got != "/alpha" {
		t.Fatalf("expected /alpha, got %q", got)
	}
	// Empty -> "/".
	if got := landingRoute(map[string]string{}); got != "/" {
		t.Fatalf("expected / for empty, got %q", got)
	}
}

func TestRegisterContentRoutesServesPages(t *testing.T) {
	dir := t.TempDir()
	mustWrite(t, filepath.Join(dir, "index.md"), "# Home")
	mustWrite(t, filepath.Join(dir, "about.md"), "# About")

	b, err := (&App{ContentDir: dir}).buildSite(false)
	if err != nil {
		t.Fatalf("build: %v", err)
	}

	srv := NewServer(b.errorResponder(nil))
	landing, err := registerContentRoutes(srv, b)
	if err != nil {
		t.Fatalf("register: %v", err)
	}
	if landing != "/" {
		t.Fatalf("landing = %q", landing)
	}

	rec := httptest.NewRecorder()
	srv.mux.ServeHTTP(rec, httptest.NewRequest("GET", "/about", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("about code = %d", rec.Code)
	}

	rootRec := httptest.NewRecorder()
	srv.mux.ServeHTTP(rootRec, httptest.NewRequest("GET", "/", nil))
	if rootRec.Code != http.StatusOK {
		t.Fatalf("root code = %d", rootRec.Code)
	}
}

func TestRegisterContentRoutesNoMarkdown(t *testing.T) {
	dir := t.TempDir()
	b, err := (&App{ContentDir: dir}).buildSite(false)
	if err != nil {
		t.Fatalf("build: %v", err)
	}
	srv := NewServer(b.errorResponder(nil))
	if _, err := registerContentRoutes(srv, b); err == nil {
		t.Fatal("expected error when no markdown files exist")
	}
}

func TestRegisterContentRoutesCollision(t *testing.T) {
	dir := t.TempDir()
	// Two files that resolve to the same route via permalink override.
	mustWrite(t, filepath.Join(dir, "a.md"), "---\npermalink: /dup\n---\n# A")
	mustWrite(t, filepath.Join(dir, "b.md"), "---\npermalink: /dup\n---\n# B")

	b, err := (&App{ContentDir: dir}).buildSite(false)
	if err != nil {
		t.Fatalf("build: %v", err)
	}
	srv := NewServer(b.errorResponder(nil))
	if _, err := registerContentRoutes(srv, b); err == nil {
		t.Fatal("expected route collision error")
	}
}

func TestRenderContentPageNotFound(t *testing.T) {
	dir := t.TempDir()
	mustWrite(t, filepath.Join(dir, "index.md"), "# Home")
	b, err := (&App{ContentDir: dir}).buildSite(false)
	if err != nil {
		t.Fatalf("build: %v", err)
	}
	rec := httptest.NewRecorder()
	err = b.renderContentPage(rec, httptest.NewRequest("GET", "/ghost", nil), "/ghost", "ghost")
	httpErr, ok := err.(*HTTPError)
	if !ok || httpErr.Status != http.StatusNotFound {
		t.Fatalf("expected 404 HTTPError, got %v", err)
	}
}

func TestRenderContentPageUsesSlugTitleWhenEmpty(t *testing.T) {
	dir := t.TempDir()
	// No frontmatter title and no leading H1 -> title derived from slug.
	mustWrite(t, filepath.Join(dir, "index.md"), "# Home")
	mustWrite(t, filepath.Join(dir, "getting-started.md"), "Some body text without a heading.")
	b, err := (&App{ContentDir: dir}).buildSite(false)
	if err != nil {
		t.Fatalf("build: %v", err)
	}
	rec := httptest.NewRecorder()
	if err := b.renderContentPage(rec, httptest.NewRequest("GET", "/getting-started", nil), "/getting-started", "getting-started"); err != nil {
		t.Fatalf("render: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Fatalf("code = %d", rec.Code)
	}
}

func TestEachContentRouteIndexHandling(t *testing.T) {
	dir := t.TempDir()
	writeExportFile(t, dir, "index.md", "# Home")
	writeExportFile(t, dir, filepath.Join("guides", "index.md"), "# Guides")
	writeExportFile(t, dir, filepath.Join("guides", "install.md"), "# Install")

	routes := map[string]string{}
	err := eachContentRoute(dir, func(slug, route, path string) error {
		routes[route] = slug
		return nil
	})
	if err != nil {
		t.Fatalf("walk: %v", err)
	}

	if _, ok := routes["/"]; !ok {
		t.Errorf("expected root index route /, got %v", routes)
	}
	if _, ok := routes["/guides"]; !ok {
		t.Errorf("expected folder index route /guides, got %v", routes)
	}
	if _, ok := routes["/guides/install"]; !ok {
		t.Errorf("expected /guides/install, got %v", routes)
	}
}
