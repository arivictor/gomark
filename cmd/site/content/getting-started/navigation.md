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

Keep the sidebar tidy by capping how deep it nests. The default is `2`. This option isn't exposed on the CLI, so reach for the [Go API](/getting-started/configuration#use-it-as-a-library):

```go:title="main.go"
s := gomark.NewSite(
	gomark.WithSiteContentDir("content"),
	gomark.WithSiteSidebarDepth(3),
)
```

Use a smaller depth when your content tree is broad and you want a simpler sidebar. Everything else on this page works the same whether you preview with `gomark serve` or build with `gomark build`.
