package gomark

import (
	"testing"
)

func TestSiteOptionsSetAppFields(t *testing.T) {
	s := NewSite(
		WithSiteAddress("  :9000  "),
		WithSiteTitle("  My Site  "),
		WithSiteLogo("  /logo.svg  "),
		WithSiteLogoLight("/light.svg"),
		WithSiteLogoDark("/dark.svg"),
		WithSiteLang("  fr  "),
		WithSiteThemeColor("  #123456  "),
		WithSiteDescription("  hello  "),
		WithSiteFooter("  footer text  "),
		WithSiteOGImage("  /og.png  "),
		WithSiteTwitterImage("  /tw.png  "),
		WithSiteTwitterSite("  @site  "),
		WithSiteTwitterCreator("  @creator  "),
		WithSiteImageAlt("  alt  "),
		WithSiteContentDir("  docs  "),
		WithSitePublicDir("  public  "),
		WithSiteSidebarDepth(4),
		WithSiteURL("  https://example.com  "),
		WithSiteMode(PreRender),
		WithSiteExportDir("  dist  "),
	)

	if s.addr != ":9000" {
		t.Errorf("addr = %q", s.addr)
	}
	if s.App.Title != "My Site" {
		t.Errorf("title = %q", s.App.Title)
	}
	// WithSiteLogo set both, then LogoLight/Dark overrode.
	if s.App.LogoLight != "/light.svg" || s.App.LogoDark != "/dark.svg" {
		t.Errorf("logos = %q/%q", s.App.LogoLight, s.App.LogoDark)
	}
	if s.App.Lang != "fr" {
		t.Errorf("lang = %q", s.App.Lang)
	}
	if s.App.ThemeColor != "#123456" {
		t.Errorf("themeColor = %q", s.App.ThemeColor)
	}
	if s.App.Description != "hello" {
		t.Errorf("description = %q", s.App.Description)
	}
	if s.App.Footer != "footer text" {
		t.Errorf("footer = %q", s.App.Footer)
	}
	if s.App.OGImagePath != "/og.png" {
		t.Errorf("og = %q", s.App.OGImagePath)
	}
	if s.App.TwitterImagePath != "/tw.png" {
		t.Errorf("tw = %q", s.App.TwitterImagePath)
	}
	if s.App.TwitterSite != "@site" {
		t.Errorf("twitterSite = %q", s.App.TwitterSite)
	}
	if s.App.TwitterCreator != "@creator" {
		t.Errorf("twitterCreator = %q", s.App.TwitterCreator)
	}
	if s.App.ImageAlt != "alt" {
		t.Errorf("imageAlt = %q", s.App.ImageAlt)
	}
	if s.App.ContentDir != "docs" {
		t.Errorf("contentDir = %q", s.App.ContentDir)
	}
	if s.App.PublicDir != "public" {
		t.Errorf("publicDir = %q", s.App.PublicDir)
	}
	if s.App.SidebarDepth != 4 {
		t.Errorf("sidebarDepth = %d", s.App.SidebarDepth)
	}
	if s.App.SiteURL != "https://example.com" {
		t.Errorf("siteURL = %q", s.App.SiteURL)
	}
	if s.App.Mode != PreRender {
		t.Errorf("mode = %q", s.App.Mode)
	}
	if s.App.ExportDir != "dist" {
		t.Errorf("exportDir = %q", s.App.ExportDir)
	}
}

func TestWithSiteLogoSetsBoth(t *testing.T) {
	s := NewSite(WithSiteLogo(" /brand.svg "))
	if s.App.LogoLight != "/brand.svg" || s.App.LogoDark != "/brand.svg" {
		t.Fatalf("logo = %q/%q", s.App.LogoLight, s.App.LogoDark)
	}
}

func TestWithSiteNavAndSocialLinks(t *testing.T) {
	nav := []ConfigLink{{Label: "Docs", URL: "/docs"}}
	social := []ConfigLink{{Label: "GitHub", URL: "https://github.com", Icon: "github"}}
	s := NewSite(WithSiteNavLinks(nav...), WithSiteSocialLinks(social...))
	if len(s.App.NavLinks) != 1 || s.App.NavLinks[0].Label != "Docs" {
		t.Errorf("navLinks = %+v", s.App.NavLinks)
	}
	if len(s.App.SocialLinks) != 1 || s.App.SocialLinks[0].Icon != "github" {
		t.Errorf("socialLinks = %+v", s.App.SocialLinks)
	}
}

func TestWithSiteAnalyticsNormalizes(t *testing.T) {
	s := NewSite(WithSiteAnalytics("  GA4  ", "  G-123  "))
	if s.App.Analytics.Provider != "ga4" {
		t.Errorf("provider = %q", s.App.Analytics.Provider)
	}
	if s.App.Analytics.ID != "G-123" {
		t.Errorf("id = %q", s.App.Analytics.ID)
	}
}

func TestWithSiteToggleOptions(t *testing.T) {
	s := NewSite(
		WithSiteSitemapEnabled(false),
		WithSiteRobotsEnabled(false),
		WithSiteRunnerEnabled(false),
	)
	if !s.App.DisableSitemap {
		t.Error("expected sitemap disabled")
	}
	if !s.App.DisableRobots {
		t.Error("expected robots disabled")
	}
	if !s.App.DisableRunner {
		t.Error("expected runner disabled")
	}

	on := NewSite(
		WithSiteSitemapEnabled(true),
		WithSiteRobotsEnabled(true),
		WithSiteRunnerEnabled(true),
	)
	if on.App.DisableSitemap || on.App.DisableRobots || on.App.DisableRunner {
		t.Error("expected all enabled")
	}
	if !on.App.sitemapEnabled() || !on.App.robotsEnabled() {
		t.Error("expected sitemapEnabled/robotsEnabled true")
	}
}

func TestNewSiteSkipsNilOptions(t *testing.T) {
	s := NewSite(nil, WithSiteTitle("X"), nil)
	if s.App.Title != "X" {
		t.Fatalf("title = %q", s.App.Title)
	}
}

func TestExportDirFallsBackToEnv(t *testing.T) {
	a := &App{}
	if got := a.exportDir(); got != "" {
		t.Fatalf("expected empty exportDir, got %q", got)
	}
	t.Setenv("EXPORT_DIR", "  /tmp/out  ")
	if got := a.exportDir(); got != "/tmp/out" {
		t.Fatalf("env exportDir = %q", got)
	}
	a.ExportDir = "explicit"
	if got := a.exportDir(); got != "explicit" {
		t.Fatalf("explicit exportDir = %q", got)
	}
}

func TestGetRunnerEnabledEnvOverrides(t *testing.T) {
	cases := []struct {
		env      string
		disable  bool
		expected bool
	}{
		{"", false, true},
		{"", true, false},
		{"0", false, false},
		{"false", false, false},
		{"no", false, false},
		{"off", false, false},
		{"1", true, true},
		{"true", true, true},
		{"yes", true, true},
		{"on", true, true},
		{"garbage", true, false}, // unknown falls back to !DisableRunner
		{"garbage", false, true},
	}
	for _, tc := range cases {
		t.Setenv("PLAYGROUND_ENABLED", tc.env)
		a := &App{DisableRunner: tc.disable}
		if got := a.GetRunnerEnabled(); got != tc.expected {
			t.Errorf("env=%q disable=%v: got %v, want %v", tc.env, tc.disable, got, tc.expected)
		}
	}
}

func TestParseRenderMode(t *testing.T) {
	pre := []string{"prerender", "pre-render", "pre_render", "pre", "prod", "production", "  PROD  "}
	for _, raw := range pre {
		if ParseRenderMode(raw) != PreRender {
			t.Errorf("ParseRenderMode(%q) != PreRender", raw)
		}
	}
	live := []string{"liverender", "live-render", "live_render", "live", "dev", "development", "", "weird"}
	for _, raw := range live {
		if ParseRenderMode(raw) != LiveRender {
			t.Errorf("ParseRenderMode(%q) != LiveRender", raw)
		}
	}
}

func TestAppModeFromEnvChain(t *testing.T) {
	// Explicit Mode wins.
	a := &App{Mode: PreRender}
	if a.mode() != PreRender {
		t.Fatalf("explicit mode")
	}

	// Clear all, then exercise each env fallback in priority order.
	for _, k := range []string{"APP_MODE", "APP_ENV", "GO_ENV", "ENV"} {
		t.Setenv(k, "")
	}
	b := &App{}
	if b.mode() != LiveRender {
		t.Fatalf("empty env should default to LiveRender")
	}

	t.Setenv("ENV", "prod")
	if b.mode() != PreRender {
		t.Fatalf("ENV=prod should be PreRender")
	}
	t.Setenv("GO_ENV", "dev")
	if b.mode() != LiveRender {
		t.Fatalf("GO_ENV should take priority over ENV")
	}
	t.Setenv("APP_ENV", "prod")
	if b.mode() != PreRender {
		t.Fatalf("APP_ENV should take priority over GO_ENV")
	}
	t.Setenv("APP_MODE", "live")
	if b.mode() != LiveRender {
		t.Fatalf("APP_MODE should take highest priority")
	}
}

func TestAppContentDirDefaultsAndCleans(t *testing.T) {
	if (&App{}).contentDir() != "content" {
		t.Error("expected default content dir")
	}
	if got := (&App{ContentDir: "docs/./pages/"}).contentDir(); got != "docs/pages" {
		t.Errorf("contentDir = %q", got)
	}
}

func TestAppSidebarDepthDefault(t *testing.T) {
	if (&App{}).sidebarDepth() != 2 {
		t.Error("expected default sidebar depth 2")
	}
	if (&App{SidebarDepth: -5}).sidebarDepth() != 2 {
		t.Error("expected non-positive depth to default to 2")
	}
	if (&App{SidebarDepth: 5}).sidebarDepth() != 5 {
		t.Error("expected configured depth")
	}
}

func TestAppSiteURL(t *testing.T) {
	t.Setenv("SITE_URL", "")
	if got := (&App{SiteURL: "example.com"}).siteURL(); got != "https://example.com" {
		t.Errorf("siteURL = %q", got)
	}
	t.Setenv("SITE_URL", "env.example.com")
	if got := (&App{}).siteURL(); got != "https://env.example.com" {
		t.Errorf("env siteURL = %q", got)
	}
	t.Setenv("SITE_URL", "")
	if got := (&App{}).siteURL(); got != defaultSiteURL {
		t.Errorf("default siteURL = %q", got)
	}
}

func TestAppSiteTitleDefault(t *testing.T) {
	if (&App{}).siteTitle() != defaultSiteName {
		t.Error("expected default site name")
	}
	if (&App{Title: "  Custom  "}).siteTitle() != "Custom" {
		t.Error("expected trimmed title")
	}
}

func TestAppLangDefault(t *testing.T) {
	if (&App{}).lang() != "en" {
		t.Error("expected default lang en")
	}
	if (&App{Lang: " de "}).lang() != "de" {
		t.Error("expected configured lang")
	}
}

func TestAppImageAltDefaultsToTitle(t *testing.T) {
	if (&App{Title: "Foo"}).imageAlt() != "Foo" {
		t.Error("expected imageAlt to default to title")
	}
	if (&App{ImageAlt: " Bar "}).imageAlt() != "Bar" {
		t.Error("expected configured imageAlt")
	}
}

func TestAppOGAndTwitterImageDefaults(t *testing.T) {
	if (&App{}).ogImagePath() != defaultOGImagePath {
		t.Error("expected default og image")
	}
	if (&App{OGImagePath: " /custom-og.png "}).ogImagePath() != "/custom-og.png" {
		t.Error("expected configured og image")
	}
	if (&App{}).twitterImagePath() != defaultTwitterImagePath {
		t.Error("expected default twitter image")
	}
	if (&App{TwitterImagePath: " /custom-tw.png "}).twitterImagePath() != "/custom-tw.png" {
		t.Error("expected configured twitter image")
	}
}

func TestAppLogoAndPublicDirHelpers(t *testing.T) {
	a := &App{LogoLight: " /l.svg ", LogoDark: " /d.svg "}
	if a.logoLight() != "/l.svg" || a.logoDark() != "/d.svg" {
		t.Error("expected trimmed logos")
	}
	t.Setenv("PUBLIC_DIR", "")
	if (&App{PublicDir: " disk "}).publicDir() != "disk" {
		t.Error("expected configured public dir")
	}
	t.Setenv("PUBLIC_DIR", " envdir ")
	if (&App{}).publicDir() != "envdir" {
		t.Error("expected env public dir")
	}
}
