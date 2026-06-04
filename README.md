# gomark

![GoMark splash](public/gomark-twitter-1200x628.png)

**Docs for Go libraries where every example actually runs.**

GoMark turns a folder of markdown into a real documentation site — and the Go code
blocks in your docs run live, right in the reader's browser. No playground server, no
backend, no infrastructure: execution happens client-side via a WebAssembly build of
the [yaegi](https://github.com/traefik/yaegi) interpreter, so your examples stay honest
and your servers stay boring.

```go:title="main.go":run=true
package main

func main() {
	println("Hello, Gophers!") // edit me and hit Run
}
```

> _The block above is runnable on [gomark.dev](https://gomark.dev) — a site built with
> GoMark itself. Try editing it and clicking **Run**._

<!-- TODO: replace the static splash above with a short GIF of a reader editing and
running this snippet — the differentiator should be visible in the first 3 seconds. -->

`go get` the package, point it at your markdown, and ship. Routing, navigation, search,
sitemap, robots, and static export are all built in.

Read the docs at [gomark.dev](https://gomark.dev).

## Install

```bash
go get github.com/arivictor/gomark@latest
```

GoMark is a single importable package: `github.com/arivictor/gomark`.

## Quick Start

Create `main.go` in your project:

```go
package main

import (
	"log"

	gm "github.com/arivictor/gomark"
)

func main() {
	s := gm.NewSite(
		gm.WithSiteTitle("My Docs"),
		gm.WithSiteContentDir("content"),
		gm.WithSiteMode(gm.PreRender),
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