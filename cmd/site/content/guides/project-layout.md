---
title: Project Layout
description: Organize content, templates, and public assets for a GoMark site.
order: 2
---

# Project Layout

GoMark needs just one thing to render a site: a content directory. Custom templates and public assets are optional — add them when you want more control, skip them and the embedded defaults take over.

## Repository structure

GoMark is organized as three importable packages, plus two reference binaries under `cmd/`:

```text
gomark/
  site/                # markdown-to-HTML site service (+ embedded templates/ and public/)
  runner/              # Go code executor service
  protocol/            # shared site<->runner wire contract (RunRequest, RunResponse, AuthMode)
  cmd/
    site/
      main.go          # site server entrypoint (reference consumer)
      content/         # your markdown lives here by default
        index.md
        guides/
          index.md
    runner/
      main.go          # runner server entrypoint (reference consumer)
```

The `site` and `runner` packages are the public API; `protocol` holds the request, response, and auth types both sides share. Templates and public assets are embedded inside the `site` package, so the binaries are self-contained.

## Your content directory

A content tree is all you need to get started. Optionally bring your own templates and public assets to override the embedded defaults:

```text
cmd/site/content/
  index.md
  getting-started/
    index.md
    install.md
```

## Routing rules

Routing is convention, not configuration:

- Every markdown file becomes a page route
- `index.md` becomes the route for its folder
- The content root is mounted at `/`
- Folder names become route segments

## Frontmatter fields

Control each page with a small set of metadata fields:

- `title` — page title and default nav label
- `description` — page description for rendering and SEO metadata
- `nav_title` — optional shorter label for navigation

## Choosing folders vs files

Reach for a folder with `index.md` when you want a section landing page with child pages beneath it.

```text
content/
  guides/
    index.md
    configuration.md
    deployment.md
```

This produces `/guides`, `/guides/configuration`, and `/guides/deployment`.