---
title: Customization
description: Override GoMark templates and public assets while keeping the built-in defaults as a fallback.
order: 4
---

# Customization

GoMark ships with embedded templates and public assets, so you get a presentable site with zero frontend setup. When you're ready to make it yours, point the app at your own templates and assets — and anything you don't override keeps falling back to the built-in defaults.

## Custom templates

The simplest override is a directory holding `layout.html` and the page templates GoMark expects.

```go:title="cmd/site/main.go"
s := site.NewSite(
	site.WithSiteContentDir("cmd/site/content"),
	site.WithSiteTemplatesDir("templates"),
	site.WithSiteMode(site.PreRender),
)
```

If you need explicit paths instead of a directory convention, use `LayoutPath` and `TemplateGlob`.

```go:title="cmd/site/main.go"
s := site.NewSite(
	site.WithSiteContentDir("cmd/site/content"),
	site.WithSiteLayoutPath("templates/layout.html"),
	site.WithSiteTemplateGlob("templates/*.html"),
)
```

## Custom public assets

Set `PublicDir` to serve your own favicons, OG images, or any additional static files.

```go:title="cmd/site/main.go"
s := site.NewSite(
	site.WithSiteContentDir("cmd/site/content"),
	site.WithSitePublicDir("public"),
)
```

If `PublicDir` is empty, GoMark serves embedded defaults baked into the `site` package.

## What embedded defaults cover

Out of the box, GoMark serves:

- `favicon.ico`
- PNG favicon variants
- Apple touch icon
- Default OG images

That's why you can launch a polished docs site before you've designed a single piece of branding.