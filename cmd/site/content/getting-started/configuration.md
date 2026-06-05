---
title: Configuration
description: Configure GoMark with a gomark.yaml file, CLI flags, or Site options.
order: 2
---

# Configuration

GoMark is configured through a single declarative `gomark.yaml` file, read by both
`gomark build` and `gomark serve`. Custom layouts and CSS are intentionally not
configurable — every site uses the built-in theme, so you configure identity, SEO,
navigation, and build behavior, then focus on writing content.

## `gomark.yaml`

Drop a `gomark.yaml` in your project root (or content directory). It is
auto-discovered; pass `--config <path>` to point elsewhere.

```yaml:title="gomark.yaml"
title: My Docs
url: https://docs.example.com
lang: en
theme_color: "#0070f3"
footer: © 2026 Example, Inc.

logo:
  light: /logo-light.png
  dark: /logo-dark.png

seo:
  description: Short default description for pages without their own.
  og_image: /og-1200x630.png
  twitter_image: /twitter-1200x628.png
  twitter_site: "@myhandle"
  twitter_creator: "@myhandle"
  image_alt: My Docs

nav:
  - label: Home
    url: /
  - label: GitHub
    url: https://github.com/me/my-docs

social:
  - label: X
    url: https://x.com/myhandle
    icon: twitter

analytics:
  provider: plausible   # ga4 | gtm | plausible
  id: docs.example.com

build:
  content_dir: content
  output_dir: dist
  sidebar_depth: 2
  runner: true          # in-browser Go runner
  sitemap: true
  robots: true
```

Every field is optional; omit what you don't need.

## Precedence

When the same setting comes from more than one place, the highest wins:

```
CLI flag  >  environment variable  >  gomark.yaml  >  built-in default
```

So `gomark build --url https://staging.example.com` overrides the `url:` in your
config for a one-off build, while the file stays the durable home for everything
else.

## CLI flags

```bash
gomark build [<content-dir> [<output-dir>]] [flags]
gomark serve [<content-dir>] [flags]
```

- `--config` — path to `gomark.yaml` (auto-discovered by default)
- `--title` — site title
- `--url` — public site URL (canonical links, sitemap, SEO)
- `--no-runner` — disable the in-browser Go runner
- `--live` (serve) — render live and auto-reload on file changes
- `--port` (serve) — port to listen on, default `8080`

Paths can come from the config too: with `build.content_dir` and
`build.output_dir` set, `gomark build` needs no positional arguments.

## Library options

Driving GoMark from Go uses the matching `WithSite...` options:

- `WithSiteTitle`, `WithSiteLang`, `WithSiteThemeColor`, `WithSiteFooter`
- `WithSiteLogoLight`, `WithSiteLogoDark` (or `WithSiteLogo` for both)
- `WithSiteDescription`, `WithSiteOGImage`, `WithSiteTwitterImage`,
  `WithSiteTwitterSite`, `WithSiteTwitterCreator`, `WithSiteImageAlt`
- `WithSiteNavLinks`, `WithSiteSocialLinks`, `WithSiteAnalytics`
- `WithSiteContentDir`, `WithSiteURL`, `WithSiteSidebarDepth`, `WithSiteMode`
- `WithSiteRunnerEnabled`, `WithSiteSitemapEnabled`, `WithSiteRobotsEnabled`

`FileConfig` (loaded via `gomark.LoadConfigFile`) exposes `.Options()` so you can
load `gomark.yaml` from your own program too.

## Render modes

GoMark renders one of two ways.

### `gomark.LiveRender`

- Reads markdown from disk on each request
- Best for local development (`gomark serve --live`)
- Reflects file edits without restart

### `gomark.PreRender`

- Builds markdown output up front at startup
- Best for production or stable content (`gomark build` always pre-renders)
- Fails fast on content issues during boot

GoMark also recognizes common environment aliases such as `prod`, `production`,
`live`, and `development` when resolving render mode.
