---
title: Customization
description: Override GoMark templates and public assets while keeping the built-in defaults as a fallback.
---

# Customization

GoMark ships with embedded templates and embedded public assets so you can start without any frontend setup.

When you want more control, point the app at your own template and asset directories.

## Custom templates

The simplest template override is a directory with `layout.html` and the page templates GoMark expects.

```go:title="main.go"
app := gomark.App{
	ContentDir:   "content",
	TemplatesDir: "templates",
	Mode:         gomark.PreRender,
}
```

If you need explicit paths instead of a directory convention, use `LayoutPath` and `TemplateGlob`.

```go:title="main.go"
app := gomark.App{
	ContentDir:   "content",
	LayoutPath:   "templates/layout.html",
	TemplateGlob: "templates/*.html",
}
```

## Custom public assets

Use `PublicDir` when you want your own favicons, OG images, or additional static files.

```go:title="main.go"
app := gomark.App{
	ContentDir: "content",
	PublicDir:  "public",
}
```

If `PublicDir` is empty, GoMark serves embedded defaults from the package.

## What embedded defaults cover

- `favicon.ico`
- PNG favicon variants
- Apple touch icon
- Default OG images

This means you can launch a presentable docs site before you have any custom branding.