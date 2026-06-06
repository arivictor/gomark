package gomark

import (
	"crypto/tls"
	"net/http/httptest"
	"testing"
)

func TestBuildRoutePath(t *testing.T) {
	cases := map[string]string{
		"":           "/",
		"  ":         "/",
		"/go/about/": "/go/about",
		"guide":      "/guide",
		"a/b/c":      "/a/b/c",
	}
	for in, want := range cases {
		if got := buildRoutePath(in); got != want {
			t.Errorf("buildRoutePath(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestPageTitleFromSlug(t *testing.T) {
	cases := map[string]string{
		"getting-started": "Getting Started",
		"guides/install":  "Install",
		"guides/index":    "Guides", // index uses parent folder name
		"":                "Content",
		"index":           "Index",
	}
	for in, want := range cases {
		if got := pageTitleFromSlug(in); got != want {
			t.Errorf("pageTitleFromSlug(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestNormalizeBaseURL(t *testing.T) {
	if got := normalizeBaseURL(""); got != defaultSiteURL {
		t.Errorf("empty = %q", got)
	}
	if got := normalizeBaseURL("example.com/"); got != "https://example.com" {
		t.Errorf("bare host = %q", got)
	}
	if got := normalizeBaseURL("http://example.com/path/"); got != "http://example.com/path" {
		t.Errorf("http = %q", got)
	}
}

func TestJoinAbsoluteURL(t *testing.T) {
	if got := joinAbsoluteURL("https://example.com", "/about"); got != "https://example.com/about" {
		t.Errorf("got %q", got)
	}
	if got := joinAbsoluteURL("https://example.com", ""); got != "https://example.com/" {
		t.Errorf("empty route got %q", got)
	}
	if got := joinAbsoluteURL("https://example.com", "about"); got != "https://example.com/about" {
		t.Errorf("relative route got %q", got)
	}
}

func TestRequestBaseURL(t *testing.T) {
	if got := requestBaseURL(nil, "example.com"); got != "https://example.com" {
		t.Errorf("nil request = %q", got)
	}

	// Host header (http).
	r := httptest.NewRequest("GET", "http://host.test/x", nil)
	r.Host = "host.test"
	if got := requestBaseURL(r, ""); got != "http://host.test" {
		t.Errorf("host = %q", got)
	}

	// TLS request -> https.
	tlsReq := httptest.NewRequest("GET", "https://secure.test/x", nil)
	tlsReq.Host = "secure.test"
	tlsReq.TLS = &tls.ConnectionState{}
	if got := requestBaseURL(tlsReq, ""); got != "https://secure.test" {
		t.Errorf("tls = %q", got)
	}

	// X-Forwarded-Host and X-Forwarded-Proto with comma-lists.
	fwd := httptest.NewRequest("GET", "http://internal/x", nil)
	fwd.Header.Set("X-Forwarded-Host", "public.test, internal")
	fwd.Header.Set("X-Forwarded-Proto", "https, http")
	if got := requestBaseURL(fwd, ""); got != "https://public.test" {
		t.Errorf("forwarded = %q", got)
	}

	// No host at all -> fallback.
	empty := httptest.NewRequest("GET", "http://x/", nil)
	empty.Host = ""
	if got := requestBaseURL(empty, "fallback.test"); got != "https://fallback.test" {
		t.Errorf("fallback = %q", got)
	}
}

func TestNavNodeID(t *testing.T) {
	if navNodeID(nil) != "nav-root" {
		t.Error("nil parts -> nav-root")
	}
	if got := navNodeID([]string{"guides", "install"}); got != "nav-guides-install" {
		t.Errorf("got %q", got)
	}
}

func TestHumanizeSlug(t *testing.T) {
	if humanizeSlug("getting-started") != "Getting Started" {
		t.Errorf("got %q", humanizeSlug("getting-started"))
	}
}
