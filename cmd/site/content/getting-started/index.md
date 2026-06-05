---
title: Getting Started
description: Install the GoMark CLI, create a content tree, preview it live, and build a static site you can host anywhere.
order: 2
---

# Getting Started

GoMark turns a folder of markdown into a real website. This guide walks through the whole flow: install the CLI, create a content tree, preview it live as you edit, and build a static site you can host anywhere.

## Install the CLI

The `gomark` command is the primary way to use GoMark. Install it with `go install`:

```shell
go install github.com/arivictor/gomark/cmd/gomark@latest
```

This puts a `gomark` binary on your `PATH` (in `$(go env GOPATH)/bin`). Check it:

```shell
gomark --help
```

> Prefer to drive GoMark from Go instead of the CLI? It's also a single importable package — see [Use it as a library](/getting-started/configuration#use-it-as-a-library).

## Create a content tree

Your markdown tree is your site. GoMark maps files and folders directly to routes — no config required. Point GoMark at any directory; this guide uses `content/`.

```text
content/
  index.md
  guides/
    index.md
    install.md
```

- `content/index.md` becomes `/`
- `content/guides/index.md` becomes `/guides`
- `content/guides/install.md` becomes `/guides/install`

## Add your first page

Add frontmatter to give each page a title, description, and navigation label.

```markdown:title="content/index.md"
---
title: My Docs # Shows in the page header and meta tags
description: The home page for my GoMark site. # Optional: shows in search results and meta tags.
nav_title: Home # Optional: controls sidebar label
order: 0 # Optional: controls sidebar order
---

# My Docs

This site is powered by GoMark.
```

## Preview it live

Run the dev server with `--live` and GoMark renders pages on every request and auto-reloads your browser as you edit:

```shell
gomark serve ./content --live
```

Visit `http://localhost:8080`. Edit a markdown file and the page reloads on its own — including structural changes: adding, renaming, or deleting files updates routes, the sidebar, search, and the sitemap without a restart.

Drop `--live` for a quick static-style preview, and use `--port` to listen elsewhere:

```shell
gomark serve ./content --port 3000
```

> `gomark serve` is a **development tool**, not a production server. You deploy the static output of `gomark build` (next section).

## Build a static site

When you're ready to publish, render your content into an output directory:

```shell
gomark build ./content ./dist --url https://docs.example.com
```

- `./content` — your content directory.
- `./dist` — the output directory to create.
- `--url` — your public origin. Set it: it drives canonical URLs, `sitemap.xml`, `robots.txt`, and Open Graph / Twitter metadata.

The output is plain HTML, CSS, JS, and assets — no server to run in production. Host it on GitHub Pages, Netlify, S3, nginx, or a container. See the [deployment guide](/getting-started/deployment) for ready-to-use recipes.

## Opinionated features out of the box

No extra setup, no plugins — the moment your site renders, you have:

- HTML rendering for every markdown page
- Sidebar navigation built from your folders and pages
- Top-level navigation from your top-level sections
- Full-text search (a `/api/search` endpoint when serving, a client-side index when built)
- Generated `sitemap.xml` and `robots.txt`
- The in-browser Go runner for runnable code examples
- Default templates and public assets

## Next steps

- [Configure the CLI and render modes](/getting-started/configuration)
- [Understand the content layout](/getting-started/project-layout)
- [Customize templates and assets](/getting-started/customization)
- [Set up the Go runner](/getting-started/runner)
- [Deploy your site](/getting-started/deployment)
