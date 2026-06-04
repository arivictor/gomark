---
title: Search
description: Understand the built-in search endpoint and how GoMark indexes markdown content.
---

# Search

GoMark generates a search index from your content and exposes it over HTTP.

## Endpoint

`GET /api/search`

### Query parameters

- `q`: search query
- `limit`: optional result limit, default `8`, capped at `25`

## Example request

```terminal
curl "http://localhost:8080/api/search?q=render&limit=5"
```

## Example response

```json:title="response.json"
{
  "query": "render",
  "results": [
    {
      "title": "Configuration",
      "path": "/guides/configuration",
      "snippet": "...Configure GoMark with App fields, environment-driven behavior, and feature toggles..."
    }
  ]
}
```

## Notes

- Empty queries return an empty result set.
- Search is built from your markdown content.
- In `PreRender` mode the search index is built at startup.