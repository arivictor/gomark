---
title: Troubleshooting
description: Fix common GoMark setup issues around content discovery, routes, rendering modes, and runner auth.
order: 5
---

# Troubleshooting

Hit a snag? These are the issues people run into most, each with a one-step fix.

## The site fails to start with no markdown pages found

Make sure your content directory exists and contains at least one markdown file, usually `content/index.md`.

## A folder does not have its own route

Add `index.md` inside the folder if you want the folder itself to be routable.

## A page title looks wrong in the sidebar

Set `nav_title` in frontmatter when the navigation label should be shorter than the page title.

## Content changes do not appear immediately

Use `gm.LiveRender` for local development. `gm.PreRender` takes a snapshot at startup.

## Search returns no results

Make sure your content exists under the configured `ContentDir` and that the site started successfully. The search index is built from the same content tree.

## The runner run button does not appear

Check all of the following:

1. `gomark.WithSiteRunner` is set in your site configuration
2. The code fence language is `go`
3. The fence includes `run=true`
4. The runner URL and auth settings are valid

## Runner requests return unauthorized

If you use `AuthBearerStatic`, the client must send `Authorization: Bearer <token>` and the token must match the configured value.

## The runner proxy returns forbidden

The site proxy now requires a CSRF token and a same-origin browser request. Make sure the request comes from a GoMark page, that the browser still has the CSRF cookie, and that the run button is sending the `X-CSRF-Token` header.