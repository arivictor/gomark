---
title: API Reference
description: Reference for the public GoMark site and runner APIs.
---

# API Reference

The complete public surface of GoMark, in one place. Two small APIs — one to build sites, one to run code — cover everything.

## Site API

### `type App struct`

Use `App` to configure and run a markdown site.

### `func (a *App) Run(addr string) error`

Starts the site server on the provided address.

### `func Run(addr string, app App) error`

Convenience wrapper for `app.Run(addr)`.

### `type RenderMode string`

- `gomark.LiveRender`
- `gomark.PreRender`

### `func ParseRenderMode(raw string) RenderMode`

Resolves strings like `production`, `prod`, `development`, and `live` into a render mode.

## Runner API

### `func Start(options ...Option) error`

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
app := gomark.App{
	Title:      "My Docs",
	ContentDir: "content",
	Mode:       gomark.PreRender,
}

if err := app.Run(":8080"); err != nil {
	log.Fatal(err)
}
```

### Runner

```go:title="main.go"
if err := gomark.Start(
	gomark.WithPort("8081"),
	gomark.WithAuth(gomark.AuthModeBearerStatic, "secret"),
); err != nil {
	log.Fatal(err)
}
```