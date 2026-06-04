---
title: Runner
description: Enable runnable Go code blocks in your docs by connecting a site to the GoMark runner.
---

# Runner

Turn static code samples into live ones. GoMark can attach run controls to your Go code fences and send execution requests to a GoMark runner, so readers run and edit examples without ever leaving the page.

## Enable the feature

```go:title="cmd/site/main.go"
s := site.NewSite(
	site.WithSiteContentDir("cmd/site/content"),
	site.WithSiteMode(site.PreRender),
	site.WithSiteRunner("http://localhost:8081", protocol.AuthBearerStatic, "my-runner-token"),
)
```

The auth mode comes from the shared `internal/protocol` package, so add it to your imports:

```go
import (
	"github.com/arivictor/gomark/internal/protocol"
	"github.com/arivictor/gomark/internal/site"
)
```

## Mark runnable code fences

Opt a code block in by adding metadata to its Go fence.

~~~md
```go:title="hello.go" run=true editable=true
package main

import "fmt"

func main() {
    fmt.Println("hello")
}
```
~~~

When runner support is enabled, GoMark can render run controls for Go code blocks marked as runnable or editable.

## Recommended setup

1. Run the docs site on one port
2. Run the GoMark runner on another port
3. Protect the runner with bearer auth unless you're working locally

See [Runner](/runner) for the runner server side.