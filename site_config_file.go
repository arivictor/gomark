package gomark

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// FileConfig is the declarative YAML configuration schema (gomark.yaml). It is
// an optional alternative to the WithSite* options: the `gomark` CLI loads it for
// both `build` and `serve`. Custom layouts and CSS are intentionally not
// configurable — every site uses the built-in theme.
//
// Resolution precedence, highest first:
//
//	CLI flag > environment variable > gomark.yaml > built-in default
type FileConfig struct {
	Title      string          `yaml:"title"`
	URL        string          `yaml:"url"`
	Lang       string          `yaml:"lang"`
	ThemeColor string          `yaml:"theme_color"`
	Footer     string          `yaml:"footer"`
	Favicon    string          `yaml:"favicon"`
	Logo       LogoConfig      `yaml:"logo"`
	SEO        SEOConfig       `yaml:"seo"`
	Build      BuildConfig     `yaml:"build"`
	Nav        []ConfigLink    `yaml:"nav"`
	Social     []ConfigLink    `yaml:"social"`
	Analytics  AnalyticsConfig `yaml:"analytics"`
}

// LogoConfig holds per-theme brand logo URLs.
type LogoConfig struct {
	Light string `yaml:"light"`
	Dark  string `yaml:"dark"`
}

// SEOConfig holds the site-wide SEO defaults (Open Graph, Twitter cards, …).
type SEOConfig struct {
	Description    string `yaml:"description"`
	OGImage        string `yaml:"og_image"`
	TwitterImage   string `yaml:"twitter_image"`
	TwitterSite    string `yaml:"twitter_site"`
	TwitterCreator string `yaml:"twitter_creator"`
	ImageAlt       string `yaml:"image_alt"`
}

// BuildConfig holds build/render behaviour. Pointer bools distinguish "unset"
// (use the default) from an explicit true/false.
type BuildConfig struct {
	ContentDir   string `yaml:"content_dir"`
	OutputDir    string `yaml:"output_dir"`
	PublicDir    string `yaml:"public_dir"`
	SidebarDepth int    `yaml:"sidebar_depth"`
	Runner       *bool  `yaml:"runner"`
	Sitemap      *bool  `yaml:"sitemap"`
	Robots       *bool  `yaml:"robots"`
}

// LoadConfigFile reads and parses a gomark.yaml file. Unknown keys (usually a
// typo such as `tittle:`) are reported as warnings rather than silently ignored,
// then the file is parsed leniently so the remaining valid keys still apply.
func LoadConfigFile(path string) (*FileConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var c FileConfig
	dec := yaml.NewDecoder(bytes.NewReader(data))
	dec.KnownFields(true)
	switch err := dec.Decode(&c); {
	case err == nil:
		return &c, nil
	case errors.Is(err, io.EOF):
		// An empty config file is valid: fall back to defaults.
		return &FileConfig{}, nil
	default:
		// KnownFields raises a *yaml.TypeError for both unknown keys and genuine
		// type mismatches. Only the former (a likely typo) is downgraded to a
		// warning; a real type error (e.g. a string where a number is expected) is
		// returned so the misconfiguration is not silently swallowed.
		var typeErr *yaml.TypeError
		if !errors.As(err, &typeErr) || !allUnknownFieldErrors(typeErr) {
			return nil, fmt.Errorf("parse %s: %w", path, err)
		}
		for _, msg := range typeErr.Errors {
			log.Printf("%s: %s (ignored)", filepath.Base(path), msg)
		}
		// Re-parse leniently so the remaining valid keys still apply.
		var lenient FileConfig
		if err := yaml.Unmarshal(data, &lenient); err != nil {
			return nil, fmt.Errorf("parse %s: %w", path, err)
		}
		return &lenient, nil
	}
}

// allUnknownFieldErrors reports whether every error in a yaml.TypeError is an
// unknown-field error (as opposed to a type mismatch). yaml.v3 phrases unknown
// keys as "field <name> not found in type ...".
func allUnknownFieldErrors(typeErr *yaml.TypeError) bool {
	for _, msg := range typeErr.Errors {
		if !strings.Contains(msg, "not found in type") {
			return false
		}
	}
	return len(typeErr.Errors) > 0
}

// DiscoverConfigFile returns the path to the first gomark.yaml (or gomark.yml)
// found in the given directories, or "" if none exists.
func DiscoverConfigFile(dirs ...string) string {
	names := []string{"gomark.yaml", "gomark.yml"}
	for _, dir := range dirs {
		for _, name := range names {
			candidate := filepath.Join(dir, name)
			if info, err := os.Stat(candidate); err == nil && !info.IsDir() {
				return candidate
			}
		}
	}
	return ""
}

// Options converts the file config into SiteOptions. Only fields that are set
// produce an option, so a FileConfig overrides defaults but leaves room for
// higher-precedence options (env, CLI flags) appended after it.
func (c *FileConfig) Options() []SiteOption {
	if c == nil {
		return nil
	}

	var opts []SiteOption
	add := func(o SiteOption) { opts = append(opts, o) }

	if c.Title != "" {
		add(WithSiteTitle(c.Title))
	}
	if c.URL != "" {
		add(WithSiteURL(c.URL))
	}
	if c.Lang != "" {
		add(WithSiteLang(c.Lang))
	}
	if c.ThemeColor != "" {
		add(WithSiteThemeColor(c.ThemeColor))
	}
	if c.Footer != "" {
		add(WithSiteFooter(c.Footer))
	}
	if c.Favicon != "" {
		add(WithSiteFavicon(c.Favicon))
	}
	if c.Logo.Light != "" {
		add(WithSiteLogoLight(c.Logo.Light))
	}
	if c.Logo.Dark != "" {
		add(WithSiteLogoDark(c.Logo.Dark))
	}
	if c.SEO.Description != "" {
		add(WithSiteDescription(c.SEO.Description))
	}
	if c.SEO.OGImage != "" {
		add(WithSiteOGImage(c.SEO.OGImage))
	}
	if c.SEO.TwitterImage != "" {
		add(WithSiteTwitterImage(c.SEO.TwitterImage))
	}
	if c.SEO.TwitterSite != "" {
		add(WithSiteTwitterSite(c.SEO.TwitterSite))
	}
	if c.SEO.TwitterCreator != "" {
		add(WithSiteTwitterCreator(c.SEO.TwitterCreator))
	}
	if c.SEO.ImageAlt != "" {
		add(WithSiteImageAlt(c.SEO.ImageAlt))
	}
	if c.Build.ContentDir != "" {
		add(WithSiteContentDir(c.Build.ContentDir))
	}
	if c.Build.OutputDir != "" {
		add(WithSiteExportDir(c.Build.OutputDir))
	}
	if c.Build.PublicDir != "" {
		add(WithSitePublicDir(c.Build.PublicDir))
	}
	if c.Build.SidebarDepth > 0 {
		add(WithSiteSidebarDepth(c.Build.SidebarDepth))
	}
	if c.Build.Runner != nil {
		add(WithSiteRunnerEnabled(*c.Build.Runner))
	}
	if c.Build.Sitemap != nil {
		add(WithSiteSitemapEnabled(*c.Build.Sitemap))
	}
	if c.Build.Robots != nil {
		add(WithSiteRobotsEnabled(*c.Build.Robots))
	}
	if len(c.Nav) > 0 {
		add(WithSiteNavLinks(c.Nav...))
	}
	if len(c.Social) > 0 {
		add(WithSiteSocialLinks(c.Social...))
	}
	if c.Analytics.Provider != "" && c.Analytics.ID != "" {
		add(WithSiteAnalytics(c.Analytics.Provider, c.Analytics.ID))
	}

	return opts
}
