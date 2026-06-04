# gomark

![GoMark splash](site/public/gomark-twitter-1200x628.png)

Build a markdown-powered website in Go with batteries included: routing, rendering,
navigation, search, sitemap, robots, and static site assets.

Read the docs at [gomark.dev](https://gomark.dev).

## Install

```bash
go get github.com/arivictor/gomark/site@latest
```

GoMark is published as importable packages: `github.com/arivictor/gomark/site`
(the site server), `github.com/arivictor/gomark/runner` (the Go code runner), and
`github.com/arivictor/gomark/protocol` (the shared wire contract).

## Quick Start

Create `main.go` in your project:

```go
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