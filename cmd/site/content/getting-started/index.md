---
title: Getting Started
description: Install GoMark, create a content tree, and launch your first markdown-powered website.
---

# Getting Started

This guide gets a GoMark site running with the smallest useful setup.

## Install

```terminal
go get github.com/arivictor/gomark@latest
```

## Create a content tree

GoMark maps markdown files directly to routes.

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

Use frontmatter for page metadata.

```md:title="content/index.md"
---
title: My Docs
description: The home page for my GoMark site.
nav_title: Home
---

# My Docs

This site is powered by GoMark.
```

## Start a site

```go:title="main.go"
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

Visit `http://localhost:8080` and GoMark will render your markdown tree as a website.

## What you get immediately

- HTML rendering for your markdown pages.
- Sidebar navigation from folders and pages.
- Top-level navigation from top-level sections.
- Search endpoint at `/api/search`.
- `sitemap.xml` and `robots.txt`.
- Default templates and public assets.

## Next steps

- [Understand the content layout](/guides/project-layout)
- [Configure the app](/guides/configuration)
- [Customize templates and assets](/guides/customization)
- [Set up the Go runner](/runner)