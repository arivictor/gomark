package site

import (
	"html/template"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestLayoutRendersAccordionControlsForFolders(t *testing.T) {
	renderer, err := NewFileTemplateRenderer("templates/layout.html", "templates/*.html")
	if err != nil {
		t.Fatalf("new file template renderer: %v", err)
	}

	data := PageData{
		Title:           "Nav Test",
		BodyHTML:        template.HTML("<p>content</p>"),
		CurrentPath:     "/",
		OGImageURL:      "https://localhost:8080/gomark-og-1200x630.png",
		TwitterImageURL: "https://localhost:8080/gomark-twitter-1200x628.png",
		Nav: []NavNode{
			{
				Title:  "Routable Folder",
				Path:   "/routable",
				NodeID: "nav-routable",
				Folder: true,
				Open:   true,
				Children: []NavNode{
					{Title: "Leaf", Path: "/routable/leaf"},
				},
			},
			{
				Title:  "Toggle Folder",
				NodeID: "nav-toggle-only",
				Folder: true,
				Children: []NavNode{
					{Title: "Leaf 2", Path: "/toggle/leaf"},
				},
			},
		},
	}

	rec := httptest.NewRecorder()
	if err := renderer.Render(rec, "markdown", data); err != nil {
		t.Fatalf("render markdown template: %v", err)
	}

	html := rec.Body.String()
	if !strings.Contains(html, "href=\"/routable\"") || !strings.Contains(html, "<span class=\"label\">Routable Folder</span>") {
		t.Fatalf("expected routable folder label to render as link: %s", html)
	}
	if strings.Contains(html, "href=\"\">") {
		t.Fatalf("expected toggle-only folder to avoid empty href links: %s", html)
	}
	if !strings.Contains(html, "<button type=\"button\" class=\"navitem folder has-children\"") || !strings.Contains(html, "<span class=\"label\">Toggle Folder</span>") {
		t.Fatalf("expected toggle-only folder label to render as button control: %s", html)
	}
	if !strings.Contains(html, "data-nav-folder") {
		t.Fatalf("expected folder controls to be attached to folder labels: %s", html)
	}
	if !strings.Contains(html, "aria-controls=\"nav-routable\"") {
		t.Fatalf("expected aria-controls wiring for accordion group: %s", html)
	}
	if strings.Contains(html, "class=\"navtoggle\"") {
		t.Fatalf("did not expect separate nav toggle button control: %s", html)
	}
	if !strings.Contains(html, "data-current-path=\"/\"") {
		t.Fatalf("expected current route marker on body for root toggle logic: %s", html)
	}
	if !strings.Contains(html, `<meta name="twitter:image" content="https://localhost:8080/gomark-twitter-1200x628.png" />`) {
		t.Fatalf("expected twitter image meta tag to use square image: %s", html)
	}
	if !strings.Contains(html, `<meta name="twitter:card" content="summary_large_image" />`) {
		t.Fatalf("expected twitter card to be summary_large_image when twitter image is set: %s", html)
	}
	if !strings.Contains(html, `<link rel="manifest" href="/site.webmanifest" />`) {
		t.Fatalf("expected site manifest link tag: %s", html)
	}
}
