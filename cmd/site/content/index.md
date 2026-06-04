---
title: GoMark
description: GoMark is a batteries-included Go package for building markdown-powered documentation sites and embeddable Go code runners.
nav_title: Home
---

# GoMark

GoMark is a batteries-included Go package for turning a folder of markdown into a real website. Routing, rendering, navigation, search, sitemap, robots, and embedded assets ship in the box — point it at your content and run it.

**This site is built with GoMark.** Check out the [source code](https://github.com/arivictor/gomark).

Build documentation sites, product handbooks, and developer guides with nothing but markdown and a few lines of Go. When you need interactive examples, GoMark also runs Go snippets over HTTP, so readers can execute code right inside your docs.

## Why GoMark

Everything you need to ship a polished site comes built in, with sensible defaults you can override when you're ready:

- **Built-in HTTP server** — no separate web framework to wire up
- **File-based routing** — your markdown tree *is* your URL structure
- **Generated navigation** — sidebar and top-level nav built from your folders
- **Search out of the box** — a ready-to-query endpoint at `/api/search`
- **SEO on by default** — generated `sitemap.xml` and `robots.txt`
- **Embedded templates and assets** — a presentable site before you touch a single template
- **Runnable Go examples** — optional playground integration for live code

## Start here

- [Install and launch your first site](/getting-started)
- [Run Go snippets with one call](/runner)
- [Configure how the app behaves](/guides/configuration)
- [Customize templates and assets](/guides/customization)
- [Ship to production](/guides/deployment)
- [Browse the public API](/reference)

## Two entry points

GoMark gives you two ways in, depending on what you're building.

### Build a site

Reach for `gomark.App` when you want GoMark to serve a content directory as a complete website.

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

### Run code snippets

> Currently supports `go` code fences. More languages may follow.

Reach for `gomark.Start()` when you want the standalone Go runner server.

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

## [What the site generates for you](#what-the-site-generates-for-you)

Point GoMark at a content tree and it does the rest:

1. Registers routes from your file and folder names
2. Treats `index.md` as the route for its folder
3. Builds sidebar navigation from the same tree
4. Serves default favicons and OG images until you override them
5. Exposes search and SEO endpoints automatically