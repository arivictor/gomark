---
title: GoMark
description: GoMark is a batteries-included Go package for building markdown-powered documentation sites and embeddable Go code runners.
nav_title: Home
---

# GoMark

GoMark turns a folder of markdown files into a navigable website with search, sitemap generation, embedded assets, and optional runnable Go examples.

It is built for two jobs:

1. Serve a markdown-powered site with `gomark.App`.
2. Run Go snippets over HTTP with `gomark.Start()`.

## Why GoMark

- Built-in HTTP server for the docs site.
- File-based routing from your markdown tree.
- Generated sidebar and top-level navigation.
- Search endpoint at `/api/search`.
- Generated `sitemap.xml` and `robots.txt`.
- Embedded default templates and public assets.
- Optional playground integration for runnable Go examples.

## Start Here

- [Install and launch a site](/getting-started)
- [Use the runner with one call](/runner)
- [Configure app behavior](/guides/configuration)
- [Customize templates and assets](/guides/customization)
- [Prepare for deployment](/guides/deployment)
- [Browse the public API](/reference)

## The Two Entry Points

## Build a site

Use `gomark.App` when you want GoMark to serve a content directory as a full website.

```go:title="main.go"
package main

import (
	"log"

	"github.com/arivictor/gomark"
)

func main() {
	app := gomark.App{
		Title:      "GoMark Docs",
		ContentDir: "content",
		Mode:       gomark.PreRender,
	}

	if err := app.Run(":8080"); err != nil {
		log.Fatal(err)
	}
}
```

## Run Go snippets

Use `gomark.Start()` when you want the Go runner server.

```go:title="main.go"
package main

import (
	"log"

	"github.com/arivictor/gomark"
)

func main() {
	if err := gomark.Start(
		gomark.WithPort("8080"),
		gomark.WithAuth(gomark.AuthModeNone, ""),
	); err != nil {
		log.Fatal(err)
	}
}
```

## What the Site Generates

Point GoMark at a content tree and it will:

1. Register routes from file and folder names.
2. Use `index.md` as the route for its folder.
3. Build sidebar navigation from the same tree.
4. Serve default favicons and OG images unless you override them.
5. Expose search and SEO endpoints automatically.

## Suggested Reading Order

1. [Getting Started](/getting-started)
2. [Runner](/runner)
3. [Configuration](/guides/configuration)
4. [Customization](/guides/customization)
5. [Navigation](/guides/navigation)
6. [Search](/guides/search)
7. [Playground](/guides/playground)
8. [Deployment](/guides/deployment)
9. [API Reference](/reference)
