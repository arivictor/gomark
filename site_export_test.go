package gomark

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestExportFilePath(t *testing.T) {
	cases := map[string]string{
		"/":               filepath.Join("out", "index.html"),
		"/guides/install": filepath.Join("out", "guides", "install", "index.html"),
		"/guides/":        filepath.Join("out", "guides", "index.html"),
		"/changelog":      filepath.Join("out", "changelog", "index.html"),
	}
	for route, want := range cases {
		if got := exportFilePath("out", route); got != want {
			t.Fatalf("exportFilePath(%q) = %q, want %q", route, got, want)
		}
	}
}

func writeExportFile(t *testing.T, root, rel, content string) {
	t.Helper()
	full := filepath.Join(root, rel)
	if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
		t.Fatalf("mkdir %s: %v", rel, err)
	}
	if err := os.WriteFile(full, []byte(content), 0o644); err != nil {
		t.Fatalf("write %s: %v", rel, err)
	}
}

func TestSiteExportProducesStaticSite(t *testing.T) {
	content := t.TempDir()
	writeExportFile(t, content, "index.md", "---\ntitle: Home\n---\n# Home\n\nWelcome.")
	writeExportFile(t, content, "guides/install.md", "---\ntitle: Install\n---\n# Install\n\nRun go get.")
	writeExportFile(t, content, "changelog.md", "---\ntitle: Changelog\npermalink: /releases\n---\n# Changelog\n\nNotes.")

	out := t.TempDir()
	s := NewSite(
		WithSiteContentDir(content),
		WithSiteURL("https://docs.example.com"),
		WithSiteMode(PreRender),
	)
	if err := s.Export(out); err != nil {
		t.Fatalf("export: %v", err)
	}

	for _, rel := range []string{
		"index.html",
		filepath.Join("guides", "install", "index.html"),
		filepath.Join("releases", "index.html"), // permalink override
		"sitemap.xml",
		"robots.txt",
		"search-index.json",
		"runner.wasm", // decompressed runner
		filepath.Join("vendor", "highlight.min.js"),
	} {
		if _, err := os.Stat(filepath.Join(out, rel)); err != nil {
			t.Fatalf("expected exported file %s: %v", rel, err)
		}
	}

	home, err := os.ReadFile(filepath.Join(out, "index.html"))
	if err != nil {
		t.Fatalf("read home: %v", err)
	}
	homeStr := string(home)
	if !strings.Contains(homeStr, `<link rel="canonical" href="https://docs.example.com/"`) {
		t.Fatalf("expected absolute canonical from site URL in home page")
	}
	if strings.Contains(homeStr, "csrf-token") {
		t.Fatalf("did not expect a csrf-token meta tag in a static build")
	}
	if !strings.Contains(homeStr, `data-static="true"`) {
		t.Fatalf("expected data-static flag on body")
	}

	idxJSON, err := os.ReadFile(filepath.Join(out, "search-index.json"))
	if err != nil {
		t.Fatalf("read search index: %v", err)
	}
	var entries []SearchEntry
	if err := json.Unmarshal(idxJSON, &entries); err != nil {
		t.Fatalf("unmarshal search index: %v", err)
	}
	if len(entries) != 3 {
		t.Fatalf("expected 3 search entries, got %d", len(entries))
	}
	foundPermalink := false
	for _, e := range entries {
		if e.Path == "/releases" {
			foundPermalink = true
		}
	}
	if !foundPermalink {
		t.Fatalf("expected search entry to use the permalink /releases: %+v", entries)
	}
}

func TestCopyFSCopiesEmbeddedPublicAssets(t *testing.T) {
	a := &App{}
	pub, err := a.publicFS()
	if err != nil {
		t.Fatalf("publicFS: %v", err)
	}
	dst := t.TempDir()
	if err := copyFS(dst, pub); err != nil {
		t.Fatalf("copyFS: %v", err)
	}
	for _, rel := range []string{"favicon.ico", "wasm_exec.js", filepath.Join("vendor", "lucide.min.js")} {
		if _, err := os.Stat(filepath.Join(dst, rel)); err != nil {
			t.Fatalf("expected copied asset %s: %v", rel, err)
		}
	}
}

func TestRenderToMatchesRenderStatusBody(t *testing.T) {
	renderer, err := NewFileTemplateRenderer("", "")
	if err != nil {
		t.Fatalf("renderer: %v", err)
	}
	data := PageData{Title: "Parity", SiteName: "GoMark", CurrentPath: "/x"}

	var buf strings.Builder
	if err := renderer.RenderTo(&buf, "markdown", data); err != nil {
		t.Fatalf("RenderTo: %v", err)
	}
	if !strings.Contains(buf.String(), "Parity") {
		t.Fatalf("expected rendered title in RenderTo output")
	}
}
