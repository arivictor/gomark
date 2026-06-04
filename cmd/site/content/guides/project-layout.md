---
title: Project Layout
description: Organize content, templates, and public assets for a GoMark site.
---

# Project Layout

GoMark needs just one thing: a content directory. Custom templates and public assets are optional — add them when you want more control, skip them and the embedded defaults take over.

## Typical project structure

```text
my-site/
  main.go
  content/
    index.md
    getting-started/
      index.md
      install.md
  templates/
    layout.html
    markdown.html
    error.html
  public/
    favicon.ico
    og-image.png
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