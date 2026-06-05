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

Point it at your markdown and ship. Routing, navigation, search, sitemap, robots, and
the in-browser runner are all built in.

Read the docs at [gomark.dev](https://gomark.dev).

## Install

```bash
go install github.com/arivictor/gomark/cmd/gomark@latest
```

## Quick start

Create a `content` directory and add markdown files. The file structure maps to the URL
structure — `content/docs/hello.md` is served at `/docs/hello`, and `index.md` files
serve at their folder path.

```markdown:title="content/docs/hello.md"
# Hello, World!

Welcome to my docs site.
```

Preview it locally with live reload, then build a static site:

```bash
# Dev server: renders live and auto-reloads the browser as you edit
gomark serve ./content --live

# Production: render to a static site you can host anywhere
gomark build ./content ./dist --url https://docs.example.com
```

The output of `gomark build` is plain HTML/CSS/JS that runs on any static host —
GitHub Pages, Netlify, S3, nginx. There is no server to run in production, and the Go
runner executes entirely in the reader's browser. See the
[deployment guide](https://gomark.dev/getting-started/deployment) for GitHub Pages,
container, and other host recipes.

## Use it as a library

GoMark is also a single importable package, `github.com/arivictor/gomark`, if you'd
rather drive it from Go:

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
	)

	// Build a static site...
	if err := s.Export("dist"); err != nil {
		log.Fatal(err)
	}
	// ...or run the dev server: s.Serve(":8080", true)
}
```