---
title: Playground
description: Enable runnable Go code blocks in your docs by connecting a site to the GoMark runner.
---

# Playground

Turn static code samples into live ones. GoMark can attach run controls to your Go code fences and send execution requests to a GoMark runner, so readers run and edit examples without ever leaving the page.

## Enable the feature

```go:title="main.go"
app := gomark.App{
	ContentDir:               "content",
	Mode:                     gomark.PreRender,
	PlaygroundEnabled:        true,
	PlaygroundRunnerURL:      "http://localhost:8081",
	PlaygroundRunnerAuthMode: "bearer_static",
	PlaygroundRunnerAuthToken: "my-runner-token",
}
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

When playground support is enabled, GoMark can render run controls for Go code blocks marked as runnable or editable.

## Recommended setup

1. Run the docs site on one port
2. Run the GoMark runner on another port
3. Protect the runner with bearer auth unless you're working locally

See [Runner](/runner) for the runner server side.