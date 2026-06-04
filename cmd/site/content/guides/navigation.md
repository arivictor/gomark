---
title: Navigation
description: Use folders, index pages, and nav_title to shape the sidebar and top-level navigation.
---

# Navigation

Navigation is generated from your content tree.

## Sidebar behavior

- Root pages and folders appear in the sidebar.
- A folder with `index.md` is routable.
- A folder without `index.md` can still act as a toggle-only section if it has child pages.
- Active branches stay open so readers can see nearby pages.

## Use `nav_title` for shorter labels

```md:title="content/index.md"
---
title: GoMark Documentation
nav_title: Home
---
```

This keeps the full page title while showing a cleaner label in navigation.

## Sidebar depth

Limit the tree depth with `SidebarDepth`.

```go:title="main.go"
app := gomark.App{
	ContentDir:   "content",
	SidebarDepth: 3,
}
```

Use a smaller depth when your content tree is broad and you want a simpler sidebar.