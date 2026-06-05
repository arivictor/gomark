package gomark

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

const defaultSiteName = "GoMark"
const defaultSiteURL = ""
const defaultOGImagePath = "/gomark-og-1200x630.png"
const defaultTwitterImagePath = "/gomark-twitter-1200x628.png"

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

	// When an export target is configured (option or EXPORT_DIR), build the
	// static site and exit instead of serving — handy for CI pipelines.
	if dir := s.App.exportDir(); dir != "" {
		log.Printf("exporting static site to %s", dir)
		return s.Export(dir)
	}

	addr := strings.TrimSpace(s.addr)
	if addr == "" {
		port := strings.TrimSpace(os.Getenv("PORT"))
		if port == "" {
			port = "8080"
		}
		addr = ":" + port
	}

	return s.run(addr, false)
}

// Serve runs the local development server on addr. When live is true it renders
// each page on every request and live-reloads connected browsers as files under
// the content dir change. Serve is a development tool — production deployments
// build a static site with Export (see the `gomark build` CLI command).
func (s *Site) Serve(addr string, live bool) error {
	if s == nil {
		return fmt.Errorf("site is nil")
	}
	addr = strings.TrimSpace(addr)
	if addr == "" {
		addr = ":8080"
	}
	return s.run(addr, live)
}

func (s *Site) run(addr string, live bool) error {
	a := &s.App

	b, err := a.buildSite(false)
	if err != nil {
		return err
	}

	dir := b.contentDir
	renderer := b.renderer
	index := b.index
	appTitle := b.siteName
	appLogo := b.logoURL
	ogImagePath := b.ogImagePath
	twitterImagePath := b.twitterImagePath
	siteURL := b.siteURL
	RunnerEnabled := b.runnerEnabled
	searchIndex := b.searchIndex
	topNav := b.topNav
	sitemapRoutes := b.sitemapRoutes
	sitemapXML := b.sitemapXML
	robotsTXT := b.robotsTXT

	httpApp := NewServer(HTMLErrorResponder{Renderer: renderer, TopNav: topNav, SiteName: appTitle, LogoURL: appLogo, SiteURL: siteURL, OGImagePath: ogImagePath, TwitterImagePath: twitterImagePath, Logger: log.Default()})
	// In live mode the reload middleware is outermost so it can inject the
	// client into the final HTML (including error pages) before it ships.
	var hub *liveReloadHub
	if live {
		hub = newLiveReloadHub()
		httpApp.Use(liveReloadMiddleware)
	}
	httpApp.Use(LoggingMiddleware)
	httpApp.Use(CSRFProtectionMiddleware(siteURL))
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
	// publicFS serves static assets: an on-disk directory when configured,
	// otherwise the embedded public/ tree. The runner module is sourced from
	// here too, so overriding assets via PublicDir keeps runner.wasm and
	// wasm_exec.js consistent with each other.
	publicFS := b.publicFS

	if RunnerEnabled {
		// The runner executes entirely in the browser via a WebAssembly build of
		// the yaegi interpreter. The pre-gzipped module is read once at startup
		// and served with Content-Encoding: gzip plus an ETag, so browsers
		// revalidate cheaply (a 304 when unchanged) instead of holding a stale
		// module under a long immutable cache; wasm_exec.js is a static asset.
		wasmGz, wasmErr := fs.ReadFile(publicFS, "runner.wasm.gz")
		if wasmErr != nil {
			return fmt.Errorf("read runner.wasm.gz: %w", wasmErr)
		}
		wasmETag := fmt.Sprintf("%q", fmt.Sprintf("%x", sha256.Sum256(wasmGz)))
		httpApp.Handle("GET", "/runner.wasm", func(w http.ResponseWriter, r *http.Request) error {
			w.Header().Set("ETag", wasmETag)
			w.Header().Set("Content-Type", "application/wasm")
			w.Header().Set("Cache-Control", "public, max-age=3600")
			if match := r.Header.Get("If-None-Match"); match != "" && match == wasmETag {
				w.WriteHeader(http.StatusNotModified)
				return nil
			}
			w.Header().Set("Content-Encoding", "gzip")
			_, writeErr := w.Write(wasmGz)
			return writeErr
		})
	}

	landing, err := a.registerContentRoutes(httpApp, renderer, dir, index, topNav, siteURL, appTitle, appLogo, ogImagePath, twitterImagePath, b.provider, RunnerEnabled)
	if err != nil {
		return err
	}

	// The content dir is mounted at "/". The root index.md (if any) is registered
	// at "/{$}", so "/" stays the catch-all: canonicalize trailing slashes, then
	// serve static assets (favicons, og-image…) from the public dir. The bare-root
	// redirect only fires when there is no root index.md.
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

	if live {
		httpApp.Handle("GET", liveReloadPath, hub.handler)
		stop := make(chan struct{})
		defer close(stop)
		go watchTree(dir, 400*time.Millisecond, hub.broadcast, stop)
		log.Printf("live reload enabled; watching %s for changes", dir)
	}

	return httpApp.Run(addr)
}
