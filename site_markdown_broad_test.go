package gomark

import (
	"strings"
	"testing"
)

func TestStdlibRendererRunnableEditableCode(t *testing.T) {
	renderer := StdlibMarkdownRenderer{RunnerEnabled: true}
	src := "```go:run=true:editable=true:title=\"main.go\"\npackage main\n```"
	out, _ := renderer.Render(src)

	if !strings.Contains(out, `data-runner-run="true"`) {
		t.Fatalf("expected runnable code: %s", out)
	}
	if !strings.Contains(out, "Format") {
		t.Fatalf("expected format button for editable runnable code: %s", out)
	}
	if !strings.Contains(out, "Run") {
		t.Fatalf("expected run button: %s", out)
	}
}

func TestStdlibRendererPlainCodeNoLang(t *testing.T) {
	renderer := StdlibMarkdownRenderer{}
	out, _ := renderer.Render("```\nsome text\n```")
	if !strings.Contains(out, "code-frame-title") {
		t.Fatalf("expected code frame: %s", out)
	}
	// Title falls back to "code" when no language is given.
	if !strings.Contains(out, ">code<") {
		t.Fatalf("expected default 'code' title: %s", out)
	}
}

func TestStdlibRendererTildeFence(t *testing.T) {
	renderer := StdlibMarkdownRenderer{}
	out, _ := renderer.Render("~~~python\nprint(1)\n~~~")
	if !strings.Contains(out, "language-python") {
		t.Fatalf("expected python code class: %s", out)
	}
}

func TestStdlibRendererTaskList(t *testing.T) {
	renderer := StdlibMarkdownRenderer{}
	out, _ := renderer.Render("- [ ] todo\n- [x] done")
	if !strings.Contains(out, "checkbox") && !strings.Contains(out, "checked") {
		t.Fatalf("expected task list checkboxes: %s", out)
	}
}

func TestStdlibRendererHorizontalRule(t *testing.T) {
	renderer := StdlibMarkdownRenderer{}
	out, _ := renderer.Render("text\n\n---\n\nmore")
	if !strings.Contains(out, "<hr />") {
		t.Fatalf("expected hr: %s", out)
	}
}

func TestStdlibRendererHeadingLevels(t *testing.T) {
	renderer := StdlibMarkdownRenderer{}
	out, headings := renderer.Render("# H1\n## H2\n### H3\n#### H4\n##### H5\n###### H6")
	for _, tag := range []string{"<h1", "<h2", "<h3", "<h4", "<h5", "<h6"} {
		if !strings.Contains(out, tag) {
			t.Fatalf("expected %s in output: %s", tag, out)
		}
	}
	// Only H2 and H3 are collected into the TOC.
	if len(headings) != 2 {
		t.Fatalf("expected 2 TOC headings, got %d: %+v", len(headings), headings)
	}
}

func TestStdlibRendererAutolinkAndEmphasis(t *testing.T) {
	renderer := StdlibMarkdownRenderer{}
	out, _ := renderer.Render("Visit https://example.com now. Some *italic*, **bold**, `code`, and a [link](/x).")
	if !strings.Contains(out, `href="https://example.com"`) {
		t.Fatalf("expected autolink: %s", out)
	}
	if !strings.Contains(out, "<em>italic</em>") {
		t.Fatalf("expected italic: %s", out)
	}
	if !strings.Contains(out, "<strong>bold</strong>") {
		t.Fatalf("expected bold: %s", out)
	}
	if !strings.Contains(out, "<code>code</code>") {
		t.Fatalf("expected inline code: %s", out)
	}
	if !strings.Contains(out, `href="/x"`) {
		t.Fatalf("expected link: %s", out)
	}
}

func TestStdlibRendererBlockquotePlain(t *testing.T) {
	renderer := StdlibMarkdownRenderer{}
	out, _ := renderer.Render("> just a quote\n> second line")
	if !strings.Contains(out, "<blockquote><p>") {
		t.Fatalf("expected blockquote: %s", out)
	}
}

func TestStdlibRendererOrderedListResume(t *testing.T) {
	renderer := StdlibMarkdownRenderer{}
	out, _ := renderer.Render("3. three\n4. four")
	if !strings.Contains(out, "<ol") {
		t.Fatalf("expected ordered list: %s", out)
	}
	if !strings.Contains(out, "three") || !strings.Contains(out, "four") {
		t.Fatalf("expected list items: %s", out)
	}
}

func TestStdlibRendererTableEmptyCells(t *testing.T) {
	renderer := StdlibMarkdownRenderer{}
	out, _ := renderer.Render("| A | B |\n| --- | --- |\n|  | y |")
	if strings.Count(out, "<td") != 2 {
		t.Fatalf("expected two cells including an empty one: %s", out)
	}
}

func TestStdlibRendererParagraphAndLineBreaks(t *testing.T) {
	renderer := StdlibMarkdownRenderer{}
	out, _ := renderer.Render("first paragraph\n\nsecond paragraph")
	if strings.Count(out, "<p>") != 2 {
		t.Fatalf("expected two paragraphs: %s", out)
	}
}
