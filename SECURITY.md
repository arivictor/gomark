# GoMark Security

**Scope:** The GoMark documentation **Site** — the markdown renderer, HTTP handlers,
search and asset serving, CSRF/auth, container packaging, and the in-browser code
runner. This document describes the security model as it actually ships today.

## System model

GoMark is a single Go module that builds and serves a documentation website:

- **Site** (`cmd/site`) — renders author-controlled markdown to HTML, serves static
  assets, and exposes a read-only `/api/search` endpoint. Distroless, non-root,
  public-facing.
- **Runner** (`cmd/wasm`) — a WebAssembly build of the [yaegi](https://github.com/traefik/yaegi)
  Go interpreter. It is shipped as a static asset (`public/runner.wasm.gz`) and executes
  **entirely in the reader's browser**. There is no execution server, no `/run` HTTP
  endpoint, no bearer token, and no Go code that runs on your infrastructure.

Because the runner is client-side, the blast radius of any snippet is the reader's own
browser tab, sandboxed by the browser itself — the same sandbox that runs every other
script on the page. The server never compiles or executes untrusted code, so the
classic docs-playground risk (server-side remote code execution) does not exist here.

> [!NOTE]
> An earlier version of GoMark ran a separate `cmd/runner` HTTP service that compiled
> and executed Go with `go run` behind bearer auth. That service has been **removed**;
> execution moved fully into the browser via WebAssembly. If you are reading older notes
> that reference `cmd/runner`, `RUNNER_AUTH_TOKEN`, or a `/api/runner/run` proxy, they no
> longer apply.

## What protects the Site

- **CSRF.** `site_csrf.go` combines a same-origin Origin/Referer check with a
  double-submit token compared in constant time (`subtle.ConstantTimeCompare`), and
  fails closed when both Origin and Referer are absent. Browsers cannot forge `Origin`,
  so Host-header games do not break it.
- **Escaping renderer.** The markdown renderer (`site_markdown.go`) routes all text
  output through `html.EscapeString`, has no raw-HTML passthrough, and neutralizes
  `javascript:`/`data:` URLs via `normalizeLinkTarget`. Rendered HTML is injected with
  `template.HTML`, so the renderer is a trusted, security-sensitive boundary: any future
  edit that forgets to escape a branch is a potential stored-XSS, so treat changes to it
  with care (the tests in `site_markdown_test.go` exercise the escaping paths).
- **Container.** The Site builds into `gcr.io/distroless/static-debian12:nonroot` and
  runs as a non-root user with a minimal image.
- **HTTP timeouts.** The Site server sets `ReadHeaderTimeout` (`site_server.go`) to limit
  Slowloris exposure.

## Known gaps / hardening notes

- **Content Security Policy / security headers.** The Site does not yet emit a CSP,
  `X-Content-Type-Options: nosniff`, `X-Frame-Options`/`frame-ancestors`,
  `Referrer-Policy`, or HSTS. Adding `nosniff` and a restrictive CSP is the highest-value
  cheap win and would backstop any future renderer-escaping slip. Consider adding these
  via middleware.
- **Third-party JavaScript.** Syntax highlighting (highlight.js) and icons (Lucide), plus
  the highlight.js theme CSS, are **self-hosted** from `public/vendor/` rather than loaded
  from a CDN. This removes the CDN supply-chain risk entirely and lets docs run fully
  offline / air-gapped (including static exports). If you reintroduce any remote
  `<script>`/`<link>`, add Subresource Integrity (`integrity=`) hashes.

## Reporting

If you believe you have found a security issue in GoMark, please open an issue at
<https://github.com/arivictor/gomark> or contact the maintainer directly. Please do not
include exploit details in a public issue until a fix is available.
