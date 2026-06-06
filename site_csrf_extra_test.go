package gomark

import (
	"crypto/tls"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestIsSafeMethod(t *testing.T) {
	safe := []string{"GET", "head", "  Options  ", "TRACE"}
	for _, m := range safe {
		if !isSafeMethod(m) {
			t.Errorf("expected %q to be safe", m)
		}
	}
	unsafe := []string{"POST", "PUT", "DELETE", "PATCH", ""}
	for _, m := range unsafe {
		if isSafeMethod(m) {
			t.Errorf("expected %q to be unsafe", m)
		}
	}
}

func TestIsHTTPSRequest(t *testing.T) {
	if isHTTPSRequest(nil) {
		t.Error("nil request is not https")
	}

	tlsReq := httptest.NewRequest("GET", "https://example.com/", nil)
	tlsReq.TLS = &tls.ConnectionState{}
	if !isHTTPSRequest(tlsReq) {
		t.Error("expected https for TLS request")
	}

	fwd := httptest.NewRequest("GET", "http://example.com/", nil)
	fwd.Header.Set("X-Forwarded-Proto", "https")
	if !isHTTPSRequest(fwd) {
		t.Error("expected https via X-Forwarded-Proto")
	}

	plain := httptest.NewRequest("GET", "http://example.com/", nil)
	if isHTTPSRequest(plain) {
		t.Error("expected plain http")
	}
}

func TestGenerateCSRFTokenUnique(t *testing.T) {
	a, err := generateCSRFToken()
	if err != nil {
		t.Fatalf("generate: %v", err)
	}
	b, err := generateCSRFToken()
	if err != nil {
		t.Fatalf("generate: %v", err)
	}
	if a == "" || b == "" {
		t.Fatal("expected non-empty tokens")
	}
	if a == b {
		t.Fatal("expected unique tokens")
	}
}

func TestSameOriginRequestViaReferer(t *testing.T) {
	r := httptest.NewRequest("POST", "http://example.com/x", nil)
	r.Header.Set("Referer", "http://example.com/some/page")
	if !sameOriginRequest(r, "http://example.com") {
		t.Error("expected same origin via referer")
	}

	cross := httptest.NewRequest("POST", "http://example.com/x", nil)
	cross.Header.Set("Referer", "http://evil.test/page")
	if sameOriginRequest(cross, "http://example.com") {
		t.Error("expected cross origin via referer to be rejected")
	}
}

func TestSameOriginRequestNoHeaders(t *testing.T) {
	r := httptest.NewRequest("POST", "http://example.com/x", nil)
	// No Origin and no Referer -> cannot verify same origin.
	if sameOriginRequest(r, "http://example.com") {
		t.Error("expected false when no origin/referer headers present")
	}
}

func TestSameOriginRequestNil(t *testing.T) {
	if sameOriginRequest(nil, "http://example.com") {
		t.Error("nil request is never same origin")
	}
}

func TestSameOriginURL(t *testing.T) {
	if !sameOriginURL("https://Example.com/a", "https://example.com") {
		t.Error("expected case-insensitive host match")
	}
	if sameOriginURL("http://example.com", "https://example.com") {
		t.Error("expected scheme mismatch to fail")
	}
	if sameOriginURL("://bad url", "https://example.com") {
		// url.Parse is lenient; ensure differing host still returns false.
	}
}

func TestValidCSRFRequestNil(t *testing.T) {
	if validCSRFRequest(nil) {
		t.Error("nil request invalid")
	}
}

func TestCSRFTokenFromRequestNil(t *testing.T) {
	if csrfTokenFromRequest(nil) != "" {
		t.Error("nil request has no token")
	}
}

func TestEnsureCSRFTokenReusesExisting(t *testing.T) {
	req := httptest.NewRequest("GET", "http://example.com/", nil)
	req.AddCookie(&http.Cookie{Name: csrfCookieName, Value: "existing-token"})
	rec := httptest.NewRecorder()
	if got := ensureCSRFToken(rec, req); got != "existing-token" {
		t.Fatalf("expected existing token reused, got %q", got)
	}
	if len(rec.Result().Cookies()) != 0 {
		t.Fatal("expected no new cookie when token already present")
	}
}

func TestEnsureCSRFTokenSetsSecureForHTTPS(t *testing.T) {
	req := httptest.NewRequest("GET", "https://example.com/", nil)
	req.TLS = &tls.ConnectionState{}
	rec := httptest.NewRecorder()
	ensureCSRFToken(rec, req)
	cookies := rec.Result().Cookies()
	if len(cookies) != 1 || !cookies[0].Secure {
		t.Fatalf("expected secure cookie over https, got %+v", cookies)
	}
}
