package gomark

import (
	"strings"
	"testing"
)

func render(s string) string {
	out, _ := StdlibMarkdownRenderer{}.Render(s)
	return out
}

func TestInlineUnterminatedCode(t *testing.T) {
	out := render("text `unterminated code")
	if strings.Contains(out, "<code>") {
		t.Fatalf("did not expect a <code> for unterminated inline code: %s", out)
	}
	if !strings.Contains(out, "`unterminated code") {
		t.Fatalf("expected literal backtick text: %s", out)
	}
}

func TestInlineImageEmptySrc(t *testing.T) {
	out := render("![alt]()")
	if strings.Contains(out, "<img") {
		t.Fatalf("did not expect img for empty src: %s", out)
	}
}

func TestInlineLinkEmptyHref(t *testing.T) {
	out := render("[label]()")
	if strings.Contains(out, "<a ") {
		t.Fatalf("did not expect anchor for empty href: %s", out)
	}
}

func TestInlineWikiLinkEmptyHref(t *testing.T) {
	// Empty target -> parseWikiLink returns empty href -> literal text retained.
	out := render("[[ |just a label]]")
	if strings.Contains(out, "<a ") {
		t.Fatalf("did not expect anchor for empty wiki target: %s", out)
	}
}

func TestInlineLinkNoClosingBracket(t *testing.T) {
	out := render("[unterminated label")
	if strings.Contains(out, "<a ") {
		t.Fatalf("did not expect anchor: %s", out)
	}
}

func TestInlineBracketsWithoutParen(t *testing.T) {
	out := render("[just brackets] and text")
	if strings.Contains(out, "<a ") {
		t.Fatalf("did not expect anchor when no (url) follows: %s", out)
	}
	if !strings.Contains(out, "[just brackets]") {
		t.Fatalf("expected literal brackets: %s", out)
	}
}

func TestInlineLinkUnmatchedParen(t *testing.T) {
	out := render("[label](/unclosed")
	if strings.Contains(out, "<a ") {
		t.Fatalf("did not expect anchor for unmatched paren: %s", out)
	}
}

func TestTableEscapedPipe(t *testing.T) {
	out := render("| A | B |\n| --- | --- |\n| a \\| b | c |")
	if !strings.Contains(out, "a | b") {
		t.Fatalf("expected escaped pipe inside cell: %s", out)
	}
}

func TestParseTaskListItemEdges(t *testing.T) {
	// Too short.
	if state, _ := parseTaskListItem("[]"); state != 0 {
		t.Errorf("expected no task for '[]'")
	}
	// No space after marker.
	if state, _ := parseTaskListItem("[ ]x"); state != 0 {
		t.Errorf("expected no task when no space follows marker")
	}
	// Unchecked.
	if state, text := parseTaskListItem("[ ] do it"); state != 1 || text != "do it" {
		t.Errorf("got %d %q", state, text)
	}
	// Checked.
	if state, _ := parseTaskListItem("[x] done"); state != 2 {
		t.Errorf("expected checked")
	}
	// Bracket content that is not a marker.
	if state, _ := parseTaskListItem("[y] huh"); state != 0 {
		t.Errorf("expected no task for non-marker char")
	}
}

func TestNormalizeLinkTargetVariants(t *testing.T) {
	cases := map[string]string{
		"mailto:a@b.com":  "mailto:a@b.com",
		"http://x.test":   "http://x.test",
		"#Section-Header": "#section-header",
		"/go/about":       "/go/about", // absolute paths preserved as-is
		"./relative.md":   "/relative",
		"a\\b\\c":         "/a/b/c", // backslashes normalized to slashes
		"/":               "/",
		"  ":              "",
	}
	for in, want := range cases {
		if got := normalizeLinkTarget(in); got != want {
			t.Errorf("normalizeLinkTarget(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestRenderInlineCombinedEmphasis(t *testing.T) {
	out := render("***both*** and ~~strike~~ and _under_")
	if !strings.Contains(out, "<strong>") || !strings.Contains(out, "<em>") {
		t.Fatalf("expected combined emphasis: %s", out)
	}
}
