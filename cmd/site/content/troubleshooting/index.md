---
title: Troubleshooting
description: Fix common GoMark setup issues around content discovery, routes, rendering modes, and runner auth.
order: 5
---

# Troubleshooting

Hit a snag? These are the issues people run into most, each with a one-step fix.

## The site fails to start with no markdown pages found

Make sure your content directory exists and contains at least one markdown file, usually `content/index.md`.

## A folder does not have its own route

Add `index.md` inside the folder if you want the folder itself to be routable.

## A page title looks wrong in the sidebar

Set `nav_title` in frontmatter when the navigation label should be shorter than the page title.

## Content changes do not appear immediately

Run the dev server with `gomark serve ./content --live`. Without `--live` (and in anything `gomark build` produces) GoMark takes a snapshot up front, so edits won't show until you rebuild.

## Search returns no results

Make sure your content exists under the content directory you passed to `gomark serve`/`gomark build` and that the command started successfully. The search index is built from the same content tree.

## The runner run button does not appear

Check all of the following:

1. The runner is not disabled (no `--no-runner` flag — or, via the Go API, `gomark.WithSiteRunnerEnabled(false)` / `PLAYGROUND_ENABLED=false`)
2. The code fence language is `go`
3. The fence includes `run=true` (or `editable=true`)

## Clicking Run does nothing or shows "runner failed to load"

The runner downloads a WebAssembly module on first use. Make sure `/runner.wasm` and `/wasm_exec.js` are reachable and that the browser supports WebAssembly. The first run can take a moment while the module downloads; it is cached afterward.

## A snippet errors or behaves differently than `go run`

Snippets run through the yaegi interpreter compiled to WebAssembly, which covers a large subset of Go but not all of it. Reflection-heavy code, `unsafe`, and `cgo` are unsupported, and there is no local filesystem or raw network socket access — snippets are confined to the browser sandbox (Go's WebAssembly HTTP client goes through the browser's `fetch`, subject to CORS). See [How the Runner Works](/runner) for the full list.