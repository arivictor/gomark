package gomark

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}

func TestLoadConfigFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "gomark.yaml")
	writeFile(t, path, `
title: My Docs
url: https://docs.example.com
lang: fr
theme_color: "#0070f3"
footer: "© 2026 Example"
logo:
  light: /logo-light.png
  dark: /logo-dark.png
seo:
  description: Short default description.
  og_image: /og.png
  twitter_image: /tw.png
  twitter_site: "@example"
  twitter_creator: "@author"
  image_alt: My Docs Cover
build:
  content_dir: content
  output_dir: dist
  sidebar_depth: 3
  runner: false
  sitemap: false
  robots: true
nav:
  - label: Home
    url: /
  - label: GitHub
    url: https://github.com/example
social:
  - label: Twitter
    url: https://twitter.com/example
    icon: twitter
analytics:
  provider: plausible
  id: docs.example.com
`)

	cfg, err := LoadConfigFile(path)
	if err != nil {
		t.Fatalf("load config: %v", err)
	}

	if cfg.Title != "My Docs" || cfg.Lang != "fr" || cfg.URL != "https://docs.example.com" {
		t.Fatalf("unexpected top-level fields: %+v", cfg)
	}
	if cfg.Logo.Light != "/logo-light.png" || cfg.Logo.Dark != "/logo-dark.png" {
		t.Fatalf("unexpected logo: %+v", cfg.Logo)
	}
	if cfg.SEO.TwitterSite != "@example" || cfg.SEO.ImageAlt != "My Docs Cover" {
		t.Fatalf("unexpected seo: %+v", cfg.SEO)
	}
	if cfg.Build.SidebarDepth != 3 || cfg.Build.Runner == nil || *cfg.Build.Runner {
		t.Fatalf("unexpected build: %+v", cfg.Build)
	}
	if cfg.Build.Sitemap == nil || *cfg.Build.Sitemap {
		t.Fatalf("expected sitemap=false, got %+v", cfg.Build.Sitemap)
	}
	if len(cfg.Nav) != 2 || cfg.Nav[1].URL != "https://github.com/example" {
		t.Fatalf("unexpected nav: %+v", cfg.Nav)
	}
	if len(cfg.Social) != 1 || cfg.Social[0].Icon != "twitter" {
		t.Fatalf("unexpected social: %+v", cfg.Social)
	}
	if cfg.Analytics.Provider != "plausible" || cfg.Analytics.ID != "docs.example.com" {
		t.Fatalf("unexpected analytics: %+v", cfg.Analytics)
	}
}

func TestDiscoverConfigFile(t *testing.T) {
	dir := t.TempDir()
	if got := DiscoverConfigFile(dir); got != "" {
		t.Fatalf("expected no config, got %q", got)
	}

	ymlPath := filepath.Join(dir, "gomark.yml")
	writeFile(t, ymlPath, "title: T\n")
	if got := DiscoverConfigFile(dir); got != ymlPath {
		t.Fatalf("expected %q, got %q", ymlPath, got)
	}

	// gomark.yaml wins over gomark.yml when both exist.
	yamlPath := filepath.Join(dir, "gomark.yaml")
	writeFile(t, yamlPath, "title: T\n")
	if got := DiscoverConfigFile(dir); got != yamlPath {
		t.Fatalf("expected %q to win, got %q", yamlPath, got)
	}
}

func TestFileConfigOptionsApply(t *testing.T) {
	runner := false
	cfg := &FileConfig{
		Title:      "Docs",
		Lang:       "de",
		ThemeColor: "#111",
		Logo:       LogoConfig{Light: "/l.png", Dark: "/d.png"},
		SEO:        SEOConfig{TwitterSite: "@x", ImageAlt: "Alt"},
		Build:      BuildConfig{SidebarDepth: 4, Runner: &runner},
		Nav:        []ConfigLink{{Label: "Home", URL: "/"}},
		Analytics:  AnalyticsConfig{Provider: "GA4", ID: "G-123"},
	}

	s := NewSite(cfg.Options()...)
	a := s.App
	if a.Title != "Docs" || a.Lang != "de" || a.ThemeColor != "#111" {
		t.Fatalf("identity not applied: %+v", a)
	}
	if a.LogoLight != "/l.png" || a.LogoDark != "/d.png" {
		t.Fatalf("logos not applied: %+v", a)
	}
	if a.TwitterSite != "@x" || a.imageAlt() != "Alt" {
		t.Fatalf("seo not applied: %+v", a)
	}
	if a.SidebarDepth != 4 || !a.DisableRunner {
		t.Fatalf("build opts not applied: %+v", a)
	}
	if len(a.NavLinks) != 1 || a.NavLinks[0].Label != "Home" {
		t.Fatalf("nav not applied: %+v", a.NavLinks)
	}
	// Provider is normalized to lower-case.
	if a.Analytics.Provider != "ga4" || a.Analytics.ID != "G-123" {
		t.Fatalf("analytics not applied: %+v", a.Analytics)
	}
}

func TestExportWithConfigRendersMetadataAndDropsSitemap(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "content", "index.md"), "# Home\n\nWelcome.\n")

	noSitemap := false
	cfg := &FileConfig{
		Title: "Acme Docs",
		URL:   "https://docs.acme.test",
		SEO: SEOConfig{
			Description: "Acme product documentation.",
			TwitterSite: "@acme",
			ImageAlt:    "Acme",
		},
		Nav:       []ConfigLink{{Label: "GitHub", URL: "https://github.com/acme"}},
		Social:    []ConfigLink{{Label: "X", URL: "https://x.com/acme", Icon: "twitter"}},
		Analytics: AnalyticsConfig{Provider: "plausible", ID: "docs.acme.test"},
		Build:     BuildConfig{ContentDir: filepath.Join(dir, "content"), Sitemap: &noSitemap},
	}

	out := filepath.Join(dir, "dist")
	if err := NewSite(cfg.Options()...).Export(out); err != nil {
		t.Fatalf("export: %v", err)
	}

	html := readFileString(t, filepath.Join(out, "index.html"))
	wants := []string{
		`<meta name="description" content="Acme product documentation." />`,
		`<meta name="twitter:site" content="@acme" />`,
		`data-domain="docs.acme.test"`,
		`href="https://github.com/acme"`,
		`href="https://x.com/acme"`,
	}
	for _, want := range wants {
		if !strings.Contains(html, want) {
			t.Fatalf("rendered html missing %q\n---\n%s", want, html)
		}
	}

	if _, err := os.Stat(filepath.Join(out, "sitemap.xml")); !os.IsNotExist(err) {
		t.Fatalf("expected no sitemap.xml when disabled, stat err = %v", err)
	}
	if _, err := os.Stat(filepath.Join(out, "robots.txt")); err != nil {
		t.Fatalf("expected robots.txt to be written, got %v", err)
	}
}

func TestAnalyticsRequiresIDToRender(t *testing.T) {
	// A provider without an id is treated as unset: no option, no script tag.
	cfg := &FileConfig{Analytics: AnalyticsConfig{Provider: "ga4"}}
	s := NewSite(cfg.Options()...)
	if s.App.Analytics.Provider != "" {
		t.Fatalf("expected analytics dropped when id is empty, got %+v", s.App.Analytics)
	}

	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "content", "index.md"), "# Home\n")
	out := filepath.Join(dir, "dist")
	// Even if a provider somehow reaches the App without an id, the template must
	// not emit a broken analytics tag.
	site := NewSite(
		WithSiteContentDir(filepath.Join(dir, "content")),
		WithSiteAnalytics("ga4", ""),
	)
	if err := site.Export(out); err != nil {
		t.Fatalf("export: %v", err)
	}
	html := readFileString(t, filepath.Join(out, "index.html"))
	if strings.Contains(html, "googletagmanager.com/gtag/js") {
		t.Fatalf("expected no analytics script when id is empty\n---\n%s", html)
	}
}

func readFileString(t *testing.T, path string) string {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	return string(data)
}
