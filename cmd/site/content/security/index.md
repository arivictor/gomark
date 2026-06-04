---
title: Security
description: How GoMark's security model works, including the in-browser runner and CSRF protection.
order: 4
---

# Security

## The runner runs in the browser

GoMark executes Go snippets **entirely in the reader's browser**, using a WebAssembly build of the yaegi interpreter. There is no execution server, no code-execution endpoint, and no Go code runs on your infrastructure.

This eliminates the largest risk a docs playground usually carries: server-side remote code execution. The blast radius of any snippet is the reader's own browser tab, sandboxed by the browser itself — the same sandbox that runs every other script on the page.

What this means in practice:

- **Nothing to secure on the server.** There is no runner service, no bearer token, and no `/run` endpoint to lock down or rate-limit.
- **No local filesystem and no raw network sockets.** Snippets are confined to the browser's sandbox — the same boundary as any other script on the page, not a GoMark-enforced allow-list. (Go's WebAssembly HTTP client is backed by the browser's `fetch`, so an outbound HTTP request is possible only within the page's CORS/same-origin rules, never as unrestricted network access.)
- **Output is capped** so a runaway print loop cannot exhaust browser memory.

The one remaining caveat is that a deliberate infinite loop freezes the reader's *own* tab (execution is single-threaded on the page's main thread). It affects no one else and no server. Running the interpreter in a Web Worker with a watchdog is a planned improvement.

## CSRF

GoMark includes built-in protection against Cross-Site Request Forgery for any state-changing requests to the site. It generates a per-session CSRF token, validates a matching token on unsafe requests, and additionally requires the request to come from the same origin as the site. Safe methods (GET, HEAD, OPTIONS) are exempt.

## Keep GoMark updated

Regularly update GoMark to pick up security patches and improvements, including updates to the bundled runner runtime.
