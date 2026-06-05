# GoMark deploys as a static site: build the CLI, render the docs to static
# files, then serve them with a static web server. There is no long-running Go
# process in production — and the in-browser Go runner needs no backend.

# Stage 1: build the gomark CLI and render the static site.
FROM golang:1.24-alpine AS builder

WORKDIR /src

# Cache modules first for faster rebuilds.
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/gomark ./cmd/gomark
RUN /out/gomark build ./cmd/site/content /out/site --url https://gomark.dev

# Stage 2: serve the static output. Caddy sends the correct application/wasm
# content type for the runner module automatically.
FROM caddy:2-alpine AS site

COPY --from=builder /out/site /usr/share/caddy

EXPOSE 80
