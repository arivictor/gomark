package gomark

import (
	"fmt"
	"os"
	"path/filepath"

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
	SidebarDepth int    `yaml:"sidebar_depth"`
	Runner       *bool  `yaml:"runner"`
	Sitemap      *bool  `yaml:"sitemap"`
	Robots       *bool  `yaml:"robots"`
}

// LoadConfigFile reads and parses a gomark.yaml file.
func LoadConfigFile(path string) (*FileConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var c FileConfig
	if err := yaml.Unmarshal(data, &c); err != nil {
		return nil, fmt.Errorf("parse %s: %w", path, err)
	}
	return &c, nil
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
