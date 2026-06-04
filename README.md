# gomark

![GoMark splash](public/gomark-twitter-1200x628.png)

Build a markdown-powered website in Go with batteries included: routing, rendering,
navigation, search, sitemap, robots, and static site assets.

Read the docs at [gomark.dev](https://gomark.dev).

## Install

```bash
go get github.com/arivictor/gomark@latest
```

## Quick Start

Create `main.go` in your project:

```go
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

The HTTP server is part of the package. You do not need to set one up yourself
for the default use case.

## Write your docs

Create a `content` directory and add markdown files. The file structure maps to
the URL structure. For example, `content/docs/hello.md` is served at `/docs/hello`.

```markdown:title="content/docs/hello.md"
# Hello, World!

Welcome to my docs site.
```

You can create index pages with `index.md` files. For example, `content/docs/index.md` is served at `/docs/`.