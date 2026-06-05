---
title: Deployment
description: Build a GoMark site to static files and host it anywhere — GitHub Pages, Netlify, S3, or your own container. No server required.
order: 7
---

# Deployment

GoMark deploys as a **static site**. You build your markdown into plain HTML, CSS,
JavaScript, and assets, then host the output on any static host. There is no
long-running Go server to operate, and — because the in-browser Go runner executes
client-side via WebAssembly — there is no execution backend to secure or scale either.

## Build the site

Use the `gomark` CLI to render your content directory into an output directory:

```terminal
gomark build ./my_docs ./dist --url https://docs.example.com
```

- `./my_docs` — your content directory (markdown files).
- `./dist` — the output directory to create.
- `--url` — your public origin. Set it: it drives canonical URLs, `sitemap.xml`,
  `robots.txt`, and Open Graph / Twitter image URLs. SEO metadata is wrong without it.
  (You can also set `url:` in `gomark.yaml`; the flag overrides it.)

Title, logo, SEO, navigation, and analytics come from an optional `gomark.yaml` that
`build` auto-discovers. See the [configuration guide](/guides/configuration).

The output is self-contained. It includes your rendered pages (`<route>/index.html`),
copied assets, `sitemap.xml`, `robots.txt`, `search-index.json` for client-side search,
and — when the runner is enabled — `runner.wasm` and `wasm_exec.js`.

> Prefer Go over the CLI? `gomark.NewSite(...).Export("./dist")` does the same thing,
> and so does setting the `EXPORT_DIR` environment variable.

## Preview locally

```terminal
gomark serve ./my_docs --live
```

`gomark serve` is a **development tool**, not a production server. With `--live` it
renders pages on every request and auto-reloads your browser as files under the content
directory change — including structural changes: adding, renaming, or deleting markdown
updates routes, the sidebar, search, and the sitemap without a restart. Drop `--live`
for a quick static-style preview. (You can't open the built files directly over
`file://` — the runner and search use `fetch`, which needs an HTTP origin — so use
`serve` to preview.)

## What you need from a host

Any static host works. Two things make the experience seamless:

1. **`.wasm` served as `application/wasm`.** Required for the fastest runner load path
   (`WebAssembly.instantiateStreaming`). GoMark falls back to a slower load if the host
   sends the wrong type, so the runner still works either way — but the correct MIME
   type is preferred. Most modern hosts get this right automatically.
2. **Clean URLs.** GoMark writes each page as `<route>/index.html`, so hosts that serve
   `index.html` for a directory (almost all of them) give you extensionless URLs with no
   configuration.

## GitHub Pages

Build in CI and publish the output. Save this as `.github/workflows/pages.yml`:

```yaml
name: Deploy docs to GitHub Pages
on:
  push:
    branches: [main]
permissions:
  contents: read
  pages: write
  id-token: write
jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version: '1.24'
      - name: Build site
        run: |
          go install github.com/arivictor/gomark/cmd/gomark@latest
          gomark build ./my_docs ./dist --url https://<user>.github.io/<repo>
      - uses: actions/upload-pages-artifact@v3
        with:
          path: ./dist
  deploy:
    needs: build
    runs-on: ubuntu-latest
    environment:
      name: github-pages
      url: ${{ steps.deployment.outputs.page_url }}
    steps:
      - id: deployment
        uses: actions/deploy-pages@v4
```

Set `--url` to your Pages origin so canonical links and SEO metadata are correct. For a
project site that origin includes the repo path (`https://<user>.github.io/<repo>`).

## Netlify, Vercel, Cloudflare Pages

Point the platform at your repo and configure:

- **Build command:** `go install github.com/arivictor/gomark/cmd/gomark@latest && gomark build ./my_docs ./dist --url https://your.site`
- **Publish / output directory:** `dist`

These platforms handle `application/wasm` and clean URLs out of the box.

## Amazon S3 + CloudFront

```terminal
gomark build ./my_docs ./dist --url https://docs.example.com
aws s3 sync ./dist s3://my-docs-bucket --delete
```

Set the bucket's static-website index document to `index.html`. Ensure objects ending in
`.wasm` are served with `Content-Type: application/wasm` (set it on upload with
`aws s3 cp --content-type`, or via a CloudFront response-headers policy), and invalidate
the CloudFront distribution after each deploy.

## Self-hosted (containers, nginx, Caddy)

Serve the `dist` directory with any static web server. A two-stage container builds the
site and serves it with [Caddy](https://caddyserver.com/) (which sends the correct
`application/wasm` type automatically):

```dockerfile:title="Dockerfile"
# Stage 1: build the CLI and render the static site
FROM golang:1.24-alpine AS builder
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/gomark ./cmd/gomark
RUN /out/gomark build ./my_docs /out/site --url https://docs.example.com

# Stage 2: serve the static output
FROM caddy:2-alpine
COPY --from=builder /out/site /usr/share/caddy
EXPOSE 80
```

```terminal
docker build -t my-docs .
docker run -p 8080:80 my-docs
```

With nginx, copy `dist` into the web root and confirm `application/wasm` is in
`mime.types` (it is on current nginx builds); add
`try_files $uri $uri/ $uri/index.html =404;` for clean URLs.

## Updating the runner

The committed `public/runner.wasm.gz` is the prebuilt in-browser runner and is embedded
in the CLI, so `gomark build` ships it automatically. If you change the runner source
under `cmd/wasm`, regenerate it with `scripts/build-wasm.sh` before building.

## Deployment checklist

1. Build with `gomark build <content> <output>`.
2. Pass `--url` (your public origin) for correct canonical links and SEO.
3. Upload the output directory to your static host.
4. Confirm `.wasm` is served as `application/wasm` (most hosts do this automatically).
5. Drop the runner with `--no-runner` if your docs have no runnable examples — the build omits the decompressed `runner.wasm` and the run controls (the vendored `wasm_exec.js` and `runner.wasm.gz` are still copied as public assets).
