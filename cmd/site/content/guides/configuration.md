---
title: Configuration
description: Configure GoMark with Site options, environment-driven behavior, and feature toggles.
order: 3
---

# Configuration

`site.NewSite(...)` with `WithSite...` options is the single configuration surface for the site server. Set the options you need in `cmd/site/main.go`, leave the rest to sensible defaults.

## Core fields

```go:title="cmd/site/main.go"
s := site.NewSite(
	site.WithSiteTitle("My Docs"),
	site.WithSiteLogo("/logo.svg"),
	site.WithSiteContentDir("cmd/site/content"),
	site.WithSiteURL("https://docs.example.com"),
	site.WithSiteMode(site.PreRender),
)
```

## Site options

Common options you can set on `site.NewSite(...)`:

- `WithSiteTitle` — site name used in layout and metadata
- `WithSiteLogo` — optional logo URL shown in the header
- `WithSiteContentDir` — markdown content root, default `content`
- `WithSiteTemplatesDir` — directory containing `layout.html` and page templates
- `WithSiteLayoutPath` — explicit path to a layout template
- `WithSiteTemplateGlob` — explicit glob for page templates
- `WithSitePublicDir` — static asset directory that overrides embedded defaults
- `WithSiteSidebarDepth` — max sidebar depth, default `2`
- `WithSiteURL` — base URL used for sitemap and canonical URLs
- `WithSiteMode` — `site.LiveRender` or `site.PreRender`
- `WithSiteRunnerEnabled` — enables runner UI for compatible Go code fences
- `WithSiteRunnerURL` — runner base URL for runner execution
- `WithSiteRunnerAuth` — auth mode and token sent to the runner client
- `WithSiteRunner` — enable + URL + auth in one option

## Render modes

GoMark renders one of two ways. Pick the one that matches what you're doing.

### `site.LiveRender`

- Reads markdown from disk on each request
- Best for local development
- Reflects file edits without restart

### `site.PreRender`

- Builds markdown output up front at startup
- Best for production or stable content
- Fails fast on content issues during boot

GoMark also recognizes common environment aliases such as `prod`, `production`, `live`, and `development` when resolving render mode.

## Recommended defaults

Two starting points that cover most projects.

For local work:

```go:title="cmd/site/main.go"
s := site.NewSite(
	site.WithSiteContentDir("cmd/site/content"),
	site.WithSiteMode(site.LiveRender),
)
```

For deployment:

```go:title="cmd/site/main.go"
s := site.NewSite(
	site.WithSiteTitle("My Docs"),
	site.WithSiteContentDir("cmd/site/content"),
	site.WithSiteURL("https://docs.example.com"),
	site.WithSiteMode(site.PreRender),
)
```
