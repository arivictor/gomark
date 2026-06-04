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

func (a *App) registerContentRoutes(app *Server, renderer *FileTemplateRenderer, dir string, index *ContentIndex, topNav []NavLink, siteURL, siteName, logoURL, ogImagePath, twitterImagePath string, markdownRenderer MarkdownRenderer, RunnerEnabled bool) (string, error) {
	cleanDir := filepath.Clean(dir)
	provider, err := a.newContentPageProvider(cleanDir, markdownRenderer)
	if err != nil {
		return "", err
	}
	depth := a.sidebarDepth()
	registered := map[string]string{}

	err = filepath.WalkDir(cleanDir, func(path string, d os.DirEntry, walkErr error) error {
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

		pageTitle := pageTitleFromSlug(serviceSlug)
		app.Handle("GET", registerPath, func(w http.ResponseWriter, r *http.Request) error {
			page, getErr := provider.Get(serviceSlug)
			if getErr != nil {
				return mapContentProviderError(getErr)
			}

			title := page.Title
			if strings.TrimSpace(title) == "" {
				title = pageTitle
			}

			navTitle, nav := index.Sidebar(pageRoute, depth)
			baseURL := requestBaseURL(r, siteURL)

			return renderer.Render(w, "markdown", withCSRFToken(w, r, PageData{
				Title:           title,
				Description:     page.Description,
				SiteName:        siteName,
				LogoURL:         logoURL,
				CanonicalURL:    joinAbsoluteURL(baseURL, pageRoute),
				OGImageURL:      joinAbsoluteURL(baseURL, ogImagePath),
				TwitterImageURL: joinAbsoluteURL(baseURL, twitterImagePath),
				RunnerEnabled:   RunnerEnabled,
				Robots:          "index,follow",
				Time:            time.Now().UTC().Format(time.RFC3339),
				MarkdownFile:    page.Path,
				BodyHTML:        template.HTML(page.HTML),
				Headings:        page.Headings,
				NavTitle:        navTitle,
				Nav:             nav,
				TopNav:          topNav,
				CurrentPath:     pageRoute,
			}))
		})

		return nil
	})
	if err != nil {
		return "", err
	}

	if len(registered) == 0 {
		return "", fmt.Errorf("no markdown files found in content dir %s", cleanDir)
	}

	return landingRoute(registered), nil
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
