---
title: Project Layout
description: Organize content, templates, and public assets for a GoMark site.
order: 4
---

# Project Layout

GoMark needs just one thing to render a site: a content directory. Custom templates and public assets are optional — add them when you want more control, skip them and the embedded defaults take over.

## Recommended structure

You can organize your project however you like, but here's a recommended structure to get you started:

```text
project/
  main.go
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

- `main.go` is your app entry point. Point it at your content and optionally templates and public assets.
- `content/` is your markdown tree. GoMark maps files and folders directly to routes — no config required.
- `templates/` holds your custom templates. Only add this if you want to override the embedded defaults.
- `public/` holds your static assets. Only add this if you want to override the embedded defaults.

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