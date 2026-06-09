package gomark

import (
	"errors"
	"log"
	"net/http"
	"strings"
	"time"
)

type BadRequestError struct {
	Message string
}

func (e *BadRequestError) Error() string {
	return e.Message
}

type HTTPError struct {
	Status  int
	Message string
}

func (e *HTTPError) Error() string {
	return e.Message
}

type HTMLErrorResponder struct {
	Renderer         TemplateRenderer
	TopNav           []NavLink
	SiteName         string
	Lang             string
	ThemeColor       string
	LogoLight        string
	LogoDark         string
	Favicon          string
	SiteURL          string
	OGImagePath      string
	TwitterImagePath string
	TwitterSite      string
	TwitterCreator   string
	ImageAlt         string
	Footer           string
	NavLinks         []ConfigLink
	SocialLinks      []ConfigLink
	Analytics        AnalyticsConfig
	Logger           *log.Logger
}

func (r HTMLErrorResponder) Handle(w http.ResponseWriter, req *http.Request, err error) {
	if r.Logger == nil {
		r.Logger = log.Default()
	}
	if r.Renderer == nil {
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	status := http.StatusInternalServerError
	title := "Internal Server Error"
	description := "Something went wrong while rendering this page."

	var httpErr *HTTPError
	if errors.As(err, &httpErr) {
		status = httpErr.Status
		title = http.StatusText(status)
		description = httpErr.Message
	}

	var badReq *BadRequestError
	if errors.As(err, &badReq) {
		status = http.StatusBadRequest
		title = "Bad Request"
		description = badReq.Message
	}

	if status >= http.StatusInternalServerError {
		r.Logger.Printf("internal error on %s %s: %v", req.Method, req.URL.Path, err)
	}

	if description == "" {
		description = "The requested page could not be served."
	}

	renderErr := r.Renderer.RenderStatus(w, status, "error", withCSRFToken(w, req, PageData{
		StatusCode:   status,
		Title:        title,
		Description:  description,
		SiteName:     firstNonEmpty(r.SiteName, defaultSiteName),
		Lang:         firstNonEmpty(strings.TrimSpace(r.Lang), "en"),
		ThemeColor:   strings.TrimSpace(r.ThemeColor),
		LogoLightURL: strings.TrimSpace(r.LogoLight),
		LogoDarkURL:  strings.TrimSpace(r.LogoDark),
		FaviconURL:   strings.TrimSpace(r.Favicon),
		CanonicalURL: joinAbsoluteURL(
			requestBaseURL(req, r.SiteURL),
			req.URL.Path,
		),
		OGImageURL:      joinAbsoluteURL(requestBaseURL(req, r.SiteURL), firstNonEmpty(strings.TrimSpace(r.OGImagePath), defaultOGImagePath)),
		TwitterImageURL: joinAbsoluteURL(requestBaseURL(req, r.SiteURL), firstNonEmpty(strings.TrimSpace(r.TwitterImagePath), defaultTwitterImagePath)),
		TwitterSite:     strings.TrimSpace(r.TwitterSite),
		TwitterCreator:  strings.TrimSpace(r.TwitterCreator),
		ImageAlt:        firstNonEmpty(strings.TrimSpace(r.ImageAlt), firstNonEmpty(r.SiteName, defaultSiteName)),
		FooterText:      strings.TrimSpace(r.Footer),
		NavLinks:        r.NavLinks,
		SocialLinks:     r.SocialLinks,
		Analytics:       r.Analytics,
		Robots:          "noindex,nofollow",
		TopNav:          r.TopNav,
		CurrentPath:     req.URL.Path,
		Time:            time.Now().UTC().Format(time.RFC3339),
	}))
	if renderErr != nil {
		r.Logger.Printf("error rendering error page: %v", renderErr)
		http.Error(w, http.StatusText(status), status)
	}
}
