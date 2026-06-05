package gomark

import (
	"io/fs"
	"log"
	"path/filepath"
	"strings"
	"time"
)

// errorResponder builds the HTML error responder from the shared build, so the
// static server, live server, and error pages all share one identity/SEO config.
func (b *siteBuild) errorResponder(logger *log.Logger) HTMLErrorResponder {
	return HTMLErrorResponder{
		Renderer:         b.renderer,
		TopNav:           b.topNav,
		SiteName:         b.siteName,
		Lang:             b.lang,
		ThemeColor:       b.themeColor,
		LogoLight:        b.logoLight,
		LogoDark:         b.logoDark,
		SiteURL:          b.siteURL,
		OGImagePath:      b.ogImagePath,
		TwitterImagePath: b.twitterImagePath,
		TwitterSite:      b.twitterSite,
		TwitterCreator:   b.twitterCreator,
		ImageAlt:         b.imageAlt,
		Footer:           b.footer,
		NavLinks:         b.navLinks,
		SocialLinks:      b.socialLinks,
		Analytics:        b.analytics,
		Logger:           logger,
	}
}

// siteBuild holds the artifacts that both the live HTTP server (run) and the
// static exporter (Export) need, built once from the App config. Centralizing
// this keeps the served site and the exported site byte-for-byte consistent.
type siteBuild struct {
	renderer         *FileTemplateRenderer
	contentDir       string
	index            *ContentIndex
	provider         contentPageProvider
	searchIndex      *SearchIndex
	topNav           []NavLink
	sitemapRoutes    []string
	sitemapXML       string
	robotsTXT        string
	publicFS         fs.FS
	siteURL          string
	siteName         string
	lang             string
	themeColor       string
	description      string
	logoLight        string
	logoDark         string
	ogImagePath      string
	twitterImagePath string
	twitterSite      string
	twitterCreator   string
	imageAlt         string
	footer           string
	navLinks         []ConfigLink
	socialLinks      []ConfigLink
	analytics        AnalyticsConfig
	runnerEnabled    bool
	sidebarDepth     int
}

// buildSite assembles the shared artifacts. When preRender is true the page
// provider renders every page up front (required for export); otherwise it
// honors the configured render mode.
func (a *App) buildSite(preRender bool) (*siteBuild, error) {
	dir := a.contentDir()

	renderer, err := NewFileTemplateRenderer("", "")
	if err != nil {
		return nil, err
	}

	index, err := BuildContentIndex(dir)
	if err != nil {
		return nil, err
	}

	runnerEnabled := a.GetRunnerEnabled()
	markdownRenderer := StdlibMarkdownRenderer{RunnerEnabled: runnerEnabled}

	var provider contentPageProvider
	if preRender {
		provider, err = newPreRenderedMarkdownProvider(filepath.Clean(dir), markdownRenderer)
	} else {
		provider, err = a.newContentPageProvider(filepath.Clean(dir), markdownRenderer)
	}
	if err != nil {
		return nil, err
	}

	searchIndex, err := BuildSearchIndex(dir)
	if err != nil {
		return nil, err
	}

	siteURL := a.siteURL()
	sitemapRoutes := buildSitemapRoutes(index)

	// sitemap.xml and robots.txt are generated unless disabled; an empty string
	// means "do not serve/write this file" downstream.
	sitemapXML := ""
	if a.sitemapEnabled() {
		sitemapXML, err = renderSitemapXML(siteURL, sitemapRoutes, time.Now())
		if err != nil {
			return nil, err
		}
	}
	robotsTXT := ""
	if a.robotsEnabled() {
		robotsTXT = renderRobotsTXT(siteURL)
	}

	publicFS, err := a.publicFS()
	if err != nil {
		return nil, err
	}

	return &siteBuild{
		renderer:         renderer,
		contentDir:       filepath.Clean(dir),
		index:            index,
		provider:         provider,
		searchIndex:      searchIndex,
		topNav:           index.TopNav(),
		sitemapRoutes:    sitemapRoutes,
		sitemapXML:       sitemapXML,
		robotsTXT:        robotsTXT,
		publicFS:         publicFS,
		siteURL:          siteURL,
		siteName:         a.siteTitle(),
		lang:             a.lang(),
		themeColor:       strings.TrimSpace(a.ThemeColor),
		description:      strings.TrimSpace(a.Description),
		logoLight:        a.logoLight(),
		logoDark:         a.logoDark(),
		ogImagePath:      a.ogImagePath(),
		twitterImagePath: a.twitterImagePath(),
		twitterSite:      strings.TrimSpace(a.TwitterSite),
		twitterCreator:   strings.TrimSpace(a.TwitterCreator),
		imageAlt:         a.imageAlt(),
		footer:           strings.TrimSpace(a.Footer),
		navLinks:         a.NavLinks,
		socialLinks:      a.SocialLinks,
		analytics:        a.Analytics,
		runnerEnabled:    runnerEnabled,
		sidebarDepth:     a.sidebarDepth(),
	}, nil
}
