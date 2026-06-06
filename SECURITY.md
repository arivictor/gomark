# Security Policy

## Supported versions

GoMark is pre-1.0. Security fixes are applied to the latest released version.
Please make sure you are on the latest release before reporting an issue.

| Version | Supported |
| ------- | --------- |
| Latest release | Yes |
| Older releases | No |

## Reporting a vulnerability

Please report suspected vulnerabilities privately rather than opening a public
issue. Email **ari@arilaverty.com** with a description of the problem, the
affected version, and steps to reproduce if you have them.

You can expect a best-effort acknowledgement within a few days. We will keep you
informed as we investigate and work on a fix.

## Scope notes

A couple of points worth understanding when assessing GoMark's security surface:

- **The in-browser Go runner executes untrusted code client-side.** Go code
  blocks in docs are interpreted by yaegi compiled to WebAssembly and run
  entirely inside the reader's own browser sandbox. There is no server-side
  execution. The blast radius of any runnable snippet is limited to the reader's
  own browser tab.
- **`gomark serve` is a development tool.** It is intended for local previews
  with live reload, not for production hosting, and should not be exposed
  publicly. Production sites are static HTML/CSS/JS produced by `gomark build`,
  which you serve from any static host.
