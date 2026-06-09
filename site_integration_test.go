package gomark

import (
	"io"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// freeAddr returns a loopback address that was free a moment ago.
func freeAddr(t *testing.T) string {
	t.Helper()
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	addr := ln.Addr().String()
	ln.Close()
	return addr
}

// waitForServer blocks until addr accepts a connection or the deadline passes.
func waitForServer(t *testing.T, addr string) {
	t.Helper()
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		c, err := net.Dial("tcp", addr)
		if err == nil {
			c.Close()
			return
		}
		time.Sleep(15 * time.Millisecond)
	}
	t.Fatalf("server at %s did not come up", addr)
}

// noRedirectClient does not follow redirects, so tests can inspect 3xx responses.
func noRedirectClient() *http.Client {
	return &http.Client{
		Timeout: 5 * time.Second,
		CheckRedirect: func(*http.Request, []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
}

func TestServeStaticIntegration(t *testing.T) {
	dir := t.TempDir()
	writeExportFile(t, dir, "index.md", "# Home\n\nWelcome.")
	writeExportFile(t, dir, "about.md", "# About")

	addr := freeAddr(t)
	site := NewSite(WithSiteContentDir(dir), WithSiteURL("https://example.com"))
	go func() { _ = site.Serve(addr, false) }()
	waitForServer(t, addr)

	client := noRedirectClient()
	base := "http://" + addr

	cases := []struct {
		path string
		code int
	}{
		{"/about", http.StatusOK},
		{"/", http.StatusOK}, // root index.md registered at /{$}
		{"/sitemap.xml", http.StatusOK},
		{"/robots.txt", http.StatusOK},
		{"/api/search?q=welcome", http.StatusOK},
		{"/runner.wasm", http.StatusOK},
		{"/favicon/favicon.ico", http.StatusOK},
		{"/nope", http.StatusNotFound},
		{"/about/", http.StatusFound}, // trailing slash canonicalization
	}
	for _, tc := range cases {
		resp, err := client.Get(base + tc.path)
		if err != nil {
			t.Fatalf("GET %s: %v", tc.path, err)
		}
		io.Copy(io.Discard, resp.Body)
		resp.Body.Close()
		if resp.StatusCode != tc.code {
			t.Errorf("GET %s = %d, want %d", tc.path, resp.StatusCode, tc.code)
		}
	}

	// runner.wasm conditional revalidation -> 304.
	resp, err := client.Get(base + "/runner.wasm")
	if err != nil {
		t.Fatalf("runner: %v", err)
	}
	etag := resp.Header.Get("ETag")
	io.Copy(io.Discard, resp.Body)
	resp.Body.Close()
	req, _ := http.NewRequest("GET", base+"/runner.wasm", nil)
	req.Header.Set("If-None-Match", etag)
	resp2, err := client.Do(req)
	if err != nil {
		t.Fatalf("runner 304: %v", err)
	}
	io.Copy(io.Discard, resp2.Body)
	resp2.Body.Close()
	if resp2.StatusCode != http.StatusNotModified {
		t.Fatalf("expected 304, got %d", resp2.StatusCode)
	}
}

func TestServeStaticRootRedirectWithoutIndex(t *testing.T) {
	dir := t.TempDir()
	writeExportFile(t, dir, "alpha.md", "# Alpha")

	addr := freeAddr(t)
	site := NewSite(WithSiteContentDir(dir))
	go func() { _ = site.Serve(addr, false) }()
	waitForServer(t, addr)

	resp, err := noRedirectClient().Get("http://" + addr + "/")
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	io.Copy(io.Discard, resp.Body)
	resp.Body.Close()
	if resp.StatusCode != http.StatusFound {
		t.Fatalf("expected redirect, got %d", resp.StatusCode)
	}
	if loc := resp.Header.Get("Location"); loc != "/alpha" {
		t.Fatalf("location = %q", loc)
	}
}

func TestServeStaticOldBaseRedirect(t *testing.T) {
	// A content dir literally named "content" makes oldBase "/content".
	base := t.TempDir()
	content := filepath.Join(base, "content")
	writeExportFile(t, content, "index.md", "# Home")
	writeExportFile(t, content, "about.md", "# About")

	addr := freeAddr(t)
	site := NewSite(WithSiteContentDir(content))
	go func() { _ = site.Serve(addr, false) }()
	waitForServer(t, addr)

	// The static server derives oldBase from the (absolute) content dir path.
	oldBase := "/" + strings.Trim(filepath.ToSlash(content), "/")
	resp, err := noRedirectClient().Get("http://" + addr + oldBase + "/about")
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	io.Copy(io.Discard, resp.Body)
	resp.Body.Close()
	if resp.StatusCode != http.StatusFound {
		t.Fatalf("expected redirect from old base, got %d", resp.StatusCode)
	}
	if resp.Header.Get("Clear-Site-Data") == "" {
		t.Fatal("expected Clear-Site-Data header")
	}
	if loc := resp.Header.Get("Location"); loc != "/about" {
		t.Fatalf("location = %q", loc)
	}
}

func TestStartUsesPortEnv(t *testing.T) {
	dir := t.TempDir()
	writeExportFile(t, dir, "index.md", "# Home")

	addr := freeAddr(t)
	port := addr[strings.LastIndex(addr, ":")+1:]
	t.Setenv("PORT", port)
	t.Setenv("EXPORT_DIR", "")

	// No WithSiteAddress -> Start falls back to PORT env, binding 0.0.0.0:PORT.
	site := NewSite(WithSiteContentDir(dir))
	go func() { _ = site.Start() }()
	waitForServer(t, "127.0.0.1:"+port)

	resp, err := http.Get("http://127.0.0.1:" + port + "/")
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	io.Copy(io.Discard, resp.Body)
	resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("code = %d", resp.StatusCode)
	}
}

func TestServeLiveIntegration(t *testing.T) {
	dir := t.TempDir()
	writeExportFile(t, dir, "index.md", "# Home")
	writeExportFile(t, dir, "about.md", "# About")

	addr := freeAddr(t)
	site := NewSite(WithSiteContentDir(dir))
	go func() { _ = site.Serve(addr, true) }()
	waitForServer(t, addr)

	base := "http://" + addr
	resp, err := http.Get(base + "/about")
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	body, _ := io.ReadAll(resp.Body)
	resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("about code = %d", resp.StatusCode)
	}
	// The live reload client script should be injected into HTML pages.
	if !strings.Contains(string(body), liveReloadPath) {
		t.Fatalf("expected live reload snippet injected")
	}

	// Hit the SSE endpoint with a short client timeout: the handler streams until
	// the request context is cancelled, then returns cleanly.
	sseClient := &http.Client{Timeout: 600 * time.Millisecond}
	if sseResp, sseErr := sseClient.Get(base + liveReloadPath); sseErr == nil {
		io.Copy(io.Discard, sseResp.Body)
		sseResp.Body.Close()
	}

	// Adding a markdown file should be picked up by the file watcher and served
	// without a restart (exercises the live rebuild + broadcast path).
	writeExportFile(t, dir, "fresh.md", "# Fresh")
	deadline := time.Now().Add(5 * time.Second)
	served := false
	for time.Now().Before(deadline) {
		r, err := http.Get(base + "/fresh")
		if err == nil {
			io.Copy(io.Discard, r.Body)
			r.Body.Close()
			if r.StatusCode == http.StatusOK {
				served = true
				break
			}
		}
		time.Sleep(100 * time.Millisecond)
	}
	if !served {
		t.Fatal("expected live server to serve a newly added page after rebuild")
	}
}

func TestServeEmptyAddrDefaults(t *testing.T) {
	// Serve("", false) should default to :8080 and then fail to build (no content),
	// returning the build error rather than panicking.
	dir := t.TempDir()
	site := NewSite(WithSiteContentDir(dir))
	if err := site.Serve("", false); err == nil {
		t.Fatal("expected build error for empty content dir")
	}
}

func TestStartExportsWhenExportDirSet(t *testing.T) {
	dir := t.TempDir()
	writeExportFile(t, dir, "index.md", "# Home")
	out := t.TempDir()

	site := NewSite(WithSiteContentDir(dir), WithSiteExportDir(out))
	if err := site.Start(); err != nil {
		t.Fatalf("start/export: %v", err)
	}
	if _, err := os.Stat(filepath.Join(out, "index.html")); err != nil {
		t.Fatalf("expected exported index.html: %v", err)
	}
}

func TestStartServesWhenNoExportDir(t *testing.T) {
	dir := t.TempDir()
	writeExportFile(t, dir, "index.md", "# Home")

	addr := freeAddr(t)
	port := addr[strings.LastIndex(addr, ":")+1:]
	t.Setenv("PORT", port)
	t.Setenv("EXPORT_DIR", "")

	site := NewSite(WithSiteContentDir(dir), WithSiteAddress("127.0.0.1:"+port))
	go func() { _ = site.Start() }()
	waitForServer(t, "127.0.0.1:"+port)

	resp, err := http.Get("http://127.0.0.1:" + port + "/")
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	io.Copy(io.Discard, resp.Body)
	resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("code = %d", resp.StatusCode)
	}
}

func TestNilSiteMethods(t *testing.T) {
	var s *Site
	if err := s.Start(); err == nil {
		t.Error("expected error from nil Start")
	}
	if err := s.Serve(":0", false); err == nil {
		t.Error("expected error from nil Serve")
	}
	if err := s.Export("out"); err == nil {
		t.Error("expected error from nil Export")
	}
}
