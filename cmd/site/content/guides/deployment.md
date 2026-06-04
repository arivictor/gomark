---
title: Deployment
description: Prepare a GoMark site for production with prerendering, a stable SiteURL, and deployable asset defaults.
---

# Deployment

Shipping to production comes down to three things: predictable output, correct canonical URLs, and a locked-down runner if you've enabled playground execution. GoMark makes all three a matter of setting a couple of fields.

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

`SiteURL` is the one field production really cares about. Set it and GoMark generates correct:

- canonical URLs
- `sitemap.xml`
- `robots.txt` sitemap links
- OG image URLs

## Deployment checklist

Run through this before you go live:

1. Use `gomark.PreRender`
2. Set `SiteURL` to your public origin
3. Provide `PublicDir` if you need custom branding
4. Add custom templates only when you need them
5. Keep the runner behind auth if playground execution is enabled

## Publishing the module

Publishing versions of GoMark itself? Tag releases in semver format.

```terminal
git tag v0.1.0
git push origin v0.1.0
```

Consumers can then install a pinned version with:

```terminal
go get github.com/arivictor/gomark@v0.1.0
```