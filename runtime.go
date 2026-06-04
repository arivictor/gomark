package gomark

import (
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"io/fs"
	"log"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"
)

const defaultSiteName = "GoMark"
const defaultSiteURL = ""
const defaultOGImagePath = "/gomark-og-1200x630.png"
const defaultTwitterImagePath = "/gomark-twitter-1200x628.png"

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
	RunnerAuthMode  RunnerAuthMode
	RunnerAuthToken string
}

type RenderMode string

const (
	LiveRender RenderMode = "live_render"
	PreRender  RenderMode = "pre_render"
)

type SiteOption func(*Site)

type Site struct {
	App  App
	addr string
}

func NewSite(options ...SiteOption) *Site {
	s := &Site{}
	for _, option := range options {
		if option == nil {
			continue
		}
		option(s)
	}
	return s
}

func (s *Site) Start() error {
	if s == nil {
		return fmt.Errorf("site is nil")
	}

	addr := strings.TrimSpace(s.addr)
	if addr == "" {
		port := strings.TrimSpace(os.Getenv("PORT"))
		if port == "" {
			port = "8080"
		}
		addr = ":" + port
	}

	return s.run(addr)
}

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

func WithSiteRunnerAuth(mode RunnerAuthMode, token string) SiteOption {
	return func(s *Site) {
		s.App.RunnerAuthMode = mode
		s.App.RunnerAuthToken = strings.TrimSpace(token)
	}
}

func WithSiteRunner(url string, mode RunnerAuthMode, token string) SiteOption {
	return func(s *Site) {
		s.App.RunnerEnabled = true
		s.App.RunnerURL = strings.TrimSpace(url)
		s.App.RunnerAuthMode = mode
		s.App.RunnerAuthToken = strings.TrimSpace(token)
	}
}

func (s *Site) run(addr string) error {
	a := &s.App

	dir := a.contentDir()
	layoutPath, templateGlob := a.templatePaths()
	renderer, err := NewFileTemplateRenderer(layoutPath, templateGlob)
	if err != nil {
		return err
	}

	index, err := BuildContentIndex(dir)
	if err != nil {
		return err
	}
	appTitle := a.siteTitle()
	appLogo := a.logoURL()
	siteURL := a.siteURL()
	RunnerEnabled := a.GetRunnerEnabled()
	RunnerURL := a.GetRunnerURL()
	RunnerAuthMode := a.GetRunnerAuthMode()
	RunnerAuthToken := a.GetRunnerAuthToken()
	var RunnerClient *RunnerClient
	if RunnerEnabled {
		RunnerClient, err = NewRunnerClient(RunnerURL, RunnerAuthConfig{
			Mode:        RunnerAuthMode,
			BearerToken: RunnerAuthToken,
		})
		if err != nil {
			return err
		}
	}
	searchIndex, err := BuildSearchIndex(dir)
	if err != nil {
		return err
	}
	topNav := index.TopNav()
	sitemapRoutes := buildSitemapRoutes(index)
	sitemapXML, err := renderSitemapXML(siteURL, sitemapRoutes, time.Now())
	if err != nil {
		return err
	}
	robotsTXT := renderRobotsTXT(siteURL)

	httpApp := NewServer(HTMLErrorResponder{Renderer: renderer, TopNav: topNav, SiteName: appTitle, LogoURL: appLogo, SiteURL: siteURL, Logger: log.Default()})
	httpApp.Use(LoggingMiddleware)
	log.Printf("seo sitemap generated with %d routes", len(sitemapRoutes))
	httpApp.Handle("GET", "/sitemap.xml", func(w http.ResponseWriter, r *http.Request) error {
		w.Header().Set("Content-Type", "application/xml; charset=utf-8")
		_, writeErr := w.Write([]byte(sitemapXML))
		return writeErr
	})
	httpApp.Handle("GET", "/robots.txt", func(w http.ResponseWriter, r *http.Request) error {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		_, writeErr := w.Write([]byte(robotsTXT))
		return writeErr
	})
	httpApp.Handle("GET", "/api/search", func(w http.ResponseWriter, r *http.Request) error {
		q := strings.TrimSpace(r.URL.Query().Get("q"))
		limit := 8
		if rawLimit := strings.TrimSpace(r.URL.Query().Get("limit")); rawLimit != "" {
			parsed, parseErr := strconv.Atoi(rawLimit)
			if parseErr == nil {
				if parsed < 1 {
					parsed = 1
				}
				if parsed > 25 {
					parsed = 25
				}
				limit = parsed
			}
		}

		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		if q == "" {
			return json.NewEncoder(w).Encode(map[string]any{"query": "", "results": []SearchResult{}})
		}

		results := searchIndex.Query(q, limit)
		return json.NewEncoder(w).Encode(map[string]any{"query": q, "results": results})
	})
	if RunnerEnabled {
		httpApp.Handle("POST", "/api/runner/run", func(w http.ResponseWriter, r *http.Request) error {
			var req RunnerRunRequest
			if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 128<<10)).Decode(&req); err != nil {
				return &BadRequestError{Message: "invalid run request"}
			}

			resp, runErr := RunnerClient.Run(r.Context(), req)
			if runErr != nil {
				w.Header().Set("Content-Type", "application/json; charset=utf-8")
				w.WriteHeader(http.StatusBadGateway)
				return json.NewEncoder(w).Encode(RunnerRunResponse{OK: false, Error: "cannot run"})
			}

			w.Header().Set("Content-Type", "application/json; charset=utf-8")
			return json.NewEncoder(w).Encode(resp)
		})
	}

	landing, err := a.registerContentRoutes(httpApp, renderer, dir, index, topNav, siteURL, appTitle, appLogo, StdlibMarkdownRenderer{RunnerEnabled: RunnerEnabled}, RunnerEnabled)
	if err != nil {
		return err
	}

	// The content dir is mounted at "/". The root index.md (if any) is registered
	// at "/{$}", so "/" stays the catch-all: canonicalize trailing slashes, then
	// serve static assets (favicons, og-image…) from the public dir. The bare-root
	// redirect only fires when there is no root index.md.
	publicFS, err := a.publicFS()
	if err != nil {
		return err
	}
	staticFiles := http.FileServerFS(publicFS)
	// An earlier version redirected "/" -> oldBase (e.g. "/content") with a 301,
	// which browsers cache permanently. Self-heal those clients: clear their cache
	// and bounce to the de-prefixed path. Harmless once no client has the stale 301.
	oldBase := "/" + filepath.ToSlash(strings.Trim(dir, "/"))
	httpApp.Handle("GET", "/", func(w http.ResponseWriter, r *http.Request) error {
		if r.URL.Path == "/" {
			http.Redirect(w, r, landing, http.StatusFound)
			return nil
		}
		if oldBase != "/" && (r.URL.Path == oldBase || strings.HasPrefix(r.URL.Path, oldBase+"/")) {
			target := strings.TrimPrefix(r.URL.Path, oldBase)
			if target == "" {
				target = "/"
			}
			w.Header().Set("Clear-Site-Data", `"cache"`)
			http.Redirect(w, r, target, http.StatusFound)
			return nil
		}
		if p := r.URL.Path; len(p) > 1 && strings.HasSuffix(p, "/") {
			target := strings.TrimRight(p, "/")
			if target == "" {
				target = "/"
			}
			if r.URL.RawQuery != "" {
				target += "?" + r.URL.RawQuery
			}
			http.Redirect(w, r, target, http.StatusFound)
			return nil
		}

		exists, existsErr := staticFileExists(publicFS, r.URL.Path)
		if existsErr != nil {
			return existsErr
		}
		if exists {
			staticFiles.ServeHTTP(w, r)
			return nil
		}

		return &HTTPError{Status: http.StatusNotFound, Message: "page not found"}
	})

	return httpApp.Run(addr)
}

func (a *App) registerContentRoutes(app *Server, renderer *FileTemplateRenderer, dir string, index *ContentIndex, topNav []NavLink, siteURL, siteName, logoURL string, markdownRenderer MarkdownRenderer, RunnerEnabled bool) (string, error) {
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

			return renderer.Render(w, "markdown", PageData{
				Title:           title,
				Description:     page.Description,
				SiteName:        siteName,
				LogoURL:         logoURL,
				CanonicalURL:    joinAbsoluteURL(baseURL, pageRoute),
				OGImageURL:      joinAbsoluteURL(baseURL, defaultOGImagePath),
				TwitterImageURL: joinAbsoluteURL(baseURL, defaultTwitterImagePath),
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
			})
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

func (a *App) GetRunnerAuthMode() RunnerAuthMode {
	if strings.TrimSpace(string(a.RunnerAuthMode)) != "" {
		return RunnerAuthMode(strings.ToLower(strings.TrimSpace(string(a.RunnerAuthMode))))
	}
	return RunnerAuthMode(strings.ToLower(strings.TrimSpace(os.Getenv("PLAYGROUND_RUNNER_AUTH_MODE"))))
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

// buildRoutePath mounts the content dir at the site root: the dir name does not
// appear in the URL, so "go/about" -> "/go/about" and "" -> "/".
func buildRoutePath(serviceSlug string) string {
	slug := strings.Trim(strings.TrimSpace(serviceSlug), "/")
	if slug == "" {
		return "/"
	}
	return "/" + slug
}

func pageTitleFromSlug(slug string) string {
	parts := strings.Split(strings.TrimSpace(slug), "/")
	if len(parts) == 0 {
		return "Content"
	}
	last := parts[len(parts)-1]
	if last == "index" && len(parts) > 1 {
		last = parts[len(parts)-2]
	}
	if last == "" {
		last = "Content"
	}

	words := strings.Fields(strings.ReplaceAll(last, "-", " "))
	for i, word := range words {
		r := []rune(word)
		if len(r) == 0 {
			continue
		}
		r[0] = []rune(strings.ToUpper(string(r[0])))[0]
		words[i] = string(r)
	}
	if len(words) == 0 {
		return "Content"
	}
	return strings.Join(words, " ")
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

func (a *App) publicFS() (fs.FS, error) {
	if dir := strings.TrimSpace(a.PublicDir); dir != "" {
		return os.DirFS(filepath.Clean(dir)), nil
	}

	publicFS, err := fs.Sub(embeddedPublicFS, "public")
	if err != nil {
		return nil, err
	}

	return publicFS, nil
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

func normalizeBaseURL(raw string) string {
	base := strings.TrimSpace(raw)
	if base == "" {
		return defaultSiteURL
	}
	if !strings.HasPrefix(base, "http://") && !strings.HasPrefix(base, "https://") {
		base = "https://" + base
	}
	return strings.TrimRight(base, "/")
}

func requestBaseURL(r *http.Request, fallback string) string {
	if r == nil {
		return normalizeBaseURL(fallback)
	}

	host := strings.TrimSpace(r.Header.Get("X-Forwarded-Host"))
	if host == "" {
		host = strings.TrimSpace(r.Host)
	}
	if host == "" {
		return normalizeBaseURL(fallback)
	}
	if i := strings.Index(host, ","); i != -1 {
		host = strings.TrimSpace(host[:i])
	}

	proto := strings.TrimSpace(r.Header.Get("X-Forwarded-Proto"))
	if i := strings.Index(proto, ","); i != -1 {
		proto = strings.TrimSpace(proto[:i])
	}
	if proto == "" {
		if r.TLS != nil {
			proto = "https"
		} else {
			proto = "http"
		}
	}

	return proto + "://" + host
}

func joinAbsoluteURL(base, route string) string {
	cleanBase := normalizeBaseURL(base)
	cleanRoute := strings.TrimSpace(route)
	if cleanRoute == "" {
		cleanRoute = "/"
	}
	if !strings.HasPrefix(cleanRoute, "/") {
		cleanRoute = "/" + cleanRoute
	}
	return cleanBase + cleanRoute
}

func staticFileExists(publicFS fs.FS, requestPath string) (bool, error) {
	relReq := strings.TrimPrefix(requestPath, "/")
	relReq = path.Clean(relReq)

	if relReq == "." || relReq == "" || strings.HasPrefix(relReq, "../") || relReq == ".." {
		return false, nil
	}

	info, statErr := fs.Stat(publicFS, relReq)
	if statErr != nil {
		if errors.Is(statErr, fs.ErrNotExist) {
			return false, nil
		}
		return false, statErr
	}

	return !info.IsDir(), nil
}
