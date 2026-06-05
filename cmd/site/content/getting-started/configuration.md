---
title: Configuration
description: Configure GoMark with the CLI commands and flags, control render modes, and drop down to the Go API when you need more.
order: 2
---

# Configuration

The `gomark` CLI is the configuration surface for most sites. There are two commands — `build` and `serve` — and a small set of flags. Set the flags you need and focus on writing content instead of wiring up features.

## Commands

```text
gomark build <content-dir> <output-dir> [flags]   Render a static site to disk
gomark serve <content-dir> [flags]                Preview locally
```

`build` produces the static output you deploy. `serve` is a development tool only — see [Deployment](/getting-started/deployment) for how production works.

## Flags

| Flag | Commands | Description |
| --- | --- | --- |
| `--url` | `build`, `serve` | Public site URL. Drives canonical links, `sitemap.xml`, `robots.txt`, and SEO metadata. |
| `--title` | `build`, `serve` | Site title shown in the layout and meta tags. |
| `--no-runner` | `build`, `serve` | Disable the in-browser Go runner. |
| `--live` | `serve` | Render on every request and auto-reload the browser as files change. |
| `--port` | `serve` | Port to listen on. Default `8080`. |

Flags can appear before or after the positional arguments, so both of these work:

```shell
gomark serve --live ./content
gomark serve ./content --live
```

## Render modes

GoMark renders one of two ways, and the command you run picks the mode for you.

### Live rendering (`gomark serve --live`)

- Reads markdown from disk on each request
- Reflects file edits — and structural changes — without a restart
- Best for local development

### Pre-rendering (`gomark build`, and `gomark serve` without `--live`)

- Builds all output up front
- Fails fast on content issues
- Best for production; it's exactly what `build` writes to disk

## Recommended flow

Two commands cover most projects.

While you write:

```shell
gomark serve ./content --live
```

When you ship:

```shell
gomark build ./content ./dist --url https://docs.example.com
```

## Use it as a library

GoMark is also a single importable package, `github.com/arivictor/gomark`, if you'd rather drive it from Go — and it's the way to reach options the CLI doesn't expose, such as [custom templates and public assets](/getting-started/customization) and [sidebar depth](/getting-started/navigation#sidebar-depth).

```shell
go get github.com/arivictor/gomark
```

`gomark.NewSite(...)` with `gomark.With...` options is the configuration surface:

```go:title="main.go"
package main

import (
	"log"

	gm "github.com/arivictor/gomark"
)

func main() {
	s := gm.NewSite(
		gm.WithSiteTitle("My Docs"),
		gm.WithSiteLogo("/logo.svg"),
		gm.WithSiteContentDir("content"),
		gm.WithSiteURL("https://docs.example.com"),
		gm.WithSiteMode(gm.PreRender),
	)

	// Build a static site...
	if err := s.Export("dist"); err != nil {
		log.Fatal(err)
	}
	// ...or run the dev server: s.Serve(":8080", true)
}
```

Common options:

- `WithSiteTitle` — site name used in layout and metadata
- `WithSiteLogo` — optional logo URL shown in the header
- `WithSiteContentDir` — markdown content root, default `content`
- `WithSiteURL` — base URL used for sitemap and canonical URLs
- `WithSiteMode` — `gomark.LiveRender` or `gomark.PreRender`
- `WithSiteRunnerEnabled` — toggles the in-browser Go runner (on by default)
- `WithSiteTemplatesDir`, `WithSiteLayoutPath`, `WithSiteTemplateGlob` — custom templates
- `WithSitePublicDir` — static asset directory that overrides embedded defaults
- `WithSiteSidebarDepth` — max sidebar depth, default `2`

`s.Export("dist")` mirrors `gomark build`; `s.Serve(":8080", true)` mirrors `gomark serve --live`. GoMark also recognizes common environment aliases such as `prod`, `production`, `live`, and `development` when resolving the render mode.
