package gomark

import (
	"bytes"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// liveReloadPath is the SSE endpoint the injected client script connects to.
// It lives under a reserved prefix so it can never collide with a content route
// and so the injection middleware can skip it (streaming must not be buffered).
const liveReloadPath = "/__gomark/livereload"

const liveReloadSnippet = `<script>(function(){try{var es=new EventSource('` + liveReloadPath +
	`');es.onmessage=function(e){if(e.data==='reload'){es.close();location.reload();}};` +
	`es.onerror=function(){es.close();setTimeout(function(){location.reload();},1000);};}catch(e){}})();</script>`

// liveReloadHub fans a single "files changed" signal out to every connected
// browser tab over Server-Sent Events.
type liveReloadHub struct {
	mu   sync.Mutex
	subs map[chan struct{}]struct{}
}

func newLiveReloadHub() *liveReloadHub {
	return &liveReloadHub{subs: make(map[chan struct{}]struct{})}
}

func (h *liveReloadHub) subscribe() chan struct{} {
	ch := make(chan struct{}, 1)
	h.mu.Lock()
	h.subs[ch] = struct{}{}
	h.mu.Unlock()
	return ch
}

func (h *liveReloadHub) unsubscribe(ch chan struct{}) {
	h.mu.Lock()
	delete(h.subs, ch)
	h.mu.Unlock()
}

func (h *liveReloadHub) broadcast() {
	h.mu.Lock()
	defer h.mu.Unlock()
	for ch := range h.subs {
		select {
		case ch <- struct{}{}:
		default: // a reload is already pending for this tab; coalesce.
		}
	}
}

// handler streams reload events to one browser tab until it disconnects.
func (h *liveReloadHub) handler(w http.ResponseWriter, r *http.Request) error {
	flusher, ok := w.(http.Flusher)
	if !ok {
		return &HTTPError{Status: http.StatusInternalServerError, Message: "streaming unsupported"}
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.WriteHeader(http.StatusOK)
	flusher.Flush()

	ch := h.subscribe()
	defer h.unsubscribe(ch)

	for {
		select {
		case <-r.Context().Done():
			return nil
		case <-ch:
			if _, err := w.Write([]byte("data: reload\n\n")); err != nil {
				return nil
			}
			flusher.Flush()
		}
	}
}

// liveReloadMiddleware injects the reload client into HTML responses so the page
// can subscribe to change events. Only extensionless content pages are candidates
// for injection; assets (the wasm module, vendor JS/CSS, images), JSON APIs, and
// the SSE stream are written straight through without buffering, preserving their
// streaming/range semantics and avoiding a needless in-memory copy.
func liveReloadMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !isInjectablePath(r.URL.Path) {
			next.ServeHTTP(w, r)
			return
		}

		rec := &bufferedResponseWriter{header: make(http.Header)}
		next.ServeHTTP(rec, r)

		body := rec.buf.Bytes()
		contentType := rec.header.Get("Content-Type")
		if contentType == "" {
			contentType = http.DetectContentType(body)
		}

		for key, values := range rec.header {
			w.Header()[key] = values
		}

		if strings.Contains(contentType, "text/html") {
			body = injectLiveReload(body)
			w.Header().Del("Content-Length") // length changed after injection.
		}

		status := rec.status
		if status == 0 {
			status = http.StatusOK
		}
		w.WriteHeader(status)
		_, _ = w.Write(body)
	})
}

// isInjectablePath reports whether a request could return an HTML document worth
// buffering for reload-client injection. Content pages are served at extensionless
// routes; anything with a file extension (assets, sitemap.xml, robots.txt,
// search-index.json), the JSON search API, and the reserved SSE endpoint are not
// HTML and stream straight through.
func isInjectablePath(requestPath string) bool {
	if strings.HasPrefix(requestPath, "/__gomark/") || strings.HasPrefix(requestPath, "/api/") {
		return false
	}
	return path.Ext(requestPath) == ""
}

// injectLiveReload places the client script just before </body> (or appends it
// when there is no closing tag).
func injectLiveReload(body []byte) []byte {
	marker := []byte("</body>")
	idx := bytes.LastIndex(body, marker)
	if idx < 0 {
		return append(body, []byte(liveReloadSnippet)...)
	}
	out := make([]byte, 0, len(body)+len(liveReloadSnippet))
	out = append(out, body[:idx]...)
	out = append(out, []byte(liveReloadSnippet)...)
	out = append(out, body[idx:]...)
	return out
}

// bufferedResponseWriter captures a handler's response so the middleware can
// rewrite the body before it reaches the client.
type bufferedResponseWriter struct {
	header      http.Header
	buf         bytes.Buffer
	status      int
	wroteHeader bool
}

func (b *bufferedResponseWriter) Header() http.Header { return b.header }

func (b *bufferedResponseWriter) WriteHeader(status int) {
	if !b.wroteHeader {
		b.status = status
		b.wroteHeader = true
	}
}

func (b *bufferedResponseWriter) Write(p []byte) (int, error) {
	if !b.wroteHeader {
		b.WriteHeader(http.StatusOK)
	}
	return b.buf.Write(p)
}

// fileSnapshot maps each file under a tree to a fingerprint of its size and
// modification time, so changes can be detected by comparing two snapshots.
type fileSnapshot map[string]int64

func snapshotTree(root string) fileSnapshot {
	snap := make(fileSnapshot)
	_ = filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return nil
		}
		info, infoErr := d.Info()
		if infoErr != nil {
			return nil
		}
		snap[path] = info.ModTime().UnixNano() ^ info.Size()
		return nil
	})
	return snap
}

func snapshotsEqual(a, b fileSnapshot) bool {
	if len(a) != len(b) {
		return false
	}
	for path, fingerprint := range a {
		if b[path] != fingerprint {
			return false
		}
	}
	return true
}

// watchTree polls root on an interval and calls onChange whenever a file is
// added, removed, or modified. Polling keeps the dev server dependency-free
// (no fsnotify) and is plenty responsive for a docs tree. It returns when stop
// is closed.
func watchTree(root string, interval time.Duration, onChange func(), stop <-chan struct{}) {
	previous := snapshotTree(root)
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-stop:
			return
		case <-ticker.C:
			current := snapshotTree(root)
			if !snapshotsEqual(previous, current) {
				previous = current
				onChange()
			}
		}
	}
}
