---
title: Customization
description: Brand your GoMark site with a logo, SEO metadata, navigation, social links, and analytics.
order: 3
---

# Customization

GoMark uses a single built-in theme, so there are no layouts or CSS to wire up —
you get a presentable, responsive site with zero frontend setup. What you *do*
customize is the site's identity: its name, logo, SEO metadata, navigation,
social links, and analytics. All of it lives in [`gomark.yaml`](/getting-started/configuration).

> Custom layouts and CSS are intentionally out of scope for now. If your project
> needs them, open an issue describing the use case.

## Branding

```yaml:title="gomark.yaml"
title: My Docs
footer: © 2026 Example, Inc.
theme_color: "#0070f3"

logo:
  light: /logo-light.png   # shown in light theme
  dark: /logo-dark.png     # shown in dark theme
```

The logo swaps automatically with the active theme. Set just one of `light` /
`dark` (or use a single `WithSiteLogo` from Go) to use the same mark for both. If
you omit the logo entirely, the bundled GoMark mark is used.

## SEO

Page-level `title` and `description` come from each markdown file's frontmatter.
Everything else — and the defaults for pages that omit a description — comes from
the `seo` block.

```yaml:title="gomark.yaml"
url: https://docs.example.com   # canonical links + sitemap

seo:
  description: Default description for pages without their own.
  og_image: /og-1200x630.png          # a file you provide at your site root
  twitter_image: /twitter-1200x628.png # a file you provide at your site root
  twitter_site: "@myhandle"
  twitter_creator: "@myhandle"
  image_alt: My Docs            # defaults to the site title
```

`og_image` and `twitter_image` point to image files **you** add at your site root —
the paths above are examples. Omit them and GoMark falls back to its bundled
`/gomark-og-1200x630.png` and `/gomark-twitter-1200x628.png` defaults (which carry
GoMark branding, so you'll usually want your own).

GoMark emits canonical links, Open Graph, and Twitter card tags on every page,
plus `sitemap.xml` and `robots.txt` (toggle them under `build`).

## Navigation and social links

The sidebar is derived from your content tree. Top-of-page navigation and footer
social links are explicit:

```yaml:title="gomark.yaml"
nav:
  - label: Home
    url: /
  - label: GitHub
    url: https://github.com/me/my-docs

social:
  - label: X
    url: https://x.com/myhandle
    icon: twitter          # optional lucide icon name
```

## Analytics

Drop in an analytics snippet without touching any HTML. Set a `provider` and its
`id`, and GoMark injects the matching script into the `<head>` of every page —
content pages and error pages alike. The snippet is only emitted when **both**
`provider` and `id` are set; leave them out and no tracking code ships.

Supported providers are `plausible`, `ga4` (Google Analytics 4), and `gtm`
(Google Tag Manager). The `id` field means something different for each one.

### Plausible

Privacy-friendly, cookie-less analytics. The `id` is your **site domain** as
registered in Plausible (its `data-domain`).

```yaml:title="gomark.yaml"
analytics:
  provider: plausible
  id: docs.example.com
```

Injects:

```html
<script defer data-domain="docs.example.com" src="https://plausible.io/js/script.js"></script>
```

### Google Analytics 4 (`ga4`)

The `id` is your GA4 **Measurement ID**, which starts with `G-`.

```yaml:title="gomark.yaml"
analytics:
  provider: ga4
  id: G-XXXXXXXXXX
```

Injects the standard `gtag.js` loader and config:

```html
<script async src="https://www.googletagmanager.com/gtag/js?id=G-XXXXXXXXXX"></script>
<script>window.dataLayer=window.dataLayer||[];function gtag(){dataLayer.push(arguments);}gtag('js',new Date());gtag('config','G-XXXXXXXXXX');</script>
```

### Google Tag Manager (`gtm`)

The `id` is your GTM **Container ID**, which starts with `GTM-`.

```yaml:title="gomark.yaml"
analytics:
  provider: gtm
  id: GTM-XXXXXXX
```

Injects the GTM loader in the `<head>` plus the `<noscript>` iframe fallback at
the top of `<body>`:

```html
<!-- in <head> -->
<script>(function(w,d,s,l,i){w[l]=w[l]||[];w[l].push({'gtm.start':new Date().getTime(),event:'gtm.js'});var f=d.getElementsByTagName(s)[0],j=d.createElement(s),dl=l!='dataLayer'?'&l='+l:'';j.async=true;j.src='https://www.googletagmanager.com/gtm.js?id='+i+dl;f.parentNode.insertBefore(j,f);})(window,document,'script','dataLayer','GTM-XXXXXXX');</script>

<!-- first thing in <body> -->
<noscript><iframe src="https://www.googletagmanager.com/ns.html?id=GTM-XXXXXXX" height="0" width="0" style="display:none;visibility:hidden"></iframe></noscript>
```

### From Go

When driving GoMark as a library, use `WithSiteAnalytics(provider, id)`:

```go
site := gomark.NewSite(
    gomark.WithSiteAnalytics("plausible", "docs.example.com"),
)
```

## Default assets

Out of the box, GoMark serves bundled `favicon.ico` and PNG variants, an Apple
touch icon, a web manifest, and default Open Graph / Twitter images. Override the
images by setting `seo.og_image` / `seo.twitter_image` to your own files served
from your site root.
