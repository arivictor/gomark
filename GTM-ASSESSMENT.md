# GoMark — Go-To-Market (GTM) Assessment

_Assessment date: 2026-06-04 · Repo: `arivictor/gomark` · Latest tag: `v0.1.11`_

This is a go-to-market readiness review of GoMark as a product, not a code review.
It looks at what GoMark is, who it's for, how it's differentiated, where it sits in
the market, and what stands between today's state and meaningful adoption. Findings
are grounded in the current `main`/branch contents, repo metadata, and shipped docs.

---

## 1. Executive summary

GoMark is a single-import Go package that turns a folder of markdown into a complete
documentation website — routing, rendering, generated nav, search, sitemap/robots,
embedded templates and assets, static export, and (its signature feature) **runnable
Go examples that execute entirely in the reader's browser** via a WebAssembly build of
the yaegi interpreter. The engineering is solid for a project this young: 71% test
coverage on the core package, a distroless non-root container, CSRF protection, a
client-side runner that eliminates the usual server-side-RCE risk of doc playgrounds,
and self-hosted vendor JS for offline/air-gapped use.

The product is **technically launch-ready but commercially pre-launch.** The repo is
~1 day old, has 0 stars / 0 forks / 0 external issues, twelve `v0.1.x` tags but no
*published* (non-draft) GitHub release, and effectively no distribution or positioning
work yet. The core GTM gap is not the product — it's the absence of a sharp wedge,
proof of traction, and a repeatable adoption path. The single most defensible thing
GoMark has — in-browser runnable Go — is underused as a positioning anchor.

**Overall GTM readiness: ~4/10.** Strong artifact, near-zero distribution.

---

## 2. Product snapshot

| Dimension | State |
|---|---|
| What it is | Go package (`github.com/arivictor/gomark`) — batteries-included markdown site engine |
| Core loop | `go get` → drop markdown in `content/` → `s.Start()` → site |
| Signature feature | In-browser runnable Go snippets (yaegi → WASM, ~8 MB gzipped, no server execution) |
| Modes | `LiveRender` (dev) and `PreRender` (prod), plus static export (`EXPORT_DIR`) |
| Batteries | File-based routing, generated sidebar/top nav, `/api/search`, sitemap.xml, robots.txt, TOC, callouts, tables, images, frontmatter route overrides, mobile nav |
| Security posture | CSRF (origin + double-submit, constant-time), escaping renderer (no raw-HTML passthrough), distroless non-root, self-hosted vendor JS, `ReadHeaderTimeout` |
| Dependencies | Exactly one: `github.com/traefik/yaegi` |
| Quality | ~3,666 non-test LoC, 71.4% package coverage, tests green |
| License | MIT |
| Docs | gomark.dev (dogfooded — the docs site is built with GoMark) |

This is a coherent, opinionated product with a real "wow" demo. The dogfooding
(gomark.dev built on GoMark) is exactly right for a developer tool.

---

## 3. Target market & ICP

**Primary ICP — Go library/tool authors who need docs.** They already write Go, want
docs in the same toolchain (no Node, no Ruby, no separate static-site generator), and
benefit disproportionately from *runnable Go examples*. This is the tightest fit and
where the runner is a genuine differentiator, not a nice-to-have.

**Secondary ICP — Go teams building internal docs/handbooks** who value a single Go
binary, a distroless container, and offline/air-gapped operation (self-hosted assets,
static export) over the richer theming of JS-ecosystem tools.

**Tertiary / aspirational — general docs authors.** Weakest fit. Against Material for
MkDocs, Docusaurus, and Starlight, GoMark's theming/plugin ecosystem and "requires
writing a little Go" are net negatives for non-Go users. Do not lead here.

The runner is the reason to pick GoMark over anything else, and it only matters to
people teaching/demoing **Go**. The ICP and the differentiator must be the same
audience — currently the messaging spreads wider than the moat.

---

## 4. Competitive landscape

| Tool | Stack | Runnable code | GoMark's edge | Their edge |
|---|---|---|---|---|
| Hugo | Go | No | Single import vs. a CLI+theme system; runnable Go | Huge ecosystem, themes, maturity, massive adoption |
| mdBook | Rust | Rust playground (external) | Native Go runner, Go-native authoring | Rust ecosystem, very mature |
| MkDocs + Material | Python | No (plugins only) | No Python; runner; single binary | Best-in-class theme, plugins, search, huge mindshare |
| Docusaurus / Starlight / VitePress | JS | Live React/JS sandboxes | No Node toolchain; Go runner; one binary | Theming, MDX, components, ecosystem, polish |
| GitBook / Mintlify | SaaS | Limited | Self-host, MIT, no vendor lock-in, offline | Hosting, polish, collaboration, zero-ops |
| Go Playground / pkg.go.dev | — | Yes (server) | Embedded *in your docs*, client-side, no infra | Canonical, trusted, ubiquitous |

**Takeaway:** GoMark cannot win "best general docs generator" — that race is over and
crowded. It *can* own a sliver: **"the docs engine for Go libraries, where every
example runs in the reader's browser."** The client-side-execution angle (no server,
no infra, no RCE risk, works offline) is a real and rare combination worth naming
explicitly.

---

## 5. Positioning & messaging

Current positioning (README/landing) is **feature-led**, not **wedge-led**: "Build a
markdown-powered website in Go with batteries included." Accurate, but it competes
on the "yet another static site generator" axis where GoMark is outgunned.

**Recommended primary message:**
> **GoMark — docs for Go libraries where every example actually runs.** Point it at a
> folder of markdown, `go get`, and ship a site with live, in-browser Go snippets. No
> Node, no playground servers, no infra.

**Supporting pillars:**
1. **Runnable by default** — readers execute Go in the page; nothing runs on your servers.
2. **One import, batteries included** — routing, nav, search, SEO, static export.
3. **Boring to operate** — single binary, distroless, offline-capable, MIT.

Lead with the demo (the live `Hello, World!` block already on the landing page is the
hook — make it the *first* thing, above the fold, with an obvious "edit & run").

---

## 6. Distribution & growth (the real gap)

Right now there is essentially no GTM motion. Concrete, ordered next steps:

**Tier 1 — Credibility & on-ramp (do first)**
- **Cut a real, non-draft `v0.x` release** with curated notes. Twelve tags but a draft
  0.1.0 reads as "not actually shipped." A published release is table stakes.
- **README hero = the runner.** Put an animated GIF/short clip of editing+running a Go
  snippet at the very top. The differentiator should be visible in 3 seconds.
- **One-command quickstart** (`go run` a starter, or a `gomark init` scaffold) so the
  time-to-first-site is under 60 seconds.
- **Add repo topics + a crisp GitHub description.** It's discoverable only by name today.

**Tier 2 — Proof & reach**
- **Convert 2–3 real Go libraries' docs to GoMark** (yours or friendly maintainers') and
  link them as showcases. Live-example docs for an actual library is the killer demo.
- **Launch posts** on r/golang, Gopher Slack, Hacker News ("Show HN"), and Lobsters,
  each leading with the runnable-docs angle, not the feature list.
- **A comparison page** ("GoMark vs Hugo/MkDocs/Docusaurus for Go projects") — honest,
  scoped to the Go-author ICP.

**Tier 3 — Compounding**
- **`pkg.go.dev` polish + examples** so the package page itself sells it.
- **A "Made with GoMark" gallery** and a badge to seed social proof.
- **GitHub Action / template repo** for "docs site from markdown on every push" (the
  static-export path already supports this).

---

## 7. Adoption friction (what stops a first user)

1. **8 MB WASM payload.** The runner module is ~8 MB gzipped. On the landing page this
   is the differentiator, but it's a real cost. Confirm it's lazy-loaded only when a
   reader hits "Run" (not on every page), document the size, and consider an opt-in
   note. For runner-free sites it should add zero bytes.
2. **yaegi's interpreter limitations.** yaegi doesn't cover 100% of Go / stdlib / cgo /
   generics edge cases. Set expectations explicitly ("runnable examples support a broad
   subset of Go"), or users will perceive bugs where there are interpreter limits.
3. **"I have to write Go to get a docs site."** True even for trivial sites. A
   zero-Go scaffold (`gomark init` / a template repo) removes this for evaluators.
4. **Theming ceiling.** Customization is templates + public dir, not a theme/plugin
   ecosystem. Fine for the Go-author ICP; a dealbreaker for design-led teams. Don't
   chase them.
5. **No CSP / security headers yet** (self-noted in SECURITY.md). Cheap, high-trust win
   for the "boring to operate / enterprise-friendly" pillar — ship `nosniff` + a
   restrictive CSP via middleware.

---

## 8. Monetization (optional, longer horizon)

GoMark is MIT and infra-light by design, which is great for adoption and weak for
direct monetization. Realistic paths, in order of fit:

- **Open-core / sponsorship** — keep the engine MIT, fund via GitHub Sponsors once
  there's a user base. Most honest fit for a dev tool at this stage.
- **Hosted GoMark** (managed gomark.dev-style hosting + previews) — the natural SaaS,
  but competes with Pages/Netlify/GitBook and needs real demand first.
- **Premium themes/components or a Pro runner** (larger stdlib coverage, multi-file,
  shareable snippets) — defer until adoption exists.

Recommendation: **do not monetize yet.** Optimize purely for adoption and the runnable-
docs narrative for the next several months; monetization is premature without traction.

---

## 9. Risks

- **Differentiator depends on a third party.** The entire moat rides on yaegi's
  capabilities and maintenance. Track upstream health; the value prop degrades if yaegi
  stalls.
- **Narrow TAM.** "Docs for Go libraries with runnable examples" is defensible but
  small. That's fine for an open-source tool; it caps any commercial ambition.
- **Solo/new project signal.** 0 stars, ~1-day-old repo, no published release — every
  external signal currently says "experiment." Tier-1 actions above directly counter this.
- **Scope creep toward "general SSG."** The roadmap and landing both drift toward
  competing with Hugo/Docusaurus on features. That's an unwinnable axis; resist it.

---

## 10. Scorecard

| Area | Score (0–5) | Notes |
|---|---|---|
| Product / artifact quality | 4 | Clean, tested (71%), opinionated, dogfooded |
| Differentiation | 4 | Client-side runnable Go is genuinely rare |
| Positioning clarity | 2 | Feature-led; wedge present but unnamed |
| ICP focus | 2.5 | Messaging wider than the moat |
| Distribution / traction | 1 | ~0 across the board; no published release |
| Docs / onboarding | 3.5 | Good docs; no <60s zero-Go on-ramp |
| Security / operability | 4 | Strong defaults; CSP gap noted |
| Monetization readiness | 1.5 | MIT, infra-light; not a near-term lever |
| **Overall GTM readiness** | **~4/10** | Strong product, near-zero go-to-market |

---

## 11. The one thing to do next

**Name the wedge and prove it once.** Reposition around "runnable docs for Go
libraries," cut a real `v0.x` release, put a 3-second runner demo at the top of the
README, and convert one real Go library's docs to GoMark as a live showcase. Everything
else in §6 compounds off those four moves. The product is ready; the market doesn't
know it exists yet.
