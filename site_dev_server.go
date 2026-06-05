package gomark

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

// liveBuild is an immutable snapshot of everything the live server needs to
// serve one revision of the site. A new one is produced on every file change
// and swapped in atomically, so in-flight requests always see a consistent view.
type liveBuild struct {
	build     *siteBuild
	routes    map[string]string // public route -> service slug
	landing   string
	oldBase   string
	responder HTMLErrorResponder
}

// liveServer is the development server. Unlike the static production path it
// dispatches requests against the current liveBuild, so adding, renaming, or
// deleting markdown is reflected without a restart: a watcher rebuilds the whole
// snapshot and swaps it in, then tells connected browsers to reload.
type liveServer struct {
	app     *App
	hub     *liveReloadHub
	current atomic.Pointer[liveBuild]

	wasmOnce sync.Once
	wasmGz   []byte
	wasmETag string
	wasmErr  error
}

func (s *Site) runLive(addr string) error {
	ls := &liveServer{app: &s.App, hub: newLiveReloadHub()}
	if err := ls.rebuild(); err != nil {
		return err
	}

	contentDir := ls.current.Load().build.contentDir

	stop := make(chan struct{})
	defer close(stop)
	go watchTree(contentDir, 400*time.Millisecond, func() {
		if err := ls.rebuild(); err != nil {
			log.Printf("live rebuild failed, keeping previous version: %v", err)
			return
		}
		log.Printf("content changed; rebuilt site")
		ls.hub.broadcast()
	}, stop)

	// The reload middleware injects the client into HTML (it skips the SSE
	// stream); logging sits inside it so timings reflect the real handler.
	handler := liveReloadMiddleware(LoggingMiddleware(ls))
	server := &http.Server{
		Addr:              addr,
		Handler:           handler,
		ReadHeaderTimeout: 5 * time.Second,
	}
	log.Printf("live reload enabled; watching %s for changes", contentDir)
	log.Printf("listening on %s", addr)
	return server.ListenAndServe()
}

// rebuild assembles a fresh liveBuild from the current content on disk and swaps
// it in. On failure the previous snapshot is left in place.
func (ls *liveServer) rebuild() error {
	b, err := ls.app.buildSite(false)
	if err != nil {
		return err
	}

	routes := make(map[string]string)
	err = eachContentRoute(b.contentDir, func(serviceSlug, pageRoute, path string) error {
		if existing, ok := routes[pageRoute]; ok {
			return fmt.Errorf("route collision for %s between %s and %s", pageRoute, existing, path)
		}
		routes[pageRoute] = serviceSlug
		return nil
	})
	if err != nil {
		return err
	}
	if len(routes) == 0 {
		return fmt.Errorf("no markdown files found in content dir %s", filepath.Clean(b.contentDir))
	}

	ls.current.Store(&liveBuild{
		build:     b,
		routes:    routes,
		landing:   landingRoute(routes),
		oldBase:   "/" + filepath.ToSlash(strings.Trim(b.contentDir, "/")),
		responder: b.errorResponder(log.Default()),
	})
	return nil
}

func (ls *liveServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	lb := ls.current.Load()
	if err := ls.serve(w, r, lb); err != nil {
		lb.responder.Handle(w, r, err)
	}
}

func (ls *liveServer) serve(w http.ResponseWriter, r *http.Request, lb *liveBuild) error {
	if r.Method != http.MethodGet {
		return &HTTPError{Status: http.StatusMethodNotAllowed, Message: "method not allowed"}
	}

	b := lb.build
	path := r.URL.Path

	switch path {
	case liveReloadPath:
		return ls.hub.handler(w, r)
	case "/sitemap.xml":
		if b.sitemapXML == "" {
			return &HTTPError{Status: http.StatusNotFound, Message: "page not found"}
		}
		w.Header().Set("Content-Type", "application/xml; charset=utf-8")
		_, err := w.Write([]byte(b.sitemapXML))
		return err
	case "/robots.txt":
		if b.robotsTXT == "" {
			return &HTTPError{Status: http.StatusNotFound, Message: "page not found"}
		}
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		_, err := w.Write([]byte(b.robotsTXT))
		return err
	case "/api/search":
		return serveSearch(w, r, b.searchIndex)
	case "/runner.wasm":
		if b.runnerEnabled {
			return ls.serveWASM(w, r, b.publicFS)
		}
	case "/":
		if slug, ok := lb.routes["/"]; ok {
			return b.renderContentPage(w, r, "/", slug)
		}
		http.Redirect(w, r, lb.landing, http.StatusFound)
		return nil
	}

	// Self-heal clients that cached an old "/content" -> base redirect.
	if lb.oldBase != "/" && (path == lb.oldBase || strings.HasPrefix(path, lb.oldBase+"/")) {
		target := strings.TrimPrefix(path, lb.oldBase)
		if target == "" {
			target = "/"
		}
		w.Header().Set("Clear-Site-Data", `"cache"`)
		http.Redirect(w, r, target, http.StatusFound)
		return nil
	}

	// Canonicalize trailing slashes so "/guides/install/" -> "/guides/install".
	if len(path) > 1 && strings.HasSuffix(path, "/") {
		target := strings.TrimRight(path, "/")
		if target == "" {
			target = "/"
		}
		if r.URL.RawQuery != "" {
			target += "?" + r.URL.RawQuery
		}
		http.Redirect(w, r, target, http.StatusFound)
		return nil
	}

	if slug, ok := lb.routes[path]; ok {
		return b.renderContentPage(w, r, path, slug)
	}

	exists, existsErr := staticFileExists(b.publicFS, path)
	if existsErr != nil {
		return existsErr
	}
	if exists {
		http.FileServerFS(b.publicFS).ServeHTTP(w, r)
		return nil
	}

	return &HTTPError{Status: http.StatusNotFound, Message: "page not found"}
}

// serveWASM serves the gzipped runner module. The bytes are read once and cached
// for the life of the server; the module does not change while developing.
func (ls *liveServer) serveWASM(w http.ResponseWriter, r *http.Request, publicFS fs.FS) error {
	ls.wasmOnce.Do(func() {
		ls.wasmGz, ls.wasmErr = fs.ReadFile(publicFS, "runner.wasm.gz")
		if ls.wasmErr == nil {
			ls.wasmETag = fmt.Sprintf("%q", fmt.Sprintf("%x", sha256.Sum256(ls.wasmGz)))
		}
	})
	if ls.wasmErr != nil {
		return fmt.Errorf("read runner.wasm.gz: %w", ls.wasmErr)
	}

	w.Header().Set("ETag", ls.wasmETag)
	w.Header().Set("Content-Type", "application/wasm")
	w.Header().Set("Cache-Control", "public, max-age=3600")
	if match := r.Header.Get("If-None-Match"); match != "" && match == ls.wasmETag {
		w.WriteHeader(http.StatusNotModified)
		return nil
	}
	w.Header().Set("Content-Encoding", "gzip")
	_, err := w.Write(ls.wasmGz)
	return err
}

// serveSearch answers the /api/search endpoint. It is shared by the static
// server and the live server so query behavior stays identical.
func serveSearch(w http.ResponseWriter, r *http.Request, index *SearchIndex) error {
	q := strings.TrimSpace(r.URL.Query().Get("q"))
	limit := 8
	if rawLimit := strings.TrimSpace(r.URL.Query().Get("limit")); rawLimit != "" {
		if parsed, parseErr := strconv.Atoi(rawLimit); parseErr == nil {
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
	return json.NewEncoder(w).Encode(map[string]any{"query": q, "results": index.Query(q, limit)})
}
