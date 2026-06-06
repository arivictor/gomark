package gomark

import (
	"strings"
	"testing"
	"time"
)

func TestBuildSitemapRoutesNil(t *testing.T) {
	if buildSitemapRoutes(nil) != nil {
		t.Fatal("expected nil routes for nil index")
	}
}

func TestRenderSitemapXML(t *testing.T) {
	xml, err := renderSitemapXML("https://example.com", []string{"/", "/about"}, time.Date(2026, 1, 2, 0, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatalf("render: %v", err)
	}
	if !strings.Contains(xml, "<loc>https://example.com/about</loc>") {
		t.Fatalf("expected about loc: %s", xml)
	}
	if !strings.Contains(xml, "<lastmod>2026-01-02</lastmod>") {
		t.Fatalf("expected lastmod: %s", xml)
	}
	if !strings.HasPrefix(xml, "<?xml") {
		t.Fatalf("expected xml header: %s", xml)
	}
}

func TestRenderRobotsTXT(t *testing.T) {
	txt := renderRobotsTXT("https://example.com")
	if !strings.Contains(txt, "User-agent: *") {
		t.Fatalf("expected user-agent: %s", txt)
	}
	if !strings.Contains(txt, "Sitemap: https://example.com/sitemap.xml") {
		t.Fatalf("expected sitemap line: %s", txt)
	}
}

func TestPageTitleFromSlugEdgeCases(t *testing.T) {
	// All-separator slug collapses to the fallback.
	if got := pageTitleFromSlug("---"); got != "Content" {
		t.Fatalf("got %q", got)
	}
	// Consecutive separators produce no empty words.
	if got := pageTitleFromSlug("a--b"); got != "A B" {
		t.Fatalf("got %q", got)
	}
	// Trailing-empty segment falls back.
	if got := pageTitleFromSlug("guides/"); got != "Content" {
		t.Fatalf("got %q", got)
	}
}
