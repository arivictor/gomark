---
title: "Playground"
description: "Run Go code snippets in your site with the playground."
order: 1
---

# Playground

GoMark's playground feature turns static code samples into live ones. When enabled, GoMark can render run controls for Go code blocks marked as runnable or editable, so readers run and edit examples without ever leaving the page.

```go:title="example.go":run=true:editable=true
package main

// Edit me! Then run me!
func main() {
    println("hello, world!")
}
```