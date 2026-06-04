---
title: Deployment
description: Prepare a GoMark site for production with prerendering, a stable SiteURL, and deployable asset defaults.
order: 7
---

# Deployment

GoMark is production-ready out of the box, but there are a few things to check before you go live. This guide walks through the recommended configuration for deployment, and how to build the binaries that ship your site.

## Recommended production app

```go:title="cmd/site/main.go"
s := gm.NewSite(
	gm.WithSiteTitle("My Docs"),
	gm.WithSiteContentDir("cmd/site/content"),
	gm.WithSiteURL("https://docs.example.com"),
	gm.WithSiteMode(gm.PreRender),
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

1. Use `gomark.PreRender` for faster response times and to catch content issues at startup
2. Set `SiteURL` to your public origin
3. Provide `PublicDir` if you need custom branding or assets beyond the embedded defaults
4. Add custom templates only when you need them
5. Keep the runner behind auth if runner execution is enabled

## Building the binaries

GoMark deploys as two self-contained binaries. Templates and public assets are embedded, so the compiled output is all you need to ship.

```terminal
go build -o bin/site cmd/site/main.go
go build -o bin/runner cmd/runner/main.go
```

The repository also includes a multi-stage `Dockerfile` that builds both binaries and serves the site by default:

```terminal
docker build -t <my-image> .
docker run -p 8080:8080 <my-image>
docker run -p 8080:8080 <my-image> /app/runner
```