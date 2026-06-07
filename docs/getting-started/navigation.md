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

## Add icons to sidebar entries

Set `icon` (or `nav_icon`) in a page's frontmatter to show a small icon next to
its sidebar entry. The value is a [Lucide](https://lucide.dev/icons/) icon name.
For a folder, set it on its `index.md` — the icon applies to the folder's entry.

```md:title="content/guides/index.md"
---
title: Guides
icon: book
---
```

```md:title="content/rocket.md"
---
title: Getting started fast
icon: rocket
---
```

Omit `icon` and the entry renders without one — existing content keeps working
unchanged.

## Hide the sidebar on a page

Set `show_nav: false` in frontmatter to render a page without its sidebar —
handy for landing pages or full-width layouts. The page still contributes to
the nav tree (so other pages can link to and through it); only its own sidebar
is hidden.

```md:title="content/landing.md"
---
title: Welcome
show_nav: false
---
```

## On-page table of contents

Pages with headings get an automatic table of contents alongside the content.
Control it per page with frontmatter:

```md:title="content/page.md"
---
title: Reference
show_toc: false   # hide the TOC on this page (alias: toc: false)
toc_depth: 2      # only show H2s (default: 3, i.e. H2-H4)
---
```

`show_toc` takes precedence if both `show_toc` and `toc` are set, so you can
use whichever name reads better in your content. The reader-mode ("focus
mode") toggle lives at the top of the TOC tools — it hides the header,
sidebar, and TOC so readers can concentrate on the page content, and a
floating button lets them exit it from anywhere.

## Suggested next/previous pages

Every page that has siblings (other pages in the same folder) automatically
gets "Previous" / "Next" links at the bottom of its content, in the same order
they appear in the sidebar — by frontmatter `order`, falling back to
alphabetical title. There's nothing to configure; add `order` to your pages if
you want to control the sequence.

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