# GoMark deploys as a static site. The docs/ directory is rendered at image-build
# time with the gomark CLI and then served by Caddy. There is no long-running Go
# process in production, and the in-browser Go runner needs no backend.

# Stage 1: render the static site by consuming gomark from the docs module.
FROM golang:1.25-alpine AS builder

ENV CGO_ENABLED=0
WORKDIR /src

# Install the gomark CLI in its own layer so it can be cached across source
# changes.
RUN go install github.com/arivictor/gomark/cmd/gomark@v0.1.22

COPY . .

# Build from the docs site using the installed CLI. Output and URL come from
# docs/gomark.yaml.
WORKDIR /src/docs
RUN gomark build

# Stage 2: serve the static output. The Caddyfile binds to Cloud Run's $PORT
# (falling back to 80 locally) and sets the application/wasm content type for
# the in-browser runner module.
FROM caddy:2-alpine AS site

COPY docs/Caddyfile /etc/caddy/Caddyfile
COPY --from=builder /out/site /usr/share/caddy

EXPOSE 8080
