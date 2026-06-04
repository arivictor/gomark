package gomark

import (
	"os"
	"path/filepath"
	"strings"
)

type App struct {
	Title            string
	Logo             string
	OGImagePath      string
	TwitterImagePath string
	ContentDir       string
	TemplatesDir     string
	LayoutPath       string
	TemplateGlob     string
	PublicDir        string
	SidebarDepth     int
	SiteURL          string
	Mode             RenderMode
	// DisableRunner turns off the in-browser Go runner. Execution is
	// client-side (a WebAssembly build of the yaegi interpreter), so the
	// runner is on by default and needs no external service; set this to hide
	// the "Run" controls entirely.
	DisableRunner bool
}

type RenderMode string

const (
	LiveRender RenderMode = "live_render"
	PreRender  RenderMode = "pre_render"
)

func WithSiteAddress(addr string) SiteOption {
	return func(s *Site) {
		s.addr = strings.TrimSpace(addr)
	}
}

func WithSiteTitle(title string) SiteOption {
	return func(s *Site) {
		s.App.Title = strings.TrimSpace(title)
	}
}

func WithSiteLogo(logoURL string) SiteOption {
	return func(s *Site) {
		s.App.Logo = strings.TrimSpace(logoURL)
	}
}

func WithSiteOGImage(imagePath string) SiteOption {
	return func(s *Site) {
		s.App.OGImagePath = strings.TrimSpace(imagePath)
	}
}

func WithSiteTwitterImage(imagePath string) SiteOption {
	return func(s *Site) {
		s.App.TwitterImagePath = strings.TrimSpace(imagePath)
	}
}

func WithSiteContentDir(dir string) SiteOption {
	return func(s *Site) {
		s.App.ContentDir = strings.TrimSpace(dir)
	}
}

func WithSiteTemplatesDir(dir string) SiteOption {
	return func(s *Site) {
		s.App.TemplatesDir = strings.TrimSpace(dir)
	}
}

func WithSiteLayoutPath(path string) SiteOption {
	return func(s *Site) {
		s.App.LayoutPath = strings.TrimSpace(path)
	}
}

func WithSiteTemplateGlob(glob string) SiteOption {
	return func(s *Site) {
		s.App.TemplateGlob = strings.TrimSpace(glob)
	}
}

func WithSitePublicDir(dir string) SiteOption {
	return func(s *Site) {
		s.App.PublicDir = strings.TrimSpace(dir)
	}
}

func WithSiteSidebarDepth(depth int) SiteOption {
	return func(s *Site) {
		s.App.SidebarDepth = depth
	}
}

func WithSiteURL(siteURL string) SiteOption {
	return func(s *Site) {
		s.App.SiteURL = strings.TrimSpace(siteURL)
	}
}

func WithSiteMode(mode RenderMode) SiteOption {
	return func(s *Site) {
		s.App.Mode = mode
	}
}

// WithSiteRunnerEnabled toggles the in-browser Go runner. It is enabled by
// default; pass false to hide the "Run" controls across the site.
func WithSiteRunnerEnabled(enabled bool) SiteOption {
	return func(s *Site) {
		s.App.DisableRunner = !enabled
	}
}

// GetRunnerEnabled reports whether the client-side runner should be active.
// It defaults to on (execution is client-side, so there is nothing to provision)
// and can be turned off via WithSiteRunnerEnabled(false) or PLAYGROUND_ENABLED.
func (a *App) GetRunnerEnabled() bool {
	switch strings.ToLower(strings.TrimSpace(os.Getenv("PLAYGROUND_ENABLED"))) {
	case "0", "false", "no", "off":
		return false
	case "1", "true", "yes", "on":
		return true
	}
	return !a.DisableRunner
}

func (a *App) mode() RenderMode {
	raw := strings.TrimSpace(string(a.Mode))
	if raw == "" {
		raw = strings.TrimSpace(os.Getenv("APP_MODE"))
	}
	if raw == "" {
		raw = strings.TrimSpace(os.Getenv("APP_ENV"))
	}
	if raw == "" {
		raw = strings.TrimSpace(os.Getenv("GO_ENV"))
	}
	if raw == "" {
		raw = strings.TrimSpace(os.Getenv("ENV"))
	}

	return ParseRenderMode(raw)
}

func ParseRenderMode(raw string) RenderMode {
	normalized := strings.ToLower(strings.TrimSpace(raw))

	switch normalized {
	case "prerender", "pre-render", "pre_render", "pre", "prod", "production":
		return PreRender
	case "liverender", "live-render", "live_render", "live", "dev", "development":
		return LiveRender
	default:
		return LiveRender
	}
}

func (a *App) contentDir() string {
	if strings.TrimSpace(a.ContentDir) == "" {
		return "content"
	}
	return filepath.Clean(a.ContentDir)
}

func (a *App) sidebarDepth() int {
	if a.SidebarDepth <= 0 {
		return 2
	}
	return a.SidebarDepth
}

func (a *App) templatePaths() (string, string) {
	if dir := strings.TrimSpace(a.TemplatesDir); dir != "" {
		return filepath.Join(dir, "layout.html"), filepath.Join(dir, "*.html")
	}
	return strings.TrimSpace(a.LayoutPath), strings.TrimSpace(a.TemplateGlob)
}

func (a *App) publicDir() string {
	if strings.TrimSpace(a.PublicDir) == "" {
		return ""
	}
	return a.PublicDir
}

func (a *App) siteURL() string {
	if strings.TrimSpace(a.SiteURL) != "" {
		return normalizeBaseURL(a.SiteURL)
	}
	if env := strings.TrimSpace(os.Getenv("SITE_URL")); env != "" {
		return normalizeBaseURL(env)
	}
	return defaultSiteURL
}

func (a *App) siteTitle() string {
	if strings.TrimSpace(a.Title) != "" {
		return strings.TrimSpace(a.Title)
	}
	return defaultSiteName
}

func (a *App) logoURL() string {
	return strings.TrimSpace(a.Logo)
}

func (a *App) ogImagePath() string {
	if strings.TrimSpace(a.OGImagePath) != "" {
		return strings.TrimSpace(a.OGImagePath)
	}
	return defaultOGImagePath
}

func (a *App) twitterImagePath() string {
	if strings.TrimSpace(a.TwitterImagePath) != "" {
		return strings.TrimSpace(a.TwitterImagePath)
	}
	return defaultTwitterImagePath
}
