---
title: Playground
description: Enable runnable Go code blocks in your docs by connecting a site to the GoMark runner.
---

# Playground

GoMark can attach run controls to Go code fences and send execution requests to a GoMark runner.

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

Use a Go code fence with metadata.

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

1. Run the docs site on one port.
2. Run the GoMark runner on another port.
3. Protect the runner with bearer auth unless you are working locally.

See [Runner](/runner) for the runner server side.