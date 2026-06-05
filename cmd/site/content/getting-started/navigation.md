---
title: Navigation
description: Use folders, index pages, and nav_title to shape the sidebar and top-level navigation.
order: 5
---

# Navigation

The sidebar and top-level nav are generated straight from your content tree, so adding a  markdown file is all it takes to add a link.

## Sidebar behavior

- Root pages and folders appear in the sidebar
- A folder with `index.md` is routable
- A folder without `index.md` still works as a toggle-only section when it has child pages
- Active branches stay open so readers always see nearby pages

## Use `nav_title` for shorter labels

```md:title="content/index.md"
---
title: GoMark Documentation
nav_title: Home
---
```

You keep the full page title while showing a cleaner label in navigation.

## Sidebar depth

Keep the sidebar tidy by capping how deep it nests. Set it in `gomark.yaml`:

```yaml:title="gomark.yaml"
build:
  sidebar_depth: 3
```

Or, driving GoMark from Go, with `WithSiteSidebarDepth`:

```go:title="main.go"
s := gomark.NewSite(
	gomark.WithSiteContentDir("content"),
	gomark.WithSiteSidebarDepth(3),
)
```

Use a smaller depth when your content tree is broad and you want a simpler sidebar.

## Top navigation and social links

The sidebar comes from your content tree, but the top-of-page navigation and footer
social links are explicit — set them in `gomark.yaml`:

```yaml:title="gomark.yaml"
nav:
  - label: Home
    url: /
  - label: GitHub
    url: https://github.com/me/my-docs

social:
  - label: X
    url: https://x.com/myhandle
    icon: twitter
```