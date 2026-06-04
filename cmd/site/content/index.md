---
title: GoMark
description: GoMark is a self-hosted Go application for building markdown-powered documentation sites and embeddable Go code runners.
nav_title: Home
---

# GoMark

GoMark is a self-hosted Go application for turning a folder of markdown into a real website. Routing, rendering, navigation, search, sitemap, robots, and embedded assets ship in the box — point it at your content and run it.

**This site is built with GoMark.** Check out the [source code](https://github.com/arivictor/gomark).

Build documentation sites, product handbooks, and developer guides with nothing but markdown. You clone the repo, drop in your content, and run the server. When you need interactive examples, GoMark also runs Go snippets over HTTP, so readers can execute code right inside your docs.

## Why GoMark

Everything you need to ship a polished site comes built in, with sensible defaults you can override when you're ready:

- **Built-in HTTP server** — no separate web framework to wire up
- **File-based routing** — your markdown tree *is* your URL structure
- **Generated navigation** — sidebar and top-level nav built from your folders
- **Search out of the box** — a ready-to-query endpoint at `/api/search`
- **SEO on by default** — generated `sitemap.xml` and `robots.txt`
- **Embedded templates and assets** — a presentable site before you touch a single template
- **Runnable Go examples** — optional runner integration for live code

## Start here

- [Install and launch your first site](/getting-started)
- [Run Go snippets with one call](/runner)
- [Configure how the app behaves](/guides/configuration)
- [Customize templates and assets](/guides/customization)
- [Ship to production](/guides/deployment)
- [Browse the public API](/reference)

## Two commands

GoMark ships as two binaries under `cmd/`, each backed by its own package: the **site** server (`cmd/site`, package `internal/site`) and the **runner** server (`cmd/runner`, package `internal/runner`). They share a small wire contract in `internal/protocol`.

### Build a site

`site.NewSite(...)` serves a content directory as a complete website. This is the `cmd/site` entrypoint you edit and run.

```go:title="cmd/site/main.go"
package main

import (
	"log"

	"github.com/arivictor/gomark/internal/site"
)

func main() {
	s := site.NewSite(
		site.WithSiteTitle("GoMark Docs"),
		site.WithSiteContentDir("content"),
		site.WithSiteMode(site.PreRender),
	)

	if err := s.Start(); err != nil {
		log.Fatal(err)
	}
}
```

Run it with `go run ./cmd/site`.

### Run code snippets

> Currently supports `go` code fences. More languages may follow.

`runner.NewRunner(...)` stands up the standalone Go runner server — the `cmd/runner` entrypoint.

```go:title="cmd/runner/main.go"
package main

import (
	"log"

	"github.com/arivictor/gomark/internal/protocol"
	"github.com/arivictor/gomark/internal/runner"
)

func main() {
	r := runner.NewRunner(
		runner.WithPort("8081"),
		runner.WithAuth(protocol.AuthNone, ""),
	)

	if err := r.Start(); err != nil {
		log.Fatal(err)
	}
}
```

Run it with `go run ./cmd/runner`.

## What the site generates for you

Point GoMark at a content tree and it does the rest:

1. Registers routes from your file and folder names
2. Treats `index.md` as the route for its folder
3. Builds sidebar navigation from the same tree
4. Serves default favicons and OG images until you override them
5. Exposes search and SEO endpoints automatically