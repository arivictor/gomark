---
title: GoMark
description: GoMark is a Go package for building markdown-powered documentation sites and embeddable Go code runners.
nav_title: Home
order: 0
---

# GoMark

**This site is built with GoMark.** Check out the [source code](https://github.com/arivictor/gomark).

GoMark is a Go package for turning a folder of markdown into a real website. Routing, rendering, navigation, search, sitemap, robots, and embedded assets ship in the box — point it at your content and run it.

Build documentation sites, product handbooks, and developer guides with nothing but markdown and a few lines of Go. `go get` the package, drop in your content, and run the server. When you need interactive examples, GoMark runs Go snippets entirely in the reader's browser (a WebAssembly build of the yaegi interpreter), so readers can execute code right inside your docs with nothing running on your servers.

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

- **Built-in HTTP server** — no separate web framework to wire up
- **File-based routing** — your markdown tree *is* your URL structure
- **Generated navigation** — sidebar and top-level nav built from your folders
- **Search out of the box** — a ready-to-query endpoint at `/api/search`
- **SEO on by default** — generated `sitemap.xml` and `robots.txt`
- **Built-in theme** — a presentable, responsive site with zero frontend setup
- **Runnable Go examples** — optional runner integration for live code