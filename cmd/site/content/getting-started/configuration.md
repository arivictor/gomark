---
title: Configuration
description: Configure GoMark with Site options, environment-driven behavior, and feature toggles.
order: 2
---

# Configuration

`gomark.NewSite(...)` with `gomark.With...` options is the single configuration surface for the site server. Set the options you need in `main.go`, and focus on writing content instead of wiring up features.

## Core fields

```go:title="main.go"
s := gomark.NewSite(
	gomark.WithSiteTitle("My Docs"),
	gomark.WithSiteLogo("/logo.svg"),
	gomark.WithSiteContentDir("content"),
	gomark.WithSiteURL("https://docs.example.com"),
	gomark.WithSiteMode(gomark.PreRender),
)
```

## Site options

Common options you can set on `gomark.NewSite(...)`:

- `WithSiteTitle` — site name used in layout and metadata
- `WithSiteLogo` — optional logo URL shown in the header
- `WithSiteContentDir` — markdown content root, default `content`
- `WithSiteTemplatesDir` — directory containing `layout.html` and page templates
- `WithSiteLayoutPath` — explicit path to a layout template
- `WithSiteTemplateGlob` — explicit glob for page templates
- `WithSitePublicDir` — static asset directory that overrides embedded defaults
- `WithSiteSidebarDepth` — max sidebar depth, default `2`
- `WithSiteURL` — base URL used for sitemap and canonical URLs
- `WithSiteMode` — `gomark.LiveRender` or `gomark.PreRender`
- `WithSiteRunnerEnabled` — toggles the in-browser Go runner UI (on by default)

## Render modes

GoMark renders one of two ways. Pick the one that matches what you're doing.

### `gomark.LiveRender`

- Reads markdown from disk on each request
- Best for local development
- Reflects file edits without restart

### `gomark.PreRender`

- Builds markdown output up front at startup
- Best for production or stable content
- Fails fast on content issues during boot

GoMark also recognizes common environment aliases such as `prod`, `production`, `live`, and `development` when resolving render mode.

## Recommended defaults

Two starting points that cover most projects.

For local work:

```go:title="main.go"
s := gomark.NewSite(
	gomark.WithSiteContentDir("content"),
	gomark.WithSiteMode(gomark.LiveRender),
)
```

For deployment:

```go:title="main.go"
s := gomark.NewSite(
	gomark.WithSiteTitle("My Docs"),
	gomark.WithSiteContentDir("content"),
	gomark.WithSiteURL("https://docs.example.com"),
	gomark.WithSiteMode(gomark.PreRender),
)
```
