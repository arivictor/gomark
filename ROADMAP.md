# GoMark
- Improve documentation
- Improve landing page
- Add more examples and guides
- Add more configurability
- Test configurability
- Route override from markdown frontmatter
- TOC config (enable disable, depth)
- Front matter: enable copy markdown, enable open with AI
- Small screen menu improvements
- Make code blocks behave like code editors (tabs, auto indent, etc.)

# Runner (in-browser WebAssembly)
- Run the interpreter in a Web Worker with a watchdog so an infinite loop can be
  cancelled without freezing the reader's tab.
- Surface a small "runs in a Go interpreter" note near runnable blocks.
- Shrink the wasm payload (brotli, trimmed stdlib symbol set).
- Broaden yaegi coverage / clearer messaging for unsupported constructs.
- Better error formatting and reporting.
