package gomark

import (
	"net/http"
	"strings"
)

// buildRoutePath turns a content slug into an absolute route, so "guide/about"
// -> "/guide/about" and "" -> "/".
func buildRoutePath(serviceSlug string) string {
	slug := strings.Trim(strings.TrimSpace(serviceSlug), "/")
	if slug == "" {
		return "/"
	}
	return "/" + slug
}

func pageTitleFromSlug(slug string) string {
	parts := strings.Split(strings.TrimSpace(slug), "/")
	if len(parts) == 0 {
		return "Content"
	}
	last := parts[len(parts)-1]
	if last == "index" && len(parts) > 1 {
		last = parts[len(parts)-2]
	}
	if last == "" {
		last = "Content"
	}

	words := strings.Fields(strings.ReplaceAll(last, "-", " "))
	for i, word := range words {
		r := []rune(word)
		if len(r) == 0 {
			continue
		}
		r[0] = []rune(strings.ToUpper(string(r[0])))[0]
		words[i] = string(r)
	}
	if len(words) == 0 {
		return "Content"
	}
	return strings.Join(words, " ")
}

func normalizeBaseURL(raw string) string {
	base := strings.TrimSpace(raw)
	if base == "" {
		return defaultSiteURL
	}
	if !strings.HasPrefix(base, "http://") && !strings.HasPrefix(base, "https://") {
		base = "https://" + base
	}
	return strings.TrimRight(base, "/")
}

func requestBaseURL(r *http.Request, fallback string) string {
	if r == nil {
		return normalizeBaseURL(fallback)
	}

	host := strings.TrimSpace(r.Header.Get("X-Forwarded-Host"))
	if host == "" {
		host = strings.TrimSpace(r.Host)
	}
	if host == "" {
		return normalizeBaseURL(fallback)
	}
	if i := strings.Index(host, ","); i != -1 {
		host = strings.TrimSpace(host[:i])
	}

	proto := strings.TrimSpace(r.Header.Get("X-Forwarded-Proto"))
	if i := strings.Index(proto, ","); i != -1 {
		proto = strings.TrimSpace(proto[:i])
	}
	if proto == "" {
		if r.TLS != nil {
			proto = "https"
		} else {
			proto = "http"
		}
	}

	return proto + "://" + host
}

func joinAbsoluteURL(base, route string) string {
	cleanBase := normalizeBaseURL(base)
	cleanRoute := strings.TrimSpace(route)
	if cleanRoute == "" {
		cleanRoute = "/"
	}
	if !strings.HasPrefix(cleanRoute, "/") {
		cleanRoute = "/" + cleanRoute
	}
	return cleanBase + cleanRoute
}
