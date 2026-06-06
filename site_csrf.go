package gomark

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"net/http"
	"net/url"
	"strings"
)

const (
	csrfCookieName = "gomark_csrf"
	csrfHeaderName = "X-CSRF-Token"
)

// withCSRFToken attaches the current browser token to the render payload and
// ensures the session cookie exists for future POSTs.
func withCSRFToken(w http.ResponseWriter, r *http.Request, data PageData) PageData {
	data.CSRFToken = ensureCSRFToken(w, r)
	return data
}

// ensureCSRFToken returns the current CSRF token, creating a new session token
// cookie when one is not already present.
func ensureCSRFToken(w http.ResponseWriter, r *http.Request) string {
	if r != nil {
		if token := csrfTokenFromRequest(r); token != "" {
			return token
		}
	}

	token, err := generateCSRFToken()
	if err != nil {
		return ""
	}

	if w != nil {
		http.SetCookie(w, &http.Cookie{
			Name:     csrfCookieName,
			Value:    token,
			Path:     "/",
			SameSite: http.SameSiteLaxMode,
			Secure:   isHTTPSRequest(r),
		})
	}

	return token
}

// CSRFProtectionMiddleware blocks unsafe browser requests unless the request
// presents the matching token and comes from the same origin as the site.
func CSRFProtectionMiddleware(siteURL string) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if isSafeMethod(r.Method) || (sameOriginRequest(r, siteURL) && validCSRFRequest(r)) {
				next.ServeHTTP(w, r)
				return
			}

			http.Error(w, "forbidden", http.StatusForbidden)
		})
	}
}

func validCSRFRequest(r *http.Request) bool {
	if r == nil {
		return false
	}

	cookieToken := csrfTokenFromRequest(r)
	headerToken := strings.TrimSpace(r.Header.Get(csrfHeaderName))
	if cookieToken == "" || headerToken == "" {
		return false
	}

	return subtle.ConstantTimeCompare([]byte(cookieToken), []byte(headerToken)) == 1
}

func csrfTokenFromRequest(r *http.Request) string {
	if r == nil {
		return ""
	}

	cookie, err := r.Cookie(csrfCookieName)
	if err != nil {
		return ""
	}

	return strings.TrimSpace(cookie.Value)
}

func generateCSRFToken() (string, error) {
	seed := make([]byte, 32)
	if _, err := rand.Read(seed); err != nil {
		return "", err
	}

	return base64.RawURLEncoding.EncodeToString(seed), nil
}

func sameOriginRequest(r *http.Request, siteURL string) bool {
	if r == nil {
		return false
	}

	allowed := normalizeBaseURL(requestBaseURL(r, siteURL))
	if allowed == "" {
		return false
	}

	if origin := strings.TrimSpace(r.Header.Get("Origin")); origin != "" {
		return sameOriginURL(origin, allowed)
	}

	if referer := strings.TrimSpace(r.Header.Get("Referer")); referer != "" {
		return sameOriginURL(referer, allowed)
	}

	return false
}

func sameOriginURL(rawURL, allowedBase string) bool {
	parsed, err := url.Parse(strings.TrimSpace(rawURL))
	if err != nil || parsed == nil {
		return false
	}

	allowed, err := url.Parse(strings.TrimSpace(allowedBase))
	if err != nil || allowed == nil {
		return false
	}

	return strings.EqualFold(parsed.Scheme, allowed.Scheme) && strings.EqualFold(parsed.Host, allowed.Host)
}

func isSafeMethod(method string) bool {
	switch strings.ToUpper(strings.TrimSpace(method)) {
	case http.MethodGet, http.MethodHead, http.MethodOptions, http.MethodTrace:
		return true
	default:
		return false
	}
}

func isHTTPSRequest(r *http.Request) bool {
	if r == nil {
		return false
	}

	if r.TLS != nil {
		return true
	}

	base := requestBaseURL(r, "")
	return strings.HasPrefix(strings.ToLower(base), "https://")
}
