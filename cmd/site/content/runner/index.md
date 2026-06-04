---
title: How the Runner Works
description: GoMark runs Go snippets entirely in the reader's browser using a WebAssembly build of the yaegi interpreter — no server, no auth, no code execution on your infrastructure.
order: 3
---

# How the Runner Works

The runner is GoMark's code-execution engine. It powers live, editable Go examples in your docs — and it runs **entirely in the reader's browser**. There is no execution server, no endpoint to secure, and no Go code ever runs on your infrastructure.

## The model

When a reader clicks **Run** on a code fence, GoMark lazy-loads a WebAssembly module and executes the snippet locally in the page:

1. On the first run, the browser fetches `runner.wasm` (~8 MB gzipped, served with long-lived immutable caching) and `wasm_exec.js`.
2. The module — a [yaegi](https://github.com/traefik/yaegi) Go interpreter compiled to `GOOS=js GOARCH=wasm` — registers a `runGo(source)` function.
3. Each run creates a fresh interpreter, executes the snippet, and returns its combined output. Nothing leaves the browser.

Because execution is client-side, the blast radius of any snippet is the reader's own tab. There is no server-side remote-code-execution surface to defend.

## What it can run

yaegi interprets a large subset of Go: the standard library, generics, goroutines, structs, slices, and maps all work, which covers the vast majority of documentation examples.

A snippet must declare `package main` and a `func main()`. The interpreter runs `main` automatically.

```go:title="hello.go":run=true:editable=true
package main

import "fmt"

func main() {
	fmt.Println("Hello from your browser")
}
```

## Known limitations

Because this is an interpreter compiled to WebAssembly rather than the `gc` toolchain, a few things differ from `go run`:

- **Not 100% of Go.** Some reflection-heavy code, `unsafe`, `cgo`, and a handful of stdlib corners are unsupported. Most teaching snippets are unaffected.
- **No local filesystem; no raw sockets.** The browser sandbox provides neither. Go's WebAssembly HTTP client routes through the browser's `fetch`, so an HTTP call is bound by the page's CORS rules rather than being open network access — don't treat this as a security boundary GoMark enforces.
- **Single-threaded, main-thread execution.** A deliberate infinite loop (`for {}`) will freeze the reader's own tab until they close it. Output is capped to protect browser memory.
- **First-run download.** The module is fetched once on first use, then cached.

## Enabling and disabling

The runner is **on by default** — there is nothing to provision. To turn the run controls off across the site:

```go:title="cmd/site/main.go"
s := gomark.NewSite(
	gomark.WithSiteContentDir("content"),
	gomark.WithSiteRunnerEnabled(false),
)
```

You can also disable it from the environment with `PLAYGROUND_ENABLED=false`.

To mark individual code fences as runnable or editable, see the [Runner Guide](/getting-started/runner).
