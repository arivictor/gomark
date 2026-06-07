---
title: Getting Started
description: Install the GoMark package, create a content tree, and launch your first markdown-powered website.
order: 2
icon: book
---

# Getting Started

GoMark turns a folder of markdown into a real website. This guide walks through the basics: installing the package, creating a content tree, and launching your first markdown-powered site.

## Install

GoMark is both a CLI and an importable Go package. For the quickest start, install the CLI:

```bash:title="terminal"
go install github.com/arivictor/gomark/cmd/gomark@latest
```

Prefer to drive it from Go? Add the package to your module instead:

```bash:title="terminal"
go get github.com/arivictor/gomark
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

```markdown:title="content/index.md"
---
title: My Docs # Shows in the page header and meta tags
description: The home page for my GoMark site. # Optional: shows in search results and meta tags.
nav_title: Home # Optional: controls sidebar label
order: 0 # Optional: controls sidebar order
---

# My Docs

This site is powered by GoMark.
```

## Start your site

The CLI previews your content with live reload — no code required:

```bash:title="terminal"
gomark serve content --live
```

Visit `http://localhost:8080` and GoMark renders your markdown tree as a live website, reloading as you edit.

### Configure it (optional)

Site title, logo, SEO, navigation, and analytics live in an optional `gomark.yaml` that `serve` and `build` auto-discover. See the [configuration guide](/getting-started/configuration) for the full schema.

```yaml:title="gomark.yaml"
title: My Docs
url: https://docs.example.com
seo:
  description: Docs for my project.
```

### Drive it from Go (optional)

Prefer code? The HTTP server is part of the package, so a few lines of Go does the same thing:

```go:title="main.go"
package main

import (
	"log"

	gm "github.com/arivictor/gomark"
)

func main() {
	s := gm.NewSite(
		gm.WithSiteTitle("My Docs"),
		gm.WithSiteContentDir("content"),
		gm.WithSiteMode(gm.PreRender), // Use LiveRender for local development to see changes immediately
	)

	if err := s.Start(); err != nil {
		log.Fatal(err)
	}
}
```

## Opinionated features out of the box

No extra setup, no plugins — the moment your site boots, you have:

- HTML rendering for every markdown page
- Sidebar navigation built from your folders and pages
- Top-level navigation from your top-level sections
- A search endpoint at `/api/search`
- Generated `sitemap.xml` and `robots.txt`
- A presentable, responsive built-in theme — no frontend setup

## Next steps

- [Understand the content layout](/getting-started/project-layout)
- [Configure the app](/getting-started/configuration)
- [Customize branding & SEO](/getting-started/customization)
- [Set up the Go runner](/runner)