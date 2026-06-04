package site

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/arivictor/gomark/internal/protocol"
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
	var runnerClient *RunnerClient
	if RunnerEnabled {
		runnerClient, err = NewRunnerClient(RunnerURL, protocol.AuthConfig{
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
			var req protocol.RunRequest
			if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 128<<10)).Decode(&req); err != nil {
				return &BadRequestError{Message: "invalid run request"}
			}

			resp, runErr := runnerClient.Run(r.Context(), req)
			if runErr != nil {
				w.Header().Set("Content-Type", "application/json; charset=utf-8")
				w.WriteHeader(http.StatusBadGateway)
				return json.NewEncoder(w).Encode(protocol.RunResponse{OK: false, Error: "cannot run"})
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
