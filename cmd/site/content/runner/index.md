---
title: Runner
description: Start the GoMark runner with one call and configure its auth and address behavior.
---

# Runner

The runner is GoMark's code-execution engine: a small HTTP server that compiles and runs Go snippets on demand. It's what powers live runners in your docs — and it's the `cmd/runner` binary, backed by the `internal/runner` package.

## Entry point

`runner.NewRunner(...).Start()` is the entry point for the runner, wired up in `cmd/runner/main.go`. Call it with options to configure in code, or call it with no options and it reads address and auth settings from the environment.

## Local development

Get a runner going locally with auth turned off:

```go:title="cmd/runner/main.go"
package main

import (
	"log"

	"github.com/arivictor/gomark/internal/protocol"
	"github.com/arivictor/gomark/internal/runner"
)

func main() {
	r := runner.NewRunner(
		runner.WithAuth(protocol.AuthNone, ""),
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

	"github.com/arivictor/gomark/internal/protocol"
	"github.com/arivictor/gomark/internal/runner"
)

func main() {
	r := runner.NewRunner(
		runner.WithPort("9090"),
		runner.WithAuth(protocol.AuthBearerStatic, "my-runner-token"),
		runner.WithTimeout(30),
	)

	if err := r.Start(); err != nil {
		log.Fatal(err)
	}
}
```

## Execution timeout

Each `/run` request is capped by an execution timeout. By default the runner allows 2 seconds per snippet; raise or lower it with `WithTimeout`, which takes a whole number of seconds.

```go:title="cmd/runner/main.go"
r := runner.NewRunner(
	runner.WithAuth(protocol.AuthNone, ""),
	runner.WithTimeout(10), // give snippets up to 10 seconds
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
	runner.WithAuth(protocol.AuthNone, ""),
)

if err := r.Start(); err != nil {
	log.Fatal(err)
}
```

## Endpoints

- `GET /healthz` — returns `ok`
- `POST /run` — executes a Go snippet request

## Pairing with site runners

The runner really shines when you wire it up to a docs site with runner execution enabled. See [Runner](/guides/playground) for the site-side settings.