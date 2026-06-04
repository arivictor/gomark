---
title: Customization
description: Override GoMark templates and public assets while keeping the built-in defaults as a fallback.
order: 3
---

# Customization

GoMark ships with embedded templates and public assets, so you get a presentable site with zero frontend setup. When you're ready to make it yours, point the app at your own templates and assets — and anything you don't override keeps falling back to the built-in defaults.

## Custom templates

The simplest override is a directory holding `layout.html` and the page templates GoMark expects.

```go:title="main.go"
s := gomark.NewSite(
	gomark.WithSiteContentDir("content"),
	gomark.WithSiteTemplatesDir("templates"),
	gomark.WithSiteMode(gomark.PreRender),
)
```

If you need explicit paths instead of a directory convention, use `LayoutPath` and `TemplateGlob`.

```go:title="main.go"
s := gomark.NewSite(
	gomark.WithSiteContentDir("content"),
	gomark.WithSiteLayoutPath("templates/layout.html"),
	gomark.WithSiteTemplateGlob("templates/*.html"),
)
```

### Required template files

When using custom templates, make sure your template set includes:

- `layout.html` defining `layout`
- `markdown.html` defining `content` (used for docs pages)
- `error.html` defining `content` (used for error pages)

Template names are derived from file names, so `markdown.html` is rendered as `markdown`, and `error.html` as `error`.

The layout template must render the content block, for example:

```html
{{define "layout"}}
  <main>{{block "content" .}}{{end}}</main>
{{end}}
```

### Data available in templates

All templates receive the same `PageData` object.

| Field | Type | Notes |
| --- | --- | --- |
| `StatusCode` | `int` | HTTP status code (primarily for error pages). |
| `Title` | `string` | Page title. |
| `Description` | `string` | Meta description / page subtitle text. |
| `SiteName` | `string` | Site branding text. |
| `LogoURL` | `string` | Logo image URL. |
| `CanonicalURL` | `string` | Absolute canonical URL for the page. |
| `OGImageURL` | `string` | Open Graph image URL. |
| `TwitterImageURL` | `string` | Twitter image URL. |
| `RunnerEnabled` | `bool` | Whether runner UI is enabled. |
| `Robots` | `string` | Robots meta value (for example `index,follow`). |
| `Time` | `string` | Render timestamp (RFC3339 UTC). |
| `MarkdownFile` | `string` | Source markdown path (content pages). |
| `BodyHTML` | `template.HTML` | Rendered markdown HTML (content pages). |
| `Headings` | `[]Heading` | In-page headings for TOC. |
| `NavTitle` | `string` | Sidebar section title. |
| `Nav` | `[]NavNode` | Sidebar navigation tree. |
| `TopNav` | `[]NavLink` | Top-level navigation links. |
| `CurrentPath` | `string` | Current route path. |

`Heading` contains:

- `Level` (`int`)
- `Text` (`string`)
- `ID` (`string`)

`NavNode` contains:

- `Title` (`string`)
- `Path` (`string`)
- `NodeID` (`string`)
- `Folder` (`bool`)
- `Active` (`bool`)
- `Open` (`bool`)
- `Children` (`[]NavNode`)

`NavLink` contains:

- `Title` (`string`)
- `Path` (`string`)
- `Active` (`bool`)

## Custom public assets

Set `PublicDir` to serve your own favicons, OG images, or any additional static files.

```go:title="main.go"
s := gomark.NewSite(
	gomark.WithSiteContentDir("content"),
	gomark.WithSitePublicDir("public"),
)
```

If `PublicDir` is empty, GoMark serves embedded defaults baked into the `gomark` package.

## What embedded defaults cover

Out of the box, GoMark serves:

- `favicon.ico`
- PNG favicon variants
- Apple touch icon
- Default OG images