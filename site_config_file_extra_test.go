package gomark

import (
	"testing"
)

func TestFileConfigOptionsAllFields(t *testing.T) {
	runner, sitemap, robots := true, false, false
	cfg := &FileConfig{
		Title:      "T",
		URL:        "https://x.test",
		Lang:       "es",
		ThemeColor: "#abc",
		Footer:     "footer",
		Logo:       LogoConfig{Light: "/l.png", Dark: "/d.png"},
		SEO: SEOConfig{
			Description:    "desc",
			OGImage:        "/og.png",
			TwitterImage:   "/tw.png",
			TwitterSite:    "@s",
			TwitterCreator: "@c",
			ImageAlt:       "alt",
		},
		Build: BuildConfig{
			ContentDir:   "content",
			OutputDir:    "dist",
			PublicDir:    "public",
			SidebarDepth: 3,
			Runner:       &runner,
			Sitemap:      &sitemap,
			Robots:       &robots,
		},
		Nav:       []ConfigLink{{Label: "Home", URL: "/"}},
		Social:    []ConfigLink{{Label: "X", URL: "https://x.com", Icon: "x"}},
		Analytics: AnalyticsConfig{Provider: "ga4", ID: "G-1"},
	}

	s := NewSite(cfg.Options()...)
	a := s.App
	checks := map[string]bool{
		"title":           a.Title == "T",
		"url":             a.SiteURL == "https://x.test",
		"lang":            a.Lang == "es",
		"themeColor":      a.ThemeColor == "#abc",
		"footer":          a.Footer == "footer",
		"logoLight":       a.LogoLight == "/l.png",
		"logoDark":        a.LogoDark == "/d.png",
		"description":     a.Description == "desc",
		"ogImage":         a.OGImagePath == "/og.png",
		"twitterImage":    a.TwitterImagePath == "/tw.png",
		"twitterSite":     a.TwitterSite == "@s",
		"twitterCreator":  a.TwitterCreator == "@c",
		"imageAlt":        a.ImageAlt == "alt",
		"contentDir":      a.ContentDir == "content",
		"outputDir":       a.ExportDir == "dist",
		"publicDir":       a.PublicDir == "public",
		"sidebarDepth":    a.SidebarDepth == 3,
		"runnerEnabled":   !a.DisableRunner,
		"sitemapDisabled": a.DisableSitemap,
		"robotsDisabled":  a.DisableRobots,
		"nav":             len(a.NavLinks) == 1,
		"social":          len(a.SocialLinks) == 1,
		"analytics":       a.Analytics.ID == "G-1",
	}
	for name, ok := range checks {
		if !ok {
			t.Errorf("option %s not applied correctly: %+v", name, a)
		}
	}
}

func TestFileConfigOptionsNil(t *testing.T) {
	var c *FileConfig
	if c.Options() != nil {
		t.Fatal("nil config -> nil options")
	}
}

func TestLoadConfigFileMissing(t *testing.T) {
	if _, err := LoadConfigFile("/nonexistent/path/gomark.yaml"); err == nil {
		t.Fatal("expected error for missing config file")
	}
}
