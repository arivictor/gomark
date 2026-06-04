---
title: Runner
description: Start the GoMark runner with one call and configure its auth and address behavior.
---

# Runner

GoMark includes an HTTP runner for executing Go snippets. The public entry point is `gomark.Start()`.

## Entry point

`gomark.Start()` is the single public start function for the runner.

If you call it without options, it reads address and auth settings from the environment.

## Local development

```go:title="main.go"
package main

import (
	"log"

	"github.com/arivictor/gomark"
)

func main() {
	if err := gomark.Start(
		gomark.WithAuth(gomark.AuthModeNone, ""),
	); err != nil {
		log.Fatal(err)
	}
}
```

This is the easiest way to run the runner locally.

## Environment-driven startup

`gomark.Start()` also works without options when the environment provides the auth configuration.

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
	if err := gomark.Start(
		gomark.WithPort("9090"),
		gomark.WithAuth(gomark.AuthModeBearerStatic, "my-runner-token"),
	); err != nil {
		log.Fatal(err)
	}
}
```

## Environment variables

- `PORT`: listen port, default `8080`
- `RUNNER_ADDR`: full listen address, which overrides `PORT`
- `RUNNER_AUTH_MODE`: `bearer_static` or `none`
- `RUNNER_AUTH_TOKEN`: required when auth resolves to `bearer_static`

## Auth modes

### `bearer_static`

This is the safe default for any runner exposed outside local development. Clients must send an `Authorization: Bearer ...` header.

When `RUNNER_AUTH_MODE` is unset, handler creation resolves to `bearer_static`, so you must either provide a token or explicitly switch to `none`.

### `none`

Use this only for local development or fully trusted networks.

```go:title="main.go"
if err := gomark.Start(
	gomark.WithAuth(gomark.AuthModeNone, ""),
); err != nil {
	log.Fatal(err)
}
```

## Endpoints

- `GET /healthz` returns `ok`
- `POST /run` executes a Go snippet request

## Pairing with site playgrounds

The runner becomes useful when your site enables Go playground execution. See [Playground](/guides/playground) for the site-side settings.