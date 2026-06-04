---
title: Runner Server
description: Start the GoMark runner HTTP server and configure its auth, address, and execution limits.
order: 2
---

# Runner Server

The runner is GoMark's code-execution engine: a small HTTP server that compiles and runs Go snippets on demand. It powers live code execution in your docs, but this page is specifically about the runner service itself: how to start it, configure it, and secure it.

## What this page covers

Use this page when you are working on the runner process itself.

- Start the runner locally
- Configure auth and listen address
- Tune execution timeout
- Understand the runner HTTP endpoints

If you are wiring a docs site to a runner, use [Runner Guide](/guides/playground) instead.

## Entry point

`runner.NewRunner(...).Start()` is the entry point for the runner, wired up in `cmd/runner/main.go`. Call it with options to configure in code, or call it with no options and let the environment provide address and auth settings.

## Local development

Get a runner going locally with auth turned off:

```go:title="cmd/runner/main.go"
package main

import (
	"log"

	gm gm "github.com/arivictor/gomark"
)

func main() {
	r := gm.NewRunner(
		gm.WithAuth(gm.AuthNone, ""),
	)

	if err := r.Start(); err != nil {
		log.Fatal(err)
	}
}
```

That's the fastest path to a working runner on your machine.

## Environment-driven startup

Prefer config outside your code? `runner.NewRunner().Start()` works with no options at all when the environment supplies the auth configuration.

```terminal
export RUNNER_AUTH_MODE=bearer_static
export RUNNER_AUTH_TOKEN=my-runner-token
go run ./cmd/runner
```

## Configure in code

```go:title="cmd/runner/main.go"
package main

import (
	"log"

	gm gm "github.com/arivictor/gomark"
)

func main() {
	r := gm.NewRunner(
		gm.WithPort("9090"),
		gm.WithAuth(gm.AuthBearerStatic, "my-runner-token"),
		gm.WithTimeout(30),
	)

	if err := r.Start(); err != nil {
		log.Fatal(err)
	}
}
```

## Execution timeout

Each `/run` request is capped by an execution timeout. By default the runner allows 2 seconds per snippet; raise or lower it with `WithTimeout`, which takes a whole number of seconds.

```go:title="cmd/runner/main.go"
r := gm.NewRunner(
	gm.WithAuth(gm.AuthNone, ""),
	gm.WithTimeout(10), // give snippets up to 10 seconds
)
```

Values of `0` or less are ignored and the default timeout stays in effect.

## Environment variables

Configure the runner without touching code:

- `PORT` — listen port, default `8080`
- `RUNNER_ADDR` — full listen address; overrides `PORT`
- `RUNNER_AUTH_MODE` — `bearer_static` or `none`
- `RUNNER_AUTH_TOKEN` — required when auth resolves to `bearer_static`

## Auth modes

The runner executes code, so it ships secure by default. Choose the mode that fits where it's running.

### `bearer_static`

The safe default for any runner exposed outside local development. Clients must send an `Authorization: Bearer ...` header.

When `RUNNER_AUTH_MODE` is unset, the runner resolves to `bearer_static` — so you'll either provide a token or explicitly opt into `none`. No accidental open endpoints.

### `none`

Reserve this for local development or fully trusted networks.

```go:title="cmd/runner/main.go"
r := runner.NewRunner(
	runner.WithAuth(AuthNone, ""),
)

if err := r.Start(); err != nil {
	log.Fatal(err)
}
```

## Endpoints

- `GET /healthz` — returns `ok`
- `POST /run` — executes a Go snippet request

## Pairing with a docs site

The runner really shines when you wire it up to a docs site with runner execution enabled. See [Runner Guide](/guides/playground) for the site-side configuration and runnable code fences.