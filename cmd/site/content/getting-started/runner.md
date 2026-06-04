---
title: Runner Guide
description: Connect a GoMark site to a runner so readers can run and edit Go snippets in your docs.
order: 5
---

# Runner Guide

This guide is about the site side of runner support: connecting your docs site to a runner, enabling the UI, and marking code blocks as runnable.

> If you need to start or secure the runner process itself, use [Runner Server](/runner).

GoMark attaches run controls to your Go code fences and sends execution requests to a GoMark runner, so readers can run and edit examples without ever leaving the page.

## Enable the feature

```go:title="cmd/site/main.go"
s := gomark.NewSite(
	gomark.WithSiteContentDir("cmd/site/content"),
	gomark.WithSiteMode(gomark.PreRender),
	gomark.WithSiteRunner("http://localhost:8081", gomark.AuthNone, ""),
)
```

With that in place, the site proxies execution requests to the runner.

Use `gomark.AuthNone` only for local development. For anything shared or public, use bearer auth with `gomark.AuthBearerStatic` and a secret token.

## Mark runnable code fences

GoMark looks for Go code fences with `run=true` or `editable=true` to attach run controls. The runner validates that the snippet is actually runnable Go code before execution, so you can mark any code block as runnable without worrying about breaking the site.

~~~markdown
```go:title="hello.go":run=true:editable=true
package main

import "fmt"

func main() {
    fmt.Println("hello")
}
```
~~~

This renders as below, give it a try!

```go:title="hello.go":run=true:editable=true
package main

import "fmt"

func main() {
	fmt.Println("hello")
}
```

The runner expects at a minimum a `package main` declaration and a `func main()`, but the snippet can include any valid Go code beyond that. The runner compiles and runs the whole snippet, so readers can experiment with imports, helper functions, and more.:

The runner will not accept:

1. Code blocks without `package main`
2. Code blocks without `func main()`
3. Code blocks with syntax errors
4. Code blocks in languages other than Go
5. Imported modules other than the standard library

See [Runner Server](/runner) for the runner process itself.