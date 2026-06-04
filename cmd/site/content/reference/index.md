---
title: API Reference
description: Reference for the GoMark site and runner package APIs.
---

# API Reference

The configuration surface of GoMark, in one place. Two small package APIs — one to build sites, one to run code — cover everything, with a shared `protocol` package for the types they exchange.

## Packages

- `github.com/arivictor/gomark/site` — the site server
- `github.com/arivictor/gomark/runner` — the runner server
- `github.com/arivictor/gomark/protocol` — shared wire types (`AuthMode`, `RunRequest`, `RunResponse`)

These are public, importable packages. The `cmd/site` and `cmd/runner` binaries in this repository are reference consumers of them.

## Site API

### `type site.Site`

Site server configured through constructor options.

### `func site.NewSite(options ...SiteOption) *Site`

Creates a site instance.

### `func (s *Site) Start() error`

Starts the site server using `WithSiteAddress`, `PORT`, or `:8080`.

### `type site.SiteOption func(*Site)`

Common options:

- `site.WithSiteAddress(addr)`
- `site.WithSiteTitle(title)`
- `site.WithSiteLogo(url)`
- `site.WithSiteContentDir(dir)`
- `site.WithSiteTemplatesDir(dir)`
- `site.WithSiteLayoutPath(path)`
- `site.WithSiteTemplateGlob(glob)`
- `site.WithSitePublicDir(dir)`
- `site.WithSiteSidebarDepth(depth)`
- `site.WithSiteURL(url)`
- `site.WithSiteMode(mode)`
- `site.WithSiteRunnerEnabled(enabled)`
- `site.WithSiteRunnerURL(url)`
- `site.WithSiteRunnerAuth(mode, token)`
- `site.WithSiteRunner(url, mode, token)`

### `type site.RenderMode string`

- `site.LiveRender`
- `site.PreRender`

### `func site.ParseRenderMode(raw string) RenderMode`

Resolves strings like `production`, `prod`, `development`, and `live` into a render mode.

## Runner API

### `type runner.Runner`

Runner server configured through constructor options.

### `func runner.NewRunner(options ...Option) *Runner`

Creates a runner instance.

### `func (r *Runner) Start() error`

Starts the Go runner server.

### `func runner.WithPort(port string) Option`

Sets the listen port.

### `func runner.WithAddress(addr string) Option`

Sets the full listen address.

### `func runner.WithAuth(mode AuthMode, token string) Option`

Sets runner auth mode and token.

### `func runner.WithTimeout(seconds int) Option`

Sets the per-`/run` execution timeout in whole seconds. Defaults to 2 seconds; values `<= 0` are ignored.

## Shared types

### `type AuthMode string`

- `AuthBearerStatic`
- `AuthNone`

### `type RunRequest`

The body sent to the runner's `/run` endpoint (`{"code": "..."}`).

### `type RunResponse`

The runner's reply: `ok`, `output`, `error`, `exitCode`, and `durationMs`.

## Common examples

Copy, paste, ship.

### Site

```go:title="cmd/site/main.go"
s := site.NewSite(
	site.WithSiteTitle("My Docs"),
	site.WithSiteContentDir("cmd/site/content"),
	site.WithSiteMode(site.PreRender),
)

if err := s.Start(); err != nil {
	log.Fatal(err)
}
```

### Runner

```go:title="cmd/runner/main.go"
r := runner.NewRunner(
	runner.WithPort("8081"),
	runner.WithAuth(AuthBearerStatic, "secret"),
	runner.WithTimeout(30),
)

if err := r.Start(); err != nil {
	log.Fatal(err)
}
```
