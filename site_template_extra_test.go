package gomark

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestNewFileTemplateRendererRequiresBothPaths(t *testing.T) {
	if _, err := NewFileTemplateRenderer("templates/layout.html", ""); err == nil {
		t.Fatal("expected error when only layoutPath is set")
	}
	if _, err := NewFileTemplateRenderer("", "templates/*.html"); err == nil {
		t.Fatal("expected error when only pageGlob is set")
	}
}

func TestNewFilesystemFileTemplateRendererNoTemplates(t *testing.T) {
	dir := t.TempDir()
	layout := filepath.Join(dir, "layout.html")
	if err := os.WriteFile(layout, []byte(`{{define "layout"}}x{{end}}`), 0o644); err != nil {
		t.Fatal(err)
	}
	// Glob only matches the layout itself, which is skipped -> no page templates.
	if _, err := NewFileTemplateRenderer(layout, filepath.Join(dir, "*.html")); err == nil {
		t.Fatal("expected error when no page templates are found")
	}
}

func TestNewFilesystemFileTemplateRendererParsesPages(t *testing.T) {
	dir := t.TempDir()
	layout := filepath.Join(dir, "layout.html")
	page := filepath.Join(dir, "markdown.html")
	if err := os.WriteFile(layout, []byte(`{{define "layout"}}<h1>{{.Title}}</h1>{{template "content" .}}{{end}}`), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(page, []byte(`{{define "content"}}<p>page</p>{{end}}`), 0o644); err != nil {
		t.Fatal(err)
	}
	r, err := NewFileTemplateRenderer(layout, filepath.Join(dir, "*.html"))
	if err != nil {
		t.Fatalf("renderer: %v", err)
	}
	var sb strings.Builder
	if err := r.RenderTo(&sb, "markdown", PageData{Title: "Hi"}); err != nil {
		t.Fatalf("render: %v", err)
	}
	if !strings.Contains(sb.String(), "<h1>Hi</h1>") {
		t.Fatalf("output = %q", sb.String())
	}
}

func TestRenderToUnknownTemplate(t *testing.T) {
	r, err := NewFileTemplateRenderer("", "")
	if err != nil {
		t.Fatalf("renderer: %v", err)
	}
	var sb strings.Builder
	if err := r.RenderTo(&sb, "does-not-exist", PageData{}); err == nil {
		t.Fatal("expected error for unknown template")
	}
}
