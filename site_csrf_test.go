package gomark

import (
	"html/template"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestEnsureCSRFTokenSetsSessionCookie(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "http://example.com/", nil)
	rec := httptest.NewRecorder()

	token := ensureCSRFToken(rec, req)
	if token == "" {
		t.Fatalf("expected csrf token to be generated")
	}

	res := rec.Result()
	defer res.Body.Close()

	cookies := res.Cookies()
	if len(cookies) != 1 {
		t.Fatalf("expected one csrf cookie, got %d", len(cookies))
	}
	if cookies[0].Name != csrfCookieName {
		t.Fatalf("expected csrf cookie %q, got %q", csrfCookieName, cookies[0].Name)
	}
	if cookies[0].Value != token {
		t.Fatalf("expected csrf cookie to match token, got %q want %q", cookies[0].Value, token)
	}
}

func TestLayoutIncludesCSRFTokenMetaTag(t *testing.T) {
	renderer, err := NewFileTemplateRenderer("", "")
	if err != nil {
		t.Fatalf("new embedded template renderer: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "http://example.com/", nil)
	rec := httptest.NewRecorder()

	data := withCSRFToken(rec, req, PageData{
		Title:    "Home",
		BodyHTML: template.HTML("<p>content</p>"),
	})
	if data.CSRFToken == "" {
		t.Fatalf("expected csrf token in page data")
	}

	if err := renderer.Render(rec, "markdown", data); err != nil {
		t.Fatalf("render markdown with csrf token: %v", err)
	}

	if !strings.Contains(rec.Body.String(), `meta name="csrf-token"`) {
		t.Fatalf("expected csrf meta tag in rendered html: %s", rec.Body.String())
	}
}

func TestCSRFProtectionMiddlewareAllowsSameOriginRequest(t *testing.T) {
	protected := CSRFProtectionMiddleware("http://example.com")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))

	req := httptest.NewRequest(http.MethodPost, "http://example.com/submit", strings.NewReader(`{"code":"package main"}`))
	req.Header.Set("Origin", "http://example.com")
	req.Header.Set(csrfHeaderName, "secret-token")
	req.AddCookie(&http.Cookie{Name: csrfCookieName, Value: "secret-token"})

	rec := httptest.NewRecorder()
	protected.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("expected request to be allowed, got status %d", rec.Code)
	}
}

func TestCSRFProtectionMiddlewareRejectsMissingOrInvalidRequests(t *testing.T) {
	tests := []struct {
		name   string
		req    func() *http.Request
		status int
	}{
		{
			name: "missing token",
			req: func() *http.Request {
				r := httptest.NewRequest(http.MethodPost, "http://example.com/submit", nil)
				r.Header.Set("Origin", "http://example.com")
				return r
			},
			status: http.StatusForbidden,
		},
		{
			name: "wrong token",
			req: func() *http.Request {
				r := httptest.NewRequest(http.MethodPost, "http://example.com/submit", nil)
				r.Header.Set("Origin", "http://example.com")
				r.Header.Set(csrfHeaderName, "wrong-token")
				r.AddCookie(&http.Cookie{Name: csrfCookieName, Value: "secret-token"})
				return r
			},
			status: http.StatusForbidden,
		},
		{
			name: "cross-site origin",
			req: func() *http.Request {
				r := httptest.NewRequest(http.MethodPost, "http://example.com/submit", nil)
				r.Header.Set("Origin", "http://evil.example")
				r.Header.Set(csrfHeaderName, "secret-token")
				r.AddCookie(&http.Cookie{Name: csrfCookieName, Value: "secret-token"})
				return r
			},
			status: http.StatusForbidden,
		},
	}

	protected := CSRFProtectionMiddleware("http://example.com")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			protected.ServeHTTP(rec, tc.req())

			if rec.Code != tc.status {
				t.Fatalf("expected status %d, got %d", tc.status, rec.Code)
			}
		})
	}
}