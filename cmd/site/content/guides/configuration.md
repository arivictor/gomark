---
title: Configuration
description: Configure GoMark with App fields, environment-driven behavior, and feature toggles.
---

# Configuration

`gomark.App` is the single configuration surface for the site server. Every knob lives in one struct — set the fields you need, leave the rest to sensible defaults.

## Core fields

```go:title="main.go"
app := gomark.App{
	Title:      "My Docs",
	Logo:       "/logo.svg",
	ContentDir: "content",
	SiteURL:    "https://docs.example.com",
	Mode:       gomark.PreRender,
}
```

## App fields

Every field you can set on `gomark.App`:

- `Title` — site name used in layout and metadata
- `Logo` — optional logo URL shown in the header
- `ContentDir` — markdown content root, default `content`
- `TemplatesDir` — directory containing `layout.html` and page templates
- `LayoutPath` — explicit path to a layout template
- `TemplateGlob` — explicit glob for page templates
- `PublicDir` — static asset directory that overrides embedded defaults
- `SidebarDepth` — max sidebar depth, default `2`
- `SiteURL` — base URL used for sitemap and canonical URLs
- `Mode` — `gomark.LiveRender` or `gomark.PreRender`
- `PlaygroundEnabled` — enables runner UI for compatible Go code fences
- `PlaygroundRunnerURL` — runner base URL for playground execution
- `PlaygroundRunnerAuthMode` — auth mode sent to the runner client
- `PlaygroundRunnerAuthToken` — auth token sent to the runner client

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
app := gomark.App{
	ContentDir: "content",
	Mode:       gomark.LiveRender,
}
```

For deployment:

```go:title="main.go"
app := gomark.App{
	Title:      "My Docs",
	ContentDir: "content",
	SiteURL:    "https://docs.example.com",
	Mode:       gomark.PreRender,
}
```