package gomark

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestBuildSearchIndexMissingDir(t *testing.T) {
	if _, err := BuildSearchIndex(filepath.Join(t.TempDir(), "missing")); err == nil {
		t.Fatal("expected error for missing content dir")
	}
}

func TestMakeSnippetSkipsEmptyTerms(t *testing.T) {
	// An empty term in the list must be skipped before a real term matches.
	got := makeSnippet("alpha beta gamma needle delta", []string{"", "needle"})
	if !strings.Contains(got, "needle") {
		t.Fatalf("expected needle in snippet: %q", got)
	}
}

func TestQueryTieBreakByTitleAndPath(t *testing.T) {
	dir := t.TempDir()
	// Two docs with identical bodies/titles structure -> equal scores, forcing the
	// stable sort tie-break on Title then Path.
	writeSearchMarkdown(t, dir, "index.md", "---\ntitle: Home\n---\n")
	writeSearchMarkdown(t, dir, "bbb.md", "---\ntitle: Topic\n---\nshared keyword here")
	writeSearchMarkdown(t, dir, "aaa.md", "---\ntitle: Topic\n---\nshared keyword here")

	idx, err := BuildSearchIndex(dir)
	if err != nil {
		t.Fatalf("index: %v", err)
	}
	results := idx.Query("keyword", 10)
	if len(results) < 2 {
		t.Fatalf("expected 2 results, got %d", len(results))
	}
	// Equal title -> lexically smaller path wins.
	if results[0].Path != "/aaa" {
		t.Fatalf("expected /aaa first on path tie-break, got %q", results[0].Path)
	}
}

func TestNewFileTemplateRendererBadGlob(t *testing.T) {
	// An invalid glob pattern surfaces filepath.ErrBadPattern.
	if _, err := NewFileTemplateRenderer("templates/layout.html", "[invalid"); err == nil {
		t.Fatal("expected error for bad glob pattern")
	}
}

func TestNewFilesystemFileTemplateRendererParseError(t *testing.T) {
	dir := t.TempDir()
	layout := filepath.Join(dir, "layout.html")
	page := filepath.Join(dir, "markdown.html")
	if err := os.WriteFile(layout, []byte(`{{define "layout"}}ok{{end}}`), 0o644); err != nil {
		t.Fatal(err)
	}
	// Unclosed action -> template parse error.
	if err := os.WriteFile(page, []byte(`{{define "content"}}{{ .Title `), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := NewFileTemplateRenderer(layout, filepath.Join(dir, "*.html")); err == nil {
		t.Fatal("expected parse error for malformed template")
	}
}

func TestRenderToExecutionError(t *testing.T) {
	dir := t.TempDir()
	layout := filepath.Join(dir, "layout.html")
	page := filepath.Join(dir, "markdown.html")
	// Calling a method that does not exist on the data forces an execution error.
	if err := os.WriteFile(layout, []byte(`{{define "layout"}}{{.Title.NoSuchMethod}}{{end}}`), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(page, []byte(`{{define "content"}}x{{end}}`), 0o644); err != nil {
		t.Fatal(err)
	}
	r, err := NewFileTemplateRenderer(layout, filepath.Join(dir, "*.html"))
	if err != nil {
		t.Fatalf("renderer: %v", err)
	}
	var sb strings.Builder
	if err := r.RenderTo(&sb, "markdown", PageData{Title: "x"}); err == nil {
		t.Fatal("expected execution error")
	}
}

func TestInjectLiveReloadAppendsWhenNoBodyTag(t *testing.T) {
	handler := liveReloadMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = w.Write([]byte("<p>no body tag here</p>"))
	}))
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, httptest.NewRequest("GET", "/page", nil))
	body := rec.Body.String()
	if !strings.Contains(body, liveReloadPath) {
		t.Fatalf("expected reload snippet appended: %s", body)
	}
	// Snippet should be appended at the very end since there is no </body>.
	if !strings.HasSuffix(strings.TrimSpace(body), "</script>") {
		t.Fatalf("expected snippet at end: %s", body)
	}
}

func TestInjectLiveReloadDirect(t *testing.T) {
	// No closing body -> appended.
	out := injectLiveReload([]byte("<p>hi</p>"))
	if !strings.Contains(string(out), liveReloadPath) {
		t.Fatal("expected snippet appended")
	}
	// With closing body -> inserted before it.
	out = injectLiveReload([]byte("<body>hi</body>"))
	s := string(out)
	if strings.Index(s, liveReloadPath) > strings.Index(s, "</body>") {
		t.Fatal("expected snippet before </body>")
	}
}

func TestStaticFileExistsRootAndDotPaths(t *testing.T) {
	a := App{}
	pub, err := a.publicFS()
	if err != nil {
		t.Fatalf("publicFS: %v", err)
	}
	for _, p := range []string{"/", "/.", "/..", "/../x"} {
		exists, err := staticFileExists(pub, p)
		if err != nil {
			t.Fatalf("%s: %v", p, err)
		}
		if exists {
			t.Fatalf("%s should not exist as a static file", p)
		}
	}
	// A directory in the public tree is not a servable static file.
	if exists, _ := staticFileExists(pub, "/vendor"); exists {
		t.Fatal("a directory should not count as a static file")
	}
}
