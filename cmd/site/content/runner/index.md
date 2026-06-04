---
title: Runner
description: Start the GoMark runner with one call and configure its auth and address behavior.
---

# Runner

The runner is GoMark's code-execution engine: a small HTTP server that compiles and runs Go snippets on demand. It's what powers live runners in your docs — and it's a small constructor-based API to stand up.

## Entry point

`gomark.NewRunner(...).Start()` is the public entry point for the runner. Call it with options to configure in code, or call it with no options and it reads address and auth settings from the environment.

## Local development

Get a runner going locally with auth turned off:

```go:title="main.go"
package main

import (
	"log"

	"github.com/arivictor/gomark"
)

func main() {
	runner := gomark.NewRunner(
		gomark.WithAuth(gomark.AuthModeNone, ""),
	)

	if err := runner.Start(); err != nil {
		log.Fatal(err)
	}
}
```

That's the fastest path to a working runner on your machine.

## Environment-driven startup

Prefer config outside your code? `gomark.NewRunner().Start()` works with no options at all when the environment supplies the auth configuration.

```terminal
export RUNNER_AUTH_MODE=bearer_static
export RUNNER_AUTH_TOKEN=my-runner-token
go run ./cmd/runner
```

## Configure in code

```go:title="main.go"
package main

import (
	"log"

	"github.com/arivictor/gomark"
)

func main() {
	runner := gomark.NewRunner(
		gomark.WithPort("9090"),
		gomark.WithAuth(gomark.AuthModeBearerStatic, "my-runner-token"),
	)

	if err := runner.Start(); err != nil {
		log.Fatal(err)
	}
}
```

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

```go:title="main.go"
runner := gomark.NewRunner(
	gomark.WithAuth(gomark.AuthModeNone, ""),
)

if err := runner.Start(); err != nil {
	log.Fatal(err)
}
```

## Endpoints

- `GET /healthz` — returns `ok`
- `POST /run` — executes a Go snippet request

## Pairing with site runners

The runner really shines when you wire it up to a docs site with runner execution enabled. See [Runner](/guides/runner) for the site-side settings.