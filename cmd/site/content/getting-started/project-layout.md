---
title: Project Layout
description: Organize your content tree and optional gomark.yaml for a GoMark site.
order: 4
---

# Project Layout

GoMark needs just one thing to render a site: a content directory. Everything else — the theme, assets, and runner — is built in. A `gomark.yaml` is optional; add it when you want to set a title, logo, SEO, navigation, or analytics.

## Recommended structure

You can organize your project however you like, but here's a recommended structure to get you started:

```text
project/
  gomark.yaml   (optional — site config)
  content/
    index.md
    guides/
      index.md
      install.md
```

- `content/` is your markdown tree. GoMark maps files and folders directly to routes — no config required.
- `gomark.yaml` configures the site (title, logo, SEO, nav, analytics, build options). It's auto-discovered by `gomark build` and `gomark serve`. See the [configuration guide](/guides/configuration).

Driving GoMark from Go instead? Add a `main.go` that calls `gomark.NewSite(...)` — see [Getting Started](/getting-started).

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