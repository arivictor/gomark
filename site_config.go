package gomark

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/arivictor/gomark/protocol"
)

type App struct {
	Title           string
	Logo            string
	ContentDir      string
	TemplatesDir    string
	LayoutPath      string
	TemplateGlob    string
	PublicDir       string
	SidebarDepth    int
	SiteURL         string
	Mode            RenderMode
	RunnerEnabled   bool
	RunnerURL       string
	RunnerAuthMode  protocol.AuthMode
	RunnerAuthToken string
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

func WithSiteRunnerEnabled(enabled bool) SiteOption {
	return func(s *Site) {
		s.App.RunnerEnabled = enabled
	}
}

func WithSiteRunnerURL(url string) SiteOption {
	return func(s *Site) {
		s.App.RunnerURL = strings.TrimSpace(url)
	}
}

func WithSiteRunnerAuth(mode protocol.AuthMode, token string) SiteOption {
	return func(s *Site) {
		s.App.RunnerAuthMode = mode
		s.App.RunnerAuthToken = strings.TrimSpace(token)
	}
}

func WithSiteRunner(url string, mode protocol.AuthMode, token string) SiteOption {
	return func(s *Site) {
		s.App.RunnerEnabled = true
		s.App.RunnerURL = strings.TrimSpace(url)
		s.App.RunnerAuthMode = mode
		s.App.RunnerAuthToken = strings.TrimSpace(token)
	}
}

func (a *App) GetRunnerEnabled() bool {
	if a.RunnerEnabled {
		return true
	}
	raw := strings.ToLower(strings.TrimSpace(os.Getenv("PLAYGROUND_ENABLED")))
	return raw == "1" || raw == "true" || raw == "yes" || raw == "on"
}

func (a *App) GetRunnerURL() string {
	if strings.TrimSpace(a.RunnerURL) != "" {
		return strings.TrimSpace(a.RunnerURL)
	}
	return strings.TrimSpace(os.Getenv("PLAYGROUND_RUNNER_URL"))
}

func (a *App) GetRunnerAuthMode() protocol.AuthMode {
	if strings.TrimSpace(string(a.RunnerAuthMode)) != "" {
		return protocol.AuthMode(strings.ToLower(strings.TrimSpace(string(a.RunnerAuthMode))))
	}
	return protocol.AuthMode(strings.ToLower(strings.TrimSpace(os.Getenv("PLAYGROUND_RUNNER_AUTH_MODE"))))
}

func (a *App) GetRunnerAuthToken() string {
	if strings.TrimSpace(a.RunnerAuthToken) != "" {
		return strings.TrimSpace(a.RunnerAuthToken)
	}
	return strings.TrimSpace(os.Getenv("PLAYGROUND_RUNNER_AUTH_TOKEN"))
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
