package gomark

import (
	"io/fs"
	"path/filepath"
	"time"
)

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
	logoURL          string
	ogImagePath      string
	twitterImagePath string
	runnerEnabled    bool
	sidebarDepth     int
}

// buildSite assembles the shared artifacts. When preRender is true the page
// provider renders every page up front (required for export); otherwise it
// honors the configured render mode.
func (a *App) buildSite(preRender bool) (*siteBuild, error) {
	dir := a.contentDir()

	layoutPath, templateGlob := a.templatePaths()
	renderer, err := NewFileTemplateRenderer(layoutPath, templateGlob)
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
	sitemapXML, err := renderSitemapXML(siteURL, sitemapRoutes, time.Now())
	if err != nil {
		return nil, err
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
		robotsTXT:        renderRobotsTXT(siteURL),
		publicFS:         publicFS,
		siteURL:          siteURL,
		siteName:         a.siteTitle(),
		logoURL:          a.logoURL(),
		ogImagePath:      a.ogImagePath(),
		twitterImagePath: a.twitterImagePath(),
		runnerEnabled:    runnerEnabled,
		sidebarDepth:     a.sidebarDepth(),
	}, nil
}
