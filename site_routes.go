package gomark

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// eachContentRoute walks the content dir and invokes fn with each markdown
// file's service slug (path without extension) and its resolved public route,
// honoring frontmatter slug/permalink overrides. Both the live route registrar
// and the static exporter use it so the route↔slug mapping can never drift.
func eachContentRoute(contentDir string, fn func(slug, route, path string) error) error {
	cleanDir := filepath.Clean(contentDir)
	return filepath.WalkDir(cleanDir, func(path string, d os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if d.IsDir() || filepath.Ext(path) != ".md" {
			return nil
		}

		rel, relErr := filepath.Rel(cleanDir, path)
		if relErr != nil {
			return relErr
		}

		serviceSlug := strings.TrimSuffix(filepath.ToSlash(rel), filepath.Ext(rel))
		routePath := buildRoutePath(serviceSlug)

		// An index.md serves at its folder path ("/go/philosophy"), not at
		// "/go/philosophy/index"; the content-root index.md serves at "/".
		pageRoute := routePath
		if strings.HasSuffix(routePath, "/index") {
			pageRoute = strings.TrimSuffix(routePath, "/index")
			if pageRoute == "" {
				pageRoute = "/"
			}
		}

		// A page may override its served route via frontmatter (slug/permalink).
		if data, readErr := os.ReadFile(path); readErr == nil {
			if meta, _ := parseFrontmatter(string(data)); meta != nil {
				if override := routeFromFrontmatter(meta); override != "" {
					pageRoute = override
				}
			}
		}

		return fn(serviceSlug, pageRoute, path)
	})
}

func registerContentRoutes(app *Server, b *siteBuild) (string, error) {
	registered := map[string]string{}

	err := eachContentRoute(b.contentDir, func(serviceSlug, pageRoute, path string) error {
		if existing, exists := registered[pageRoute]; exists {
			return fmt.Errorf("route collision for %s between %s and %s", pageRoute, existing, path)
		}
		registered[pageRoute] = path

		// The root index registers at "/{$}" (exact root) so the subtree "/" stays
		// the static/catch-all handler.
		registerPath := pageRoute
		if registerPath == "/" {
			registerPath = "/{$}"
		}

		route, slug := pageRoute, serviceSlug
		app.Handle("GET", registerPath, func(w http.ResponseWriter, r *http.Request) error {
			return b.renderContentPage(w, r, route, slug)
		})

		return nil
	})
	if err != nil {
		return "", err
	}

	if len(registered) == 0 {
		return "", fmt.Errorf("no markdown files found in content dir %s", filepath.Clean(b.contentDir))
	}

	return landingRoute(registered), nil
}

// renderContentPage renders the page for route (served from serviceSlug) using
// the shared build artifacts. Both the static route registrar and the live
// development server call it, so served and live pages can never drift.
func (b *siteBuild) renderContentPage(w http.ResponseWriter, r *http.Request, route, serviceSlug string) error {
	page, getErr := b.provider.Get(serviceSlug)
	if getErr != nil {
		return mapContentProviderError(getErr)
	}

	title := page.Title
	if strings.TrimSpace(title) == "" {
		title = pageTitleFromSlug(serviceSlug)
	}

	navTitle, nav := b.index.Sidebar(route, b.sidebarDepth)
	prevPage, nextPage := b.index.Siblings(route)
	baseURL := requestBaseURL(r, b.siteURL)

	description := page.Description
	if strings.TrimSpace(description) == "" {
		description = b.description
	}

	return b.renderer.Render(w, "markdown", withCSRFToken(w, r, PageData{
		Title:           title,
		Description:     description,
		SiteName:        b.siteName,
		Lang:            b.lang,
		ThemeColor:      b.themeColor,
		LogoLightURL:    b.logoLight,
		LogoDarkURL:     b.logoDark,
		CanonicalURL:    joinAbsoluteURL(baseURL, route),
		OGImageURL:      joinAbsoluteURL(baseURL, b.ogImagePath),
		TwitterImageURL: joinAbsoluteURL(baseURL, b.twitterImagePath),
		TwitterSite:     b.twitterSite,
		TwitterCreator:  b.twitterCreator,
		ImageAlt:        b.imageAlt,
		FooterText:      b.footer,
		NavLinks:        b.navLinks,
		SocialLinks:     b.socialLinks,
		Analytics:       b.analytics,
		RunnerEnabled:   b.runnerEnabled,
		Robots:          "index,follow",
		Time:            time.Now().UTC().Format(time.RFC3339),
		MarkdownFile:    page.Path,
		BodyHTML:        template.HTML(page.HTML),
		Headings:        page.Headings,
		HideTOC:         page.HideTOC,
		HideNav:         page.HideNav,
		NavTitle:        navTitle,
		Nav:             nav,
		TopNav:          b.topNav,
		CurrentPath:     route,
		PrevPage:        prevPage,
		NextPage:        nextPage,
	}))
}

func (a *App) newContentPageProvider(contentDir string, renderer MarkdownRenderer) (contentPageProvider, error) {
	switch a.mode() {
	case PreRender:
		startedAt := time.Now()
		provider, err := newPreRenderedMarkdownProvider(contentDir, renderer)
		if err != nil {
			return nil, err
		}
		if preRendered, ok := provider.(preRenderedMarkdownProvider); ok {
			log.Printf("content mode=%s built %d markdown pages in %s", PreRender, len(preRendered.pages), time.Since(startedAt))
		} else {
			log.Printf("content mode=%s initialized in %s", PreRender, time.Since(startedAt))
		}
		return provider, nil
	default:
		log.Printf("content mode=%s using per-request live markdown rendering", LiveRender)
		return newLiveMarkdownProvider(contentDir, renderer), nil
	}
}

// landingRoute is only used as the bare-root fallback when there is no root
// index.md: it prefers "/" (the content-root index) and otherwise sends visitors
// to the lexically first registered route.
func landingRoute(registered map[string]string) string {
	if _, ok := registered["/"]; ok {
		return "/"
	}

	routes := make([]string, 0, len(registered))
	for route := range registered {
		routes = append(routes, route)
	}
	sort.Strings(routes)
	if len(routes) == 0 {
		return "/"
	}
	return routes[0]
}
