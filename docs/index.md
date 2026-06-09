---
title: GoMark
description: GoMark is a Go package for building markdown-powered documentation sites and embeddable Go code runners.
nav_title: Home
order: 0
icon: house
show_nav: true
show_toc: false
---

![GoMark Logo](/logo/gomark-lockup-on-dark.png)

Use GoMark to turn your folder of markdown into a real website. Build documentation sites, product handbooks, and developer guides with nothing but markdown and a few lines of Go.

```bash
gomark serve ./my-docs
```

Everything you need to ship a polished documentation site comes built in, with defaults you can override when you're ready:

- **Built-in HTTP server** — no separate web framework to wire up
- **File-based routing** — your markdown tree *is* your URL structure
- **Generated navigation** — sidebar and top-level nav built from your folders
- **Search out of the box** — a ready-to-query endpoint at `/api/search`
- **SEO on by default** — generated `sitemap.xml` and `robots.txt`
- **Built-in theme** — a presentable, responsive site with zero frontend setup
- **Runnable Go examples** — optional runner integration for live code

## Run real code examples

```go:title="main.go":run=true:editable=true
package main

// Edit me then click "Run" to see the output!
func main() {
	name := "GoMark"
	println("Hello, " + name + "!")
}
```

[Get started](/getting-started) with building your site.


