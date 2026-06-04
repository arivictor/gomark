# gomark

Build a markdown-powered website in Go with batteries included: routing, rendering,
navigation, search, sitemap, robots, and static site assets.

## Install

```bash
go get github.com/arivictor/gomark@latest
```

## Quick Start

Create `main.go` in your project:

```go
package main

import (
	"log"

	"github.com/arivictor/gomark"
)

func main() {
	app := gomark.App{
		Title:      "My Docs",
		ContentDir: "content",
		Mode:       gomark.PreRender,
	}

	if err := app.Run(":8080"); err != nil {
		log.Fatal(err)
	}
}
```

The HTTP server is part of the package. You do not need to set one up yourself
for the default use case.

## Runner Package

Use the runner as a package with one call:

```go
package main

import (
	"log"

	"github.com/arivictor/gomark/runner"
)

func main() {
	if err := runner.Start(); err != nil {
		log.Fatal(err)
	}
}
```

`runner.Start()` reads configuration from environment variables:

- `PORT` (default: `8080`)
- `RUNNER_ADDR` (optional full listen address; overrides `PORT`)
- `RUNNER_AUTH_MODE` (`bearer_static` or `none`)
- `RUNNER_AUTH_TOKEN` (required when auth mode is `bearer_static`)

You can override config in code when needed:

```go
if err := runner.Start(
	runner.WithPort("9090"),
	runner.WithAuth(runner.AuthModeBearerStatic, "my-token"),
); err != nil {
	log.Fatal(err)
}
```

## Project Layout

- Put markdown files in `content/`.
- Use frontmatter (title/description) where needed.
- Optional custom templates can be provided with `TemplatesDir` or
  `LayoutPath` + `TemplateGlob`.
- Optional static assets can be provided with `PublicDir`.

If `PublicDir` is not set, gomark serves built-in default assets (favicons and
OG images) from embedded files.

## Publishing This Module

This repository is already a Go module (`github.com/arivictor/gomark`). To
publish versions users can consume reliably:

```bash
git tag v0.1.0
git push origin v0.1.0
```

Then users can pin versions with:

```bash
go get github.com/arivictor/gomark@v0.1.0
```
