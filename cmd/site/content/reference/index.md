---
title: API Reference
description: Reference for the public GoMark site and runner APIs.
---

# API Reference

The complete public surface of GoMark, in one place. Two small APIs — one to build sites, one to run code — cover everything.

## Site API

### `type Site`

Site server configured through constructor options.

### `func NewSite(options ...SiteOption) *Site`

Creates a site instance.

### `func (s *Site) Start() error`

Starts the site server using `WithSiteAddress`, `PORT`, or `:8080`.

### `type SiteOption func(*Site)`

Common options:

- `WithSiteAddress(addr)`
- `WithSiteTitle(title)`
- `WithSiteLogo(url)`
- `WithSiteContentDir(dir)`
- `WithSiteTemplatesDir(dir)`
- `WithSiteLayoutPath(path)`
- `WithSiteTemplateGlob(glob)`
- `WithSitePublicDir(dir)`
- `WithSiteSidebarDepth(depth)`
- `WithSiteURL(url)`
- `WithSiteMode(mode)`
- `WithSiteRunnerEnabled(enabled)`
- `WithSiteRunnerURL(url)`
- `WithSiteRunnerAuth(mode, token)`
- `WithSiteRunner(url, mode, token)`

### `type RenderMode string`

- `gomark.LiveRender`
- `gomark.PreRender`

### `func ParseRenderMode(raw string) RenderMode`

Resolves strings like `production`, `prod`, `development`, and `live` into a render mode.

## Runner API

### `type Runner`

Runner server configured through constructor options.

### `func NewRunner(options ...Option) *Runner`

Creates a runner instance.

### `func (r *Runner) Start() error`

Starts the Go runner server.

### `func WithPort(port string) Option`

Sets the listen port.

### `func WithAddress(addr string) Option`

Sets the full listen address.

### `func WithAuth(mode AuthMode, token string) Option`

Sets runner auth mode and token.

### `type AuthMode string`

- `gomark.AuthModeBearerStatic`
- `gomark.AuthModeNone`

## Common examples

Copy, paste, ship.

### Site

```go:title="main.go"
site := gomark.NewSite(
	gomark.WithSiteTitle("My Docs"),
	gomark.WithSiteContentDir("content"),
	gomark.WithSiteMode(gomark.PreRender),
)

if err := site.Start(); err != nil {
	log.Fatal(err)
}
```

### Runner

```go:title="main.go"
runner := gomark.NewRunner(
	gomark.WithPort("8081"),
	gomark.WithAuth(gomark.AuthModeBearerStatic, "secret"),
)

if err := runner.Start(); err != nil {
	log.Fatal(err)
}
```