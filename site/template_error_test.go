package site

import (
	"html/template"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestNewFileTemplateRendererUsesEmbeddedTemplatesByDefault(t *testing.T) {
	renderer, err := NewFileTemplateRenderer("", "")
	if err != nil {
		t.Fatalf("new embedded template renderer: %v", err)
	}

	rec := httptest.NewRecorder()
	if err := renderer.Render(rec, "markdown", PageData{Title: "Home", BodyHTML: template.HTML("<p>content</p>")}); err != nil {
		t.Fatalf("render markdown with embedded templates: %v", err)
	}

	html := rec.Body.String()
	if !strings.Contains(html, `<title>Home`) {
		t.Fatalf("expected embedded layout to render title, got html: %s", html)
	}
}

func TestLayoutAddsErrorLayoutClassForErrorPages(t *testing.T) {
	renderer, err := NewFileTemplateRenderer("templates/layout.html", "templates/*.html")
	if err != nil {
		t.Fatalf("new file template renderer: %v", err)
	}

	rec := httptest.NewRecorder()
	data := PageData{
		StatusCode:  404,
		Title:       "Page not found",
		Description: "The page does not exist.",
	}
	if err := renderer.RenderStatus(rec, 404, "error", data); err != nil {
		t.Fatalf("render error template: %v", err)
	}

	html := rec.Body.String()
	if !strings.Contains(html, `<div class="docs docs--error">`) {
		t.Fatalf("expected error docs layout class, got html: %s", html)
	}
	if !strings.Contains(html, `<section class="errorpage">`) {
		t.Fatalf("expected error page section in html: %s", html)
	}
}

func TestLayoutRendersCustomSiteTitleAndLogo(t *testing.T) {
	renderer, err := NewFileTemplateRenderer("", "")
	if err != nil {
		t.Fatalf("new embedded template renderer: %v", err)
	}

	rec := httptest.NewRecorder()
	data := PageData{
		Title:       "Welcome",
		SiteName:    "My Docs",
		LogoURL:     "https://example.com/logo.svg",
		BodyHTML:    template.HTML("<p>content</p>"),
		Nav:         nil,
		TopNav:      nil,
		CurrentPath: "/",
	}
	if err := renderer.Render(rec, "markdown", data); err != nil {
		t.Fatalf("render markdown template: %v", err)
	}

	html := rec.Body.String()
	if !strings.Contains(html, `aria-label="My Docs home"`) {
		t.Fatalf("expected custom site title in header aria-label, got html: %s", html)
	}
	if !strings.Contains(html, `src="https://example.com/logo.svg"`) {
		t.Fatalf("expected custom logo url in header, got html: %s", html)
	}
	if !strings.Contains(html, `<span>My Docs</span>`) {
		t.Fatalf("expected custom site title text in brand, got html: %s", html)
	}
}
