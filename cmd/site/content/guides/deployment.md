---
title: Deployment
description: Prepare a GoMark site for production with prerendering, a stable SiteURL, and deployable asset defaults.
---

# Deployment

For production, the main goals are predictable output, correct canonical URLs, and safe runner settings if you enable playground execution.

## Recommended production app

```go:title="main.go"
app := gomark.App{
	Title:      "My Docs",
	ContentDir: "content",
	SiteURL:    "https://docs.example.com",
	Mode:       gomark.PreRender,
}
```

## Why `SiteURL` matters

Set `SiteURL` so GoMark can generate:

- canonical URLs
- `sitemap.xml`
- `robots.txt` sitemap links
- OG image URLs

## Deployment checklist

1. Use `gomark.PreRender`.
2. Set `SiteURL` to the public origin.
3. Provide `PublicDir` if you need custom branding.
4. Provide custom templates only when you need them.
5. Keep the runner behind auth if playground execution is enabled.

## Publishing the module

If you publish versions of GoMark itself, tag releases in semver format.

```terminal
git tag v0.1.0
git push origin v0.1.0
```

Consumers can then install a pinned version with:

```terminal
go get github.com/arivictor/gomark@v0.1.0
```