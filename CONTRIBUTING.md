# Contributing to GoMark

Thanks for your interest in GoMark. Contributions of all sizes are welcome —
bug reports, documentation fixes, and code. This guide covers how to set up,
how the project is laid out, and the conventions we follow.

## Prerequisites

- Go 1.24 or newer.

## Setup

Clone the repository and fetch dependencies:

```bash
git clone https://github.com/arivictor/gomark
cd gomark
go mod download
```

Before pushing, run the same checks CI runs:

```bash
go test ./...   # run the test suite
gofmt -l .      # list files that need formatting (should print nothing)
go vet ./...    # report suspicious constructs
```

If `gofmt -l .` lists any files, format them with `gofmt -w .`.

## Project layout

- **Root package** (`github.com/arivictor/gomark`) — the library. This is the
  importable API used to build and serve sites.
- **`cmd/gomark`** — the `gomark` CLI (`build` and `serve`).
- **`cmd/wasm`** — the in-browser Go runner. It is compiled to WebAssembly and
  runs reader-supplied Go code client-side via the yaegi interpreter.
- **`templates/`** and **`public/`** — embedded assets (the built-in theme,
  scripts, icons, and the compiled runner) that ship inside the binary.
- **`scripts/build-wasm.sh`** — builds the WASM runner.
- **`docs/`** — the GoMark documentation site, itself built with GoMark.

### The committed WASM artifact

`public/runner.wasm.gz` is a **generated artifact**, not hand-edited source. It
is the compiled, gzipped output of `cmd/wasm`. If you change the runner, rebuild
it and commit the result:

```bash
./scripts/build-wasm.sh
```

## Branches and pull requests

- Keep PRs small and focused — one logical change per PR is easiest to review.
- Run `go test ./...`, `gofmt -l .`, and `go vet ./...` before pushing.
- Write clear commit messages. Existing history is conventional-ish: a short,
  imperative subject, optionally prefixed with a type such as `docs:`, `style:`,
  or `fix:`. Match that style.
- Describe what changed and why in the PR, and link any related issue.

## Reporting bugs and requesting features

Use the issue templates under `.github/ISSUE_TEMPLATE`. For security issues,
please follow [SECURITY.md](SECURITY.md) instead of opening a public issue.
