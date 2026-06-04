---
title: Runner Guide
description: Connect a GoMark site to a runner so readers can run and edit Go snippets in your docs.
order: 8
---

# Runner Guide

This guide is about the site side of runner support: connecting your docs site to a runner, enabling the UI, and marking code blocks as runnable.

If you need to start or secure the runner process itself, use [Runner Server](/runner).

Turn static code samples into live ones. GoMark can attach run controls to your Go code fences and send execution requests to a GoMark runner, so readers run and edit examples without ever leaving the page.

## Enable the feature

```go:title="cmd/site/main.go"
s := gomark.NewSite(
	gomark.WithSiteContentDir("cmd/site/content"),
	gomark.WithSiteMode(gomark.PreRender),
	gomark.WithSiteRunner("http://localhost:8081", gomark.AuthBearerStatic, "my-runner-token"),
)
```

With that in place, the site exposes the run UI and proxies execution requests to the runner.

Add the package import in `cmd/site/main.go`:

```go
import (
	"github.com/arivictor/gomark"
)
```

Use `gomark.AuthNone` only for local development. For anything shared or public, use bearer auth.

## Mark runnable code fences

Opt a code block in by adding metadata to its Go fence.

~~~markdown
```go:title="hello.go":run=true:editable=true
package main

import "fmt"

func main() {
    fmt.Println("hello")
}
```
~~~

When runner support is enabled, GoMark renders run controls for Go code blocks marked as runnable or editable.

For runnable examples, make the snippet an actual program:

- use `package main`
- define `func main()`

The runner validates both before execution.

## Recommended setup

1. Run the docs site on one port.
2. Run the GoMark runner on another port.
3. Point `WithSiteRunner(...)` at the runner base URL.
4. Protect the runner with bearer auth unless you're working locally.

See [Runner Server](/runner) for the runner process itself.