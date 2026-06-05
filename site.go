package gomark

import (
	"crypto/sha256"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
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
	// The live development server rebuilds routes, navigation, search, and the
	// sitemap whenever files change, so it dispatches dynamically rather than
	// registering a fixed route table. Production serving stays static below.
	if live {
		return s.runLive(addr)
	}

	a := &s.App

	b, err := a.buildSite(false)
	if err != nil {
		return err
	}

	dir := b.contentDir
	siteURL := b.siteURL
	RunnerEnabled := b.runnerEnabled
	searchIndex := b.searchIndex
	sitemapRoutes := b.sitemapRoutes
	sitemapXML := b.sitemapXML
	robotsTXT := b.robotsTXT

	httpApp := NewServer(b.errorResponder(log.Default()))
	httpApp.Use(LoggingMiddleware)
	httpApp.Use(CSRFProtectionMiddleware(siteURL))
	log.Printf("seo sitemap generated with %d routes", len(sitemapRoutes))
	if sitemapXML != "" {
		httpApp.Handle("GET", "/sitemap.xml", func(w http.ResponseWriter, r *http.Request) error {
			w.Header().Set("Content-Type", "application/xml; charset=utf-8")
			_, writeErr := w.Write([]byte(sitemapXML))
			return writeErr
		})
	}
	if robotsTXT != "" {
		httpApp.Handle("GET", "/robots.txt", func(w http.ResponseWriter, r *http.Request) error {
			w.Header().Set("Content-Type", "text/plain; charset=utf-8")
			_, writeErr := w.Write([]byte(robotsTXT))
			return writeErr
		})
	}
	httpApp.Handle("GET", "/api/search", func(w http.ResponseWriter, r *http.Request) error {
		return serveSearch(w, r, searchIndex)
	})
	// publicFS serves static assets: the embedded public/ tree, with an on-disk
	// PublicDir overlaid on top when configured (a site's own favicons, og-image,
	// logos, … override the bundled ones). The runner module is sourced from here
	// too, so the overlay keeps runner.wasm and wasm_exec.js consistent.
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

	landing, err := registerContentRoutes(httpApp, b)
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

	return httpApp.Run(addr)
}
