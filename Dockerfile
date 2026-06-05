# GoMark deploys as a static site. The docs/ directory is its own gomark.dev
# site and a downstream *consumer* of this module: it declares the gomark CLI as
# a Go tool dependency and renders itself with it — exactly as an external user
# would — with a replace directive pointing the dependency at this working tree.
# There is no long-running Go process in production, and the in-browser Go runner
# needs no backend.

# Stage 1: render the static site by consuming gomark from the docs module.
FROM golang:1.24-alpine AS builder

ENV CGO_ENABLED=0
WORKDIR /src

# Pre-fetch module downloads in their own layer so they're cached across source
# changes. docs/ replaces gomark with ../, so the root go.mod must be present for
# the module graph to resolve; go mod download then fetches only what the CLI
# compiles against (yaml.v3 — yaegi is wasm-only and not pulled).
COPY go.mod go.sum ./
COPY docs/go.mod docs/go.sum ./docs/
RUN cd docs && go mod download

COPY . .

# Build from the docs site, which resolves the gomark CLI via its go.mod tool
# directive (replace => ../, i.e. this repo). Output and URL come from
# docs/gomark.yaml; the positional output dir below overrides it for the image.
WORKDIR /src/docs
RUN go tool gomark build . /out/site

# Stage 2: serve the static output. Caddy sends the correct application/wasm
# content type for the runner module automatically.
FROM caddy:2-alpine AS site

COPY --from=builder /out/site /usr/share/caddy

EXPOSE 80
