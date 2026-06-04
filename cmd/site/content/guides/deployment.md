---
title: Deployment
description: Prepare a GoMark site for production with prerendering, a stable SiteURL, and deployable asset defaults.
---

# Deployment

Shipping to production comes down to three things: predictable output, correct canonical URLs, and a locked-down runner if you've enabled runner execution. GoMark makes all three a matter of setting a couple of fields.

## Recommended production app

```go:title="cmd/site/main.go"
s := site.NewSite(
	site.WithSiteTitle("My Docs"),
	site.WithSiteContentDir("cmd/site/content"),
	site.WithSiteURL("https://docs.example.com"),
	site.WithSiteMode(site.PreRender),
)
```

## Why `SiteURL` matters

`SiteURL` is the one field production really cares about. Set it and GoMark generates correct:

- canonical URLs
- `sitemap.xml`
- `robots.txt` sitemap links
- OG image URLs

## Deployment checklist

Run through this before you go live:

1. Use `site.PreRender`
2. Set `SiteURL` to your public origin
3. Provide `PublicDir` if you need custom branding
4. Add custom templates only when you need them
5. Keep the runner behind auth if runner execution is enabled

## Building the binaries

GoMark deploys as two self-contained binaries. Templates and public assets are embedded, so the compiled output is all you need to ship.

```terminal
go build -o bin/site ./cmd/site
go build -o bin/runner ./cmd/runner
```

The repository also includes a multi-stage `Dockerfile` that builds both binaries and serves the site by default:

```terminal
docker build -t gomark .
docker run -p 8080:8080 gomark
```