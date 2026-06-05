---
title: Project Layout
description: Organize content, templates, and public assets for a GoMark site.
order: 4
---

# Project Layout

GoMark needs just one thing to render a site: a content directory. You point the `gomark` CLI at it — there's no project scaffold to generate and no entry point to write. Custom templates and public assets are optional — add them when you want more control, skip them and the embedded defaults take over.

## Recommended structure

You can organize your project however you like, but here's a recommended structure to get you started:

```text
project/
  content/
    index.md
  templates/ (optional)
    layout.html
    markdown.html
    error.html
  public/ (optional)
    favicon.ico
    logo.png
```

- `content/` is your markdown tree. GoMark maps files and folders directly to routes — no config required. Preview it with `gomark serve ./content --live` and build it with `gomark build ./content ./dist`.
- `templates/` holds your custom templates. Only add this if you want to override the embedded defaults. Custom templates are wired up through the [Go API](/getting-started/customization).
- `public/` holds your static assets. Only add this if you want to override the embedded defaults — also through the [Go API](/getting-started/customization).

> Using GoMark as a [library](/getting-started/configuration#use-it-as-a-library) instead of the CLI? Add a `main.go` at the project root as your entry point and point `gomark.NewSite(...)` at these same directories.

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