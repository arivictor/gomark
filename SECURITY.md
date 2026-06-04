I'll start by exploring the codebase architecture to understand what we're auditing.

This is a Go documentation site with a code "runner" component. Let me examine the security-critical pieces, especially the runner (which executes code) and the markdown/HTTP handling.

Now the site-side: runner client, CSRF, markdown rendering, routes, server, config, and Dockerfile.

Now the markdown renderer (XSS surface), site wiring, the playground proxy route, and Dockerfile.

Let me check the template (CSP/security headers, CSRF injection) and the security docs that make claims about the model.

I have a thorough picture. Let me confirm a few specifics before writing the audit: security headers, the cgo/`import "C"` gap, and the runner's server timeouts.

I now have a complete picture of the architecture. Here is the audit. Note this is a read-only security review — I haven't changed any code; tell me if you'd like me to implement the hardening.

---

# GoMark Security Audit

**Scope:** Full application architecture — the static-docs **Site** and the Go-code **Runner**, plus their trust boundary, the markdown renderer, CSRF/auth, container packaging, and supply chain. Reviewed at commit `345d275`.

## System model

Two services share one Go module:

- **Site** (`cmd/site`) — renders author-controlled markdown to HTML, serves static assets, exposes `/api/search` and a `/api/runner/run` proxy. Distroless, non-root. Public-facing.
- **Runner** (`cmd/runner`) — accepts arbitrary Go source over HTTP and executes it with `go run`. Bearer-auth'd. Runs in `golang:1.24-alpine` as **root** with the full Go toolchain.

The Site proxies browser "Run" clicks to the Runner, holding the bearer token server-side. **The Runner is the entire security story** — everything else is low-stakes because the rendered content is author-controlled, not user-supplied.

---

## Critical / High

### 1. The Runner sandbox is a source-level deny-list with no OS-level isolation (Critical)
`runner_executor.go` is the only thing standing between an HTTP request and code execution. Its defenses are: an import deny-list, a 64 KB source cap, a 64 KB output cap, and a timeout. There is **no kernel-level containment** — no seccomp, no gVisor/nsjail, no separate locked-down user, no read-only FS, no memory/CPU cgroup, no network namespace. The executed code runs as **root** in the same process namespace as the Runner, in an image that ships the full Go/C toolchain.

This is architecturally backwards: untrusted code execution should be contained at the kernel boundary, with the static checks as defense-in-depth — not the sole control. Concretely exploitable gaps in the deny-list:

- **`import "C"` (cgo) is allowed.** `isAllowedImport("C")` returns `true` (no `.`, not `net/`, not in the blocked map). cgo lets you smuggle `#cgo LDFLAGS`/`CFLAGS` directives to the C toolchain — the classic Go Playground escape. Today it only fails because alpine happens to lack `gcc`; you also never set `CGO_ENABLED=0` for the runtime `go run` (the `CGO_ENABLED=0` in the Dockerfile applies only to *building the binaries*, not to what the Runner executes). This is escape-by-accident, not by design.
- **No memory limit.** `make([]byte, 2<<30)` (allowed — `make` needs no import) balloons the process and can OOM-kill the host. `GOMEMLIMIT` is unset and there's no cgroup.
- **No concurrency or rate limit.** Every `/run` spawns a full compile (`go run`), which is CPU- and RAM-heavy. Nothing caps simultaneous executions, so a flood is a trivial DoS of the host — independent of the per-request timeout.

**Recommendation:** Run executions inside a throwaway, locked-down container/sandbox (gVisor, nsjail, or at minimum a non-root user + read-only rootfs + `--network none` + memory/PID cgroup limits + seccomp). Explicitly block `import "C"` and set `CGO_ENABLED=0` in `cmd.Env`. Add `GOMEMLIMIT`/`ulimit`, a concurrency semaphore, and per-IP rate limiting. Treat the import list as one layer, not the wall.

### 2. The full Go toolchain runs as root in the Runner image (High)
`Dockerfile` builds the Runner into `golang:1.24-alpine` and runs `CMD ["/app/runner"]` as root, with `go`, and a writable filesystem. Any sandbox escape (item 1) lands the attacker as **root with a compiler and writable FS** — maximal blast radius. Contrast the Site, which correctly uses `distroless/static-debian12:nonroot`.

**Recommendation:** The Runner needs `go` at runtime, so it can't be distroless, but it should run as a non-root user, with a read-only root FS, dropped capabilities, and a tmpfs scratch dir.

### 3. The Runner auth secret is exported into the untrusted child process (High)
`runner_executor.go:85`: `cmd.Env = append(os.Environ(), ...)`. This passes the Runner's **entire environment — including `RUNNER_AUTH_TOKEN` — into the `go run` process executing attacker code.** Today `os`/`syscall` are blocked so env-reading is hard, but this directly violates least privilege: the one secret that gates the Runner is handed to the code it's trying to defend against. A single future deny-list bypass (e.g., the cgo path) turns into token exfiltration → unauthenticated RCE for everyone.

**Recommendation:** Build `cmd.Env` from an explicit minimal allow-list (`PATH`, `HOME`, `GOCACHE`, `GOPATH`, `GOMOD=off`, `GOPROXY=off`, `CGO_ENABLED=0`) — never inherit `os.Environ()`.

### 4. `AuthNone` mode = unauthenticated RCE-as-a-service (High, config risk)
Both `NewHandler` and `NewRunnerClient` accept `AuthMode = "none"`. A Runner started with `RUNNER_AUTH_MODE=none` and any network exposure is an open arbitrary-code-execution endpoint. The shipped `cmd/runner/main.go` correctly uses bearer auth, but the mode exists and is reachable via env/config.

**Recommendation:** Gate `AuthNone` behind an explicit, loud flag (e.g. `RUNNER_INSECURE_ALLOW_NO_AUTH=1`) and log a prominent warning at startup. Never let it be the silent default of a misconfiguration.

---

## Medium

### 5. No security headers; third-party scripts with no Subresource Integrity (Medium — supply chain)
`templates/layout.html` loads `highlight.js` and `lucide` from `cdnjs.cloudflare.com` / `unpkg.com` with **no `integrity=` (SRI) hashes**, and the app sets **no CSP, `X-Content-Type-Options`, `X-Frame-Options`/`frame-ancestors`, `Referrer-Policy`, or HSTS** (confirmed: the only headers set anywhere are `Content-Type` and one `Clear-Site-Data`). If either CDN is compromised, arbitrary JS runs on the docs site — and since the same origin can drive the Runner via `/api/runner/run`, a CDN compromise becomes a path to code execution. The missing CSP also means any renderer-escaping slip (item 7) is immediately exploitable rather than mitigated.

**Recommendation:** Add SRI hashes (or self-host the libs), and add a restrictive CSP plus the standard header set via middleware. `X-Content-Type-Options: nosniff` is especially cheap and valuable given the JSON/SVG/HTML you emit.

### 6. Runner HTTP server has no timeouts (Medium)
`runner.go:56` uses bare `http.ListenAndServe` with no `ReadHeaderTimeout`/`ReadTimeout`/`WriteTimeout`/`IdleTimeout` — Slowloris-exposed. The Site server gets this right (`site_server.go:58`, `ReadHeaderTimeout: 5s`); the Runner, the more sensitive service, does not.

**Recommendation:** Wrap the Runner mux in a configured `http.Server` with the same timeout discipline.

### 7. Rendered markdown bypasses Go's template auto-escaping (Medium — defense-in-depth)
`site_routes.go:89` injects the body as `template.HTML(page.HTML)`, disabling auto-escaping and trusting the custom renderer entirely. I reviewed `site_markdown.go` and it does escape correctly — all text goes through `html.EscapeString`, there's no raw-HTML passthrough, and `javascript:`/`data:` URLs are neutralized by `normalizeLinkTarget`'s slugify path. So it's **safe today**, but it's a hand-rolled HTML serializer feeding a trusted sink with no CSP backstop. Any future edit that forgets to escape one branch is instant stored XSS.

**Recommendation:** Keep the renderer, but treat its output as a security-sensitive boundary: add the CSP from item 5 as a backstop, and consider fuzzing `Render` against an HTML sanitizer in CI.

---

## Low / Notes

- **CSRF is actually solid.** `site_csrf.go` combines a same-origin Origin/Referer check with a double-submit token (constant-time compare) and fails closed when both Origin and Referer are absent. Browsers can't forge `Origin`, so Host-header games don't break it. Minor: the CSRF cookie isn't `HttpOnly`/`__Host-` prefixed — fine for double-submit, but tighten if you can.
- **Bearer token in transit.** Both servers use plain `ListenAndServe`. The Site→Runner bearer token is replayable; ensure that hop is TLS or a private network, not plaintext over the public internet.
- **Constant-time comparisons** are used correctly for both the CSRF token and the bearer token (`subtle.ConstantTimeCompare`). 
- **Module download is disabled** (`GOMOD=off`, `GOPROXY=off`) and external import paths are blocked (`strings.Contains(path, ".")`), which correctly prevents pulling arbitrary dependencies — good.
- **Dead code:** `combineOutput` (`runner_executor.go:220`) appears unused.
- **Verbose request logging:** `handleRun` does `fmt.Printf` of `RemoteAddr` and decode errors — low-value, slightly noisy, no secret leakage.

---

## Priority order

1. Contain Runner execution at the kernel level + block `import "C"` / `CGO_ENABLED=0` (#1)
2. Stop inheriting `os.Environ()` into executed code (#3)
3. Non-root, read-only, network-isolated Runner container (#2)
4. Memory/concurrency/rate limits on the Runner (#1)
5. CSP + security headers + SRI (#5)
6. Gate `AuthNone` behind an explicit insecure flag (#4)
7. Runner server timeouts (#6)

**Bottom line:** The web tier is in good shape — sensible CSRF, constant-time auth, an escaping renderer, a distroless non-root Site. The architectural risk is concentrated entirely in the Runner, which executes untrusted code with a brittle source-level deny-list as its *only* real barrier, as root, with the auth secret in its environment. That's the part to redesign around an OS-level sandbox.

Want me to implement the hardening? I'd suggest starting with the highest-leverage, lowest-risk changes — block `import "C"` + `CGO_ENABLED=0`, strip the environment, add the concurrency/memory limits, and add the security-headers/SRI middleware — on `claude/serene-keller-dYvwi`. The container sandboxing (#1/#2) is a larger change I'd scope separately.