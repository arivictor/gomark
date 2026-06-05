---
title: GoMark
description: GoMark is a Go package for building markdown-powered documentation sites and embeddable Go code runners.
nav_title: Home
order: 0
---

# GoMark

**This site is built with GoMark.** Check out the [source code](https://github.com/arivictor/gomark).

GoMark turns a folder of markdown into a real website. Routing, rendering, navigation, search, sitemap, robots, and embedded assets ship in the box — point the `gomark` CLI at your content and you're running.

Build documentation sites, product handbooks, and developer guides with nothing but markdown. Install the CLI, drop in your content, preview it live with `gomark serve ./content --live`, then `gomark build` a static site you can host anywhere. When you need interactive examples, GoMark runs Go snippets entirely in the reader's browser (a WebAssembly build of the yaegi interpreter), so readers can execute code right inside your docs with nothing running on your servers. Prefer Go? It's a single importable package too.

```shell
go install github.com/arivictor/gomark/cmd/gomark@latest
gomark serve ./content --live
gomark build ./content ./dist --url https://docs.example.com
```

## Live Code Examples

```go:title="main.go":run=true
package main

func main() {
	println("Hello, World!")
}
```

Experiment in the [playground](/playground), then [get started](/getting-started) building your site.

## Why GoMark

Everything you need to ship a polished site comes built in, with sensible defaults you can override when you're ready:

- **One CLI** — `serve` to preview live, `build` to ship a static site
- **Static output** — host the build anywhere; no server to run in production
- **File-based routing** — your markdown tree *is* your URL structure
- **Generated navigation** — sidebar and top-level nav built from your folders
- **Search out of the box** — a `/api/search` endpoint when serving, a client-side index when built
- **SEO on by default** — generated `sitemap.xml` and `robots.txt`
- **Embedded templates and assets** — a presentable site before you touch a single template
- **Runnable Go examples** — optional runner integration for live code