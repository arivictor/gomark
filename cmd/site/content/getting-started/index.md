---
title: Getting Started
description: Clone GoMark, create a content tree, and launch your first markdown-powered website.
---

# Getting Started

Go from a fresh clone to a running, navigable site in a few minutes. This guide walks the shortest path that still gives you something real.

## Get the code

GoMark is a self-hosted app — clone the repository and build from source:

```terminal
git clone https://github.com/arivictor/gomark.git
cd gomark
```

## Create a content tree

Your markdown tree is your site. GoMark maps files and folders directly to routes — no config required. The `cmd/site` server reads its content from `cmd/site/content` by default.

```text
cmd/site/content/
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

The `cmd/site` entrypoint wires everything together — the HTTP server is built in. Edit it to point at your content, then run it.

```go:title="cmd/site/main.go"
package main

import (
	"log"

	"github.com/arivictor/gomark/site"
)

func main() {
	s := site.NewSite(
		site.WithSiteTitle("My Docs"),
		site.WithSiteContentDir("cmd/site/content"),
		site.WithSiteMode(site.PreRender),
	)

	if err := s.Start(); err != nil {
		log.Fatal(err)
	}
}
```

Launch it from the repository root:

```terminal
go run ./cmd/site
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