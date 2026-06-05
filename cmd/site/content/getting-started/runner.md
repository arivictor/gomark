---
title: Runner Guide
description: Make Go snippets in your docs runnable and editable, so readers can experiment without leaving the page.
order: 5
---

# Runner Guide

This guide covers marking code blocks as runnable so readers can run and edit Go examples inline. Execution happens entirely in the reader's browser — see [How the Runner Works](/runner) for the architecture.

## Enable the feature

The runner is **enabled by default**. There is no service to run and nothing to configure — `gomark serve` and `gomark build` ship it automatically, and `gomark build` even bundles the runtime into your static output.

To turn the run controls off across the whole site, pass `--no-runner` to either command:

```shell
gomark serve ./content --live --no-runner
gomark build ./content ./dist --no-runner
```

From the [Go API](/getting-started/configuration#use-it-as-a-library) the equivalent is `gomark.WithSiteRunnerEnabled(false)` (or set `PLAYGROUND_ENABLED=false`).

## Mark runnable code fences

GoMark attaches run controls to Go code fences marked with `run=true` or `editable=true`:

~~~markdown
```go:title="hello.go":run=true:editable=true
package main

import "fmt"

func main() {
    fmt.Println("hello")
}
```
~~~

This renders as below — give it a try:

```go:title="hello.go":run=true:editable=true
package main

import "fmt"

func main() {
	fmt.Println("hello")
}
```

The first run downloads the in-browser runtime (cached afterward), then executes the snippet locally.

## What runs

A snippet needs a `package main` declaration and a `func main()`. Beyond that you can use most of the standard library, generics, goroutines, and helper functions. Snippets run through the yaegi interpreter, so a few things behave differently from `go run`:

1. Some reflection-heavy code, `unsafe`, and `cgo` are unsupported
2. There is no filesystem or network access in the browser sandbox
3. A deliberate infinite loop freezes the reader's own tab
4. Output is capped to protect browser memory

See [How the Runner Works](/runner) for the full list of limitations.
