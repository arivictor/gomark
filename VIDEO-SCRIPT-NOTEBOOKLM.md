# GoMark — NotebookLM Video Script

_Use this as a source (or as the "customize" prompt) for NotebookLM's Video Overview.
Target length: ~2.5–3 minutes. Tone: confident, developer-to-developer, a little playful.
The whole video should sell ONE idea: **docs for Go libraries where every example actually runs.**_

---

## How to use this in NotebookLM

1. Create a notebook and add your sources: this script, the GoMark `README.md`,
   `GTM-ASSESSMENT.md`, and (optionally) the gomark.dev landing copy.
2. Click **Video Overview → Customize**.
3. Paste the prompt below into the customization box, then generate.

**Customization prompt to paste:**
> Create a ~3-minute explainer video aimed at Go developers who maintain libraries or
> write technical docs. Lead with the single differentiator: GoMark renders markdown
> docs where Go code examples run live in the reader's browser, with nothing executing
> on a server. Follow the beat-by-beat script in the "Video Script" source. Keep the
> tone confident and developer-to-developer. Show, don't list — open on the live-running
> code demo, not a feature bullet list. End on the call to action: `go get
> github.com/arivictor/gomark`. Do not oversell theming or position it as a general
> static-site generator.

---

## Video Script (beat by beat)

### 0:00 — COLD OPEN (the hook)
**On screen:** A docs page with a Go code block. A cursor edits `"Hello, World!"` to
`"Hello, Gophers!"`, clicks **Run**, and the output updates inline — instantly.

**Narration:**
> This is a documentation page. But watch the example. That's not a screenshot — it's
> real Go, running right here in the browser. No playground server. No backend. Just
> docs that actually work. This is GoMark.

---

### 0:20 — THE PROBLEM
**On screen:** Split shot — a static, copy-pasteable code block on one side; a tab open
to the Go Playground on the other; a maintainer wrestling with a Node/Ruby docs toolchain.

**Narration:**
> If you write a Go library, your docs have two bad options. Static code blocks readers
> can't try — so examples drift and break. Or you bolt on a whole separate stack —
> Node, themes, plugins, a playground service to babysit — just to publish some markdown.
> Either way, you're maintaining a website instead of writing docs.

---

### 0:45 — THE SOLUTION
**On screen:** A terminal. `go get github.com/arivictor/gomark`. Then a tiny `main.go`
(the quick-start from the README). Then a `content/` folder of `.md` files. Then the
site appears.

**Narration:**
> GoMark is one Go package. You `go get` it, point it at a folder of markdown, and you
> have a real website. Routing comes from your file tree. Navigation, search, sitemap,
> robots — all generated. It's the docs site you'd have built, minus the building.

---

### 1:10 — THE DIFFERENTIATOR (spend the most time here)
**On screen:** A reader hovers a code block, tweaks a line, hits **Run**, sees output.
Then a callout overlay: "Runs in YOUR browser. Nothing runs on the server."

**Narration:**
> Here's what makes GoMark different. Mark a code block as runnable, and readers can
> edit and execute it — live, on the page. And it all happens in *their* browser, using
> a WebAssembly build of the yaegi Go interpreter. Nothing compiles or runs on your
> servers. There's no execution endpoint to secure, no infrastructure to scale, and the
> classic docs-playground security risk simply doesn't exist. Your examples stay
> honest, because readers can prove they work.

---

### 1:45 — THE "BORING IS GOOD" PILLAR
**On screen:** A single binary deploying to a distroless container; a "static export"
build dropping HTML files; a "works offline" badge.

**Narration:**
> And it's boring to run — in the best way. One binary. A minimal, non-root container.
> Self-hosted assets, so your docs work offline or air-gapped. Need a static site
> instead? Export to plain HTML and host it anywhere. CSRF protection and an escaping
> renderer ship by default. It's the kind of tool you set up once and stop thinking about.

---

### 2:10 — WHO IT'S FOR
**On screen:** Logos/placeholders: "Go library authors," "developer guides,"
"internal handbooks." A live example animating in the corner.

**Narration:**
> GoMark is built for Go developers who'd rather stay in the Go toolchain — library
> maintainers who want examples that run, teams who want docs without a JavaScript
> stack, anyone who believes the best way to explain code is to let people run it.

---

### 2:30 — CALL TO ACTION
**On screen:** Big and clean: `go get github.com/arivictor/gomark` — and the URL
**gomark.dev** (note: "this site is built with GoMark").

**Narration:**
> Docs where every example actually runs. Grab the package, drop in your markdown, and
> ship. It's all at gomark.dev — a site that's built with GoMark itself. `go get
> github.com/arivictor/gomark`. Make your examples runnable.

---

## Quick-reference key messages (for any auto-generated voiceover)
- **One sentence:** "GoMark turns a folder of markdown into a Go-powered docs site where
  code examples run live in the reader's browser."
- **Three pillars:** (1) Runnable by default, fully client-side. (2) One import,
  batteries included. (3) Boring to operate — single binary, distroless, offline-capable.
- **Do say:** runnable Go, in-browser, no server execution, single package, `go get`.
- **Don't say:** "best static-site generator," heavy theming/plugin ecosystem,
  comparisons that position it as a general-purpose Hugo/Docusaurus replacement.
- **CTA:** `go get github.com/arivictor/gomark` · gomark.dev
