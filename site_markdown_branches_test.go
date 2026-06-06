package gomark

import (
	"strings"
	"testing"
)

func TestRenderDuplicateHeadingIDs(t *testing.T) {
	out := render("## Setup\n\ntext\n\n## Setup")
	if !strings.Contains(out, `id="setup"`) || !strings.Contains(out, `id="setup-1"`) {
		t.Fatalf("expected de-duplicated heading ids: %s", out)
	}
}

func TestParseFrontmatterNoClosing(t *testing.T) {
	meta, body := parseFrontmatter("---\ntitle: X\nno closing fence")
	if meta != nil {
		t.Fatalf("expected nil meta when frontmatter is unterminated, got %v", meta)
	}
	if !strings.Contains(body, "no closing fence") {
		t.Fatalf("expected original body returned: %q", body)
	}
}

func TestParseFrontmatterSkipsBlankKeys(t *testing.T) {
	meta, _ := parseFrontmatter("---\ntitle: Real\n: orphan value\nnotkeyvalue\n---\nbody")
	if meta["title"] != "Real" {
		t.Fatalf("expected title parsed, got %v", meta)
	}
	if len(meta) != 1 {
		t.Fatalf("expected only the valid key, got %v", meta)
	}
}

func TestRenderTableThenParagraph(t *testing.T) {
	// A table immediately followed by a non-pipe line ends the table body.
	out := render("| A | B |\n| --- | --- |\n| 1 | 2 |\nA following paragraph.")
	if !strings.Contains(out, "</table>") {
		t.Fatalf("expected closed table: %s", out)
	}
	if !strings.Contains(out, "following paragraph") {
		t.Fatalf("expected trailing paragraph: %s", out)
	}
}

func TestRenderQuoteThenParagraph(t *testing.T) {
	out := render("> a quote line\nplain paragraph right after")
	if !strings.Contains(out, "<blockquote>") {
		t.Fatalf("expected blockquote flushed: %s", out)
	}
	if !strings.Contains(out, "plain paragraph") {
		t.Fatalf("expected paragraph: %s", out)
	}
}

func TestRenderListThenParagraph(t *testing.T) {
	out := render("- item one\nplain paragraph right after")
	if !strings.Contains(out, "<ul>") {
		t.Fatalf("expected list flushed: %s", out)
	}
	if !strings.Contains(out, "plain paragraph") {
		t.Fatalf("expected paragraph: %s", out)
	}
}

func TestIndexAutolink(t *testing.T) {
	// Plain http:// (not https) at a boundary.
	if i := indexAutolink("see http://x.test here"); i != 4 {
		t.Fatalf("http index = %d", i)
	}
	// A non-http "://" is skipped, then a real one is found.
	if i := indexAutolink("ws://nope then https://y.test"); i == -1 {
		t.Fatalf("expected to find the https URL after skipping ws://")
	}
	// Inside a token (xhttps) is not a boundary.
	if i := indexAutolink("xhttps://no.test"); i != -1 {
		t.Fatalf("expected no autolink inside a token, got %d", i)
	}
	// No URL at all.
	if i := indexAutolink("just plain words"); i != -1 {
		t.Fatalf("expected -1, got %d", i)
	}
}

func TestRouteFromFrontmatterNonRootedSlug(t *testing.T) {
	// A mailto slug normalizes to a non-rooted target and is rejected.
	if got := routeFromFrontmatter(map[string]string{"slug": "mailto:x@y.com"}); got != "" {
		t.Fatalf("expected empty route for mailto slug, got %q", got)
	}
}

func TestDelimCanCloseUnderscoreAdjacentWord(t *testing.T) {
	// Intra-word underscores do not form emphasis: "_em_word" stays literal.
	out := render("an _em_word here")
	if strings.Contains(out, "<em>") {
		t.Fatalf("did not expect emphasis for intra-word underscore: %s", out)
	}
}

func TestRenderHTTPAutolinkInText(t *testing.T) {
	out := render("plain http://example.org link")
	if !strings.Contains(out, `href="http://example.org"`) {
		t.Fatalf("expected http autolink: %s", out)
	}
}
