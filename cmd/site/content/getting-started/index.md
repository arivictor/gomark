---
title: Getting Started
description: Install the GoMark package, create a content tree, and launch your first markdown-powered website.
---

# Getting Started

Go from an empty directory to a running, navigable site in a few minutes. This guide walks the shortest path that still gives you something real.

## Install the package

GoMark is an importable Go package. Create a module for your site and add it:

```terminal
mkdir my-docs && cd my-docs
go mod init example.com/my-docs
go get github.com/arivictor/gomark/site
```

## Create a content tree

Your markdown tree is your site. GoMark maps files and folders directly to routes — no config required. Point GoMark at any directory; this guide uses `content/`.

```text
content/
  index.md
  guides/
    index.md
    install.md
```

- `content/index.md` becomes `/`
- `content/guides/index.md` becomes `/guides`
- `content/guides/install.md` becomes `/guides/install`

## Add your first page

Add frontmatter to give each page a title, description, and navigation label.

```md:title="content/index.md"
---
title: My Docs
description: The home page for my GoMark site.
nav_title: Home
---

# My Docs

This site is powered by GoMark.
```

## Start your site

A few lines of Go is all it takes — the HTTP server is part of the package. Create a `main.go` that points at your content:

```go:title="main.go"
package main

import (
	"log"

	"github.com/arivictor/gomark/site"
)

func main() {
	s := site.NewSite(
		site.WithSiteTitle("My Docs"),
		site.WithSiteContentDir("content"),
		site.WithSiteMode(site.PreRender),
	)

	if err := s.Start(); err != nil {
		log.Fatal(err)
	}
}
```

Run it:

```terminal
go run .
```

Visit `http://localhost:8080` and GoMark renders your markdown tree as a live website.

## What you get immediately

No extra setup, no plugins — the moment your site boots, you have:

- HTML rendering for every markdown page
- Sidebar navigation built from your folders and pages
- Top-level navigation from your top-level sections
- A search endpoint at `/api/search`
- Generated `sitemap.xml` and `robots.txt`
- Default templates and public assets

## Next steps

- [Understand the content layout](/guides/project-layout)
- [Configure the app](/guides/configuration)
- [Customize templates and assets](/guides/customization)
- [Set up the Go runner](/runner)