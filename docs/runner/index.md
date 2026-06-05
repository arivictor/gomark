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

The runner is **on by default** — there is nothing to provision. To turn the run controls off across the site, pass `--no-runner` to the CLI:

```shell
gomark serve ./content --live --no-runner
gomark build ./content ./dist --no-runner
```

You can also set `build.runner: false` in `gomark.yaml`. From the [Go API](/getting-started/configuration#library-options) the equivalent is `gomark.WithSiteRunnerEnabled(false)`, or set `PLAYGROUND_ENABLED=false` in the environment.

To mark individual code fences as runnable or editable, see the [Runner Guide](/getting-started/runner).

## Security

### The runner runs in the browser

GoMark executes Go snippets **entirely in the reader's browser**, using a WebAssembly build of the yaegi interpreter. There is no execution server, no code-execution endpoint, and no Go code runs on your infrastructure.

This eliminates the largest risk a docs playground usually carries: server-side remote code execution. The blast radius of any snippet is the reader's own browser tab, sandboxed by the browser itself — the same sandbox that runs every other script on the page.

What this means in practice:

- **Nothing to secure on the server.** There is no runner service, no bearer token, and no `/run` endpoint to lock down or rate-limit.
- **No local filesystem and no raw network sockets.** Snippets are confined to the browser's sandbox — the same boundary as any other script on the page, not a GoMark-enforced allow-list. (Go's WebAssembly HTTP client is backed by the browser's `fetch`, so an outbound HTTP request is possible only within the page's CORS/same-origin rules, never as unrestricted network access.)
- **Output is capped** so a runaway print loop cannot exhaust browser memory.

The one remaining caveat is that a deliberate infinite loop freezes the reader's *own* tab (execution is single-threaded on the page's main thread). It affects no one else and no server. Running the interpreter in a Web Worker with a watchdog is a planned improvement.

### CSRF

GoMark includes built-in protection against Cross-Site Request Forgery for any state-changing requests to the site. It generates a per-session CSRF token, validates a matching token on unsafe requests, and additionally requires the request to come from the same origin as the site. Safe methods (GET, HEAD, OPTIONS) are exempt.

### Keep GoMark updated

Regularly update GoMark to pick up security patches and improvements, including updates to the bundled runner runtime.
