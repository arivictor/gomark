# Changelog

All notable changes to this project are documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
- `gomark version` command, reporting the build version, commit, and date.
- Markdown: underscore emphasis (`_em_`, `__strong__`), strikethrough (`~~del~~`),
  GitHub-style task lists (`- [ ]`, `- [x]`), bare-URL autolinking, and backslash
  escapes for punctuation.
- A themed `404.html` is now written by `gomark build` for static hosts.
- In-browser runner now executes in a Web Worker with a hard timeout, so a
  runaway snippet (e.g. an infinite loop) no longer freezes the reader's tab.
- `gomark.yaml`: unknown or mistyped keys are reported as warnings instead of
  being silently ignored.
- CI now runs `govulncheck` vulnerability scanning and validates the GoReleaser
  configuration on every change.

### Fixed
- Markdown: space-flanked asterisks (e.g. `2 * 3 * 4`) are no longer rendered as
  emphasis, and an ordered list resumed after a paragraph keeps its numbering via
  an `<ol start>` attribute instead of restarting at 1.
- Release: replaced a deprecated GoReleaser archive property
  (`format_overrides.format`) with the supported `formats` list.

## [0.1.0] - 2026-06-06

Initial public release.

### Added
- Markdown-to-static-site generator: a folder of Markdown maps to a routed,
  themed documentation site (`gomark build`), plus a live-reload development
  server (`gomark serve --live`).
- In-browser Go runner: runnable Go code blocks execute entirely client-side via
  a WebAssembly build of the [yaegi](https://github.com/traefik/yaegi) interpreter.
- Built-in full-text search, `sitemap.xml`, `robots.txt`, and SEO metadata
  (Open Graph / Twitter cards, canonical URLs).
- Configuration via `gomark.yaml` (title, logo, SEO, navigation, social links,
  analytics) with CLI-flag and environment-variable overrides.
- Importable Go library API (`github.com/arivictor/gomark`).

[Unreleased]: https://github.com/arivictor/gomark/compare/v0.1.0...HEAD
[0.1.0]: https://github.com/arivictor/gomark/releases/tag/v0.1.0
