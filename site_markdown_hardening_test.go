package gomark

import (
	"strings"
	"testing"
)

// TestStdlibMarkdownRendererLiteralAsterisksAreNotEmphasis guards the regression
// where "2 * 3 * 4" was rendered as "2 <em> 3 </em> 4": space-flanked delimiters
// must stay literal.
func TestStdlibMarkdownRendererLiteralAsterisksAreNotEmphasis(t *testing.T) {
	output, _ := StdlibMarkdownRenderer{}.Render("Compute 2 * 3 * 4 in your head.")
	if strings.Contains(output, "<em>") {
		t.Fatalf("did not expect emphasis for space-flanked asterisks: %s", output)
	}
	if !strings.Contains(output, "2 * 3 * 4") {
		t.Fatalf("expected literal asterisks preserved: %s", output)
	}
}

func TestStdlibMarkdownRendererUnderscoreEmphasis(t *testing.T) {
	output, _ := StdlibMarkdownRenderer{}.Render("Use _emphasis_ and __strong__ here.")
	if !strings.Contains(output, "<em>emphasis</em>") {
		t.Fatalf("expected underscore italic: %s", output)
	}
	if !strings.Contains(output, "<strong>strong</strong>") {
		t.Fatalf("expected double-underscore bold: %s", output)
	}
}

// TestStdlibMarkdownRendererIntrawordUnderscores ensures identifiers like
// some_var_name (common in Go prose) are never italicized.
func TestStdlibMarkdownRendererIntrawordUnderscores(t *testing.T) {
	output, _ := StdlibMarkdownRenderer{}.Render("Keep some_var_name and a_b_c literal.")
	if strings.Contains(output, "<em>") {
		t.Fatalf("did not expect intraword underscores to emphasize: %s", output)
	}
}

func TestStdlibMarkdownRendererStrikethrough(t *testing.T) {
	output, _ := StdlibMarkdownRenderer{}.Render("This has ~~old~~ text.")
	if !strings.Contains(output, "<del>old</del>") {
		t.Fatalf("expected strikethrough: %s", output)
	}
}

func TestStdlibMarkdownRendererBackslashEscapes(t *testing.T) {
	output, _ := StdlibMarkdownRenderer{}.Render(`A \*literal\* star and 50\% off.`)
	if strings.Contains(output, "<em>") || strings.Contains(output, "<strong>") {
		t.Fatalf("did not expect emphasis from escaped delimiters: %s", output)
	}
	if !strings.Contains(output, "*literal*") {
		t.Fatalf("expected literal asterisks from escapes: %s", output)
	}
	if strings.Contains(output, `\`) {
		t.Fatalf("expected backslashes consumed by escapes: %s", output)
	}
}

func TestStdlibMarkdownRendererTaskList(t *testing.T) {
	output, _ := StdlibMarkdownRenderer{}.Render("- [ ] todo\n- [x] done")
	if !strings.Contains(output, `<li class="task-list-item"><input type="checkbox" disabled> todo</li>`) {
		t.Fatalf("expected unchecked task item: %s", output)
	}
	if !strings.Contains(output, `<li class="task-list-item"><input type="checkbox" disabled checked> done</li>`) {
		t.Fatalf("expected checked task item: %s", output)
	}
}

func TestStdlibMarkdownRendererAutolinksBareURL(t *testing.T) {
	output, _ := StdlibMarkdownRenderer{}.Render("Visit https://example.com now.")
	if !strings.Contains(output, `<a href="https://example.com">https://example.com</a>`) {
		t.Fatalf("expected bare URL autolinked: %s", output)
	}
}

// TestStdlibMarkdownRendererAutolinkTrimsTrailingPunctuation checks GFM-style
// boundary trimming so a trailing ")." is not swallowed into the href.
func TestStdlibMarkdownRendererAutolinkTrimsTrailingPunctuation(t *testing.T) {
	output, _ := StdlibMarkdownRenderer{}.Render("See (https://example.com/foo).")
	if !strings.Contains(output, `<a href="https://example.com/foo">https://example.com/foo</a>`) {
		t.Fatalf("expected trailing punctuation trimmed from autolink: %s", output)
	}
}

func TestStdlibMarkdownRendererAutolinkInsideMarkdownLinkUnaffected(t *testing.T) {
	output, _ := StdlibMarkdownRenderer{}.Render("[home](https://example.com)")
	if strings.Count(output, "<a ") != 1 {
		t.Fatalf("expected a single anchor for an explicit link, got: %s", output)
	}
	if !strings.Contains(output, `>home<`) {
		t.Fatalf("expected explicit link label preserved: %s", output)
	}
}

// TestStdlibMarkdownRendererAdjacentDelimitersStayLiteral ensures empty spans
// like "****" are not emitted as <strong></strong> (CommonMark requires
// non-empty content between delimiters).
func TestStdlibMarkdownRendererAdjacentDelimitersStayLiteral(t *testing.T) {
	for _, in := range []string{"****", "Heading ** ** spaced", "a ~~~~ b"} {
		output, _ := StdlibMarkdownRenderer{}.Render(in)
		if strings.Contains(output, "<strong></strong>") || strings.Contains(output, "<em></em>") || strings.Contains(output, "<del></del>") {
			t.Fatalf("did not expect an empty emphasis span for %q: %s", in, output)
		}
	}
}

// TestStdlibMarkdownRendererOrderedListStartsAtZero checks that a list explicitly
// starting at 0 keeps its numbering via <ol start="0">.
func TestStdlibMarkdownRendererOrderedListStartsAtZero(t *testing.T) {
	output, _ := StdlibMarkdownRenderer{}.Render("0. zero\n1. one")
	if !strings.Contains(output, `<ol start="0">`) {
		t.Fatalf("expected list to carry start=0: %s", output)
	}
}

// TestStdlibMarkdownRendererInterruptedOrderedListResumesNumbering guards the
// regression where an ordered list resumed after a paragraph restarted at 1.
func TestStdlibMarkdownRendererInterruptedOrderedListResumesNumbering(t *testing.T) {
	output, _ := StdlibMarkdownRenderer{}.Render("1. one\n\nA paragraph.\n\n2. two")
	if !strings.Contains(output, `<ol start="2">`) {
		t.Fatalf("expected resumed list to carry start=2: %s", output)
	}
}
