package gomark

import (
	"strings"
	"testing"
)

func TestStdlibMarkdownRendererRendersGFMTable(t *testing.T) {
	renderer := StdlibMarkdownRenderer{}
	output, _ := renderer.Render("| Name | Count |\n| --- | --- |\n| Apples | 3 |\n| Pears | 12 |")

	for _, want := range []string{"<table>", "<thead>", "<th>Name</th>", "<th>Count</th>", "<tbody>", "<td>Apples</td>", "<td>12</td>"} {
		if !strings.Contains(output, want) {
			t.Fatalf("expected %q in table output: %s", want, output)
		}
	}
}

func TestStdlibMarkdownRendererTableAlignment(t *testing.T) {
	renderer := StdlibMarkdownRenderer{}
	output, _ := renderer.Render("| L | C | R |\n| :--- | :---: | ---: |\n| a | b | c |")

	if !strings.Contains(output, `<th style="text-align:left">L</th>`) {
		t.Fatalf("expected left-aligned header: %s", output)
	}
	if !strings.Contains(output, `<th style="text-align:center">C</th>`) {
		t.Fatalf("expected center-aligned header: %s", output)
	}
	if !strings.Contains(output, `<td style="text-align:right">c</td>`) {
		t.Fatalf("expected right-aligned cell: %s", output)
	}
}

func TestStdlibMarkdownRendererTableCellsRenderInline(t *testing.T) {
	renderer := StdlibMarkdownRenderer{}
	output, _ := renderer.Render("| Field | Note |\n| --- | --- |\n| `id` | **required** |")

	if !strings.Contains(output, "<td><code>id</code></td>") {
		t.Fatalf("expected inline code inside cell: %s", output)
	}
	if !strings.Contains(output, "<strong>required</strong>") {
		t.Fatalf("expected inline emphasis inside cell: %s", output)
	}
}

func TestStdlibMarkdownRendererTableRequiresSeparator(t *testing.T) {
	renderer := StdlibMarkdownRenderer{}
	output, _ := renderer.Render("a | b | c\nnot a separator")

	if strings.Contains(output, "<table>") {
		t.Fatalf("did not expect a table without a delimiter row: %s", output)
	}
	if !strings.Contains(output, "<p>") {
		t.Fatalf("expected a paragraph fallback: %s", output)
	}
}

func TestStdlibMarkdownRendererTablePadsRaggedRows(t *testing.T) {
	renderer := StdlibMarkdownRenderer{}
	output, _ := renderer.Render("| A | B | C |\n| --- | --- | --- |\n| only-one |")

	if strings.Count(output, "<td") != 3 {
		t.Fatalf("expected ragged row padded to 3 cells: %s", output)
	}
}

func TestStdlibMarkdownRendererRendersImages(t *testing.T) {
	renderer := StdlibMarkdownRenderer{}
	output, _ := renderer.Render("![A cat](/cat.png)")

	if !strings.Contains(output, `<img src="/cat.png" alt="A cat" loading="lazy" />`) {
		t.Fatalf("expected rendered image: %s", output)
	}
}

func TestStdlibMarkdownRendererImageEmptyAlt(t *testing.T) {
	renderer := StdlibMarkdownRenderer{}
	output, _ := renderer.Render("![](https://example.com/x.png)")

	if !strings.Contains(output, `<img src="https://example.com/x.png" alt="" loading="lazy" />`) {
		t.Fatalf("expected image with empty alt and passthrough http src: %s", output)
	}
}

func TestStdlibMarkdownRendererImageInsideTableCell(t *testing.T) {
	renderer := StdlibMarkdownRenderer{}
	output, _ := renderer.Render("| Logo |\n| --- |\n| ![alt](/l.png) |")

	if !strings.Contains(output, `<td><img src="/l.png" alt="alt" loading="lazy" /></td>`) {
		t.Fatalf("expected image inside table cell: %s", output)
	}
}

func TestStdlibMarkdownRendererBrokenImageIsLiteral(t *testing.T) {
	renderer := StdlibMarkdownRenderer{}
	output, _ := renderer.Render("![alt](unclosed")

	if strings.Contains(output, "<img") {
		t.Fatalf("did not expect an <img> for a broken image: %s", output)
	}
	if !strings.Contains(output, "![alt](unclosed") {
		t.Fatalf("expected literal text for broken image: %s", output)
	}
}

func TestStdlibMarkdownRendererNestedUnorderedList(t *testing.T) {
	renderer := StdlibMarkdownRenderer{}
	output, _ := renderer.Render("- a\n  - b\n- c")

	if strings.Count(output, "<ul>") != 2 {
		t.Fatalf("expected an outer and a nested <ul>, got: %s", output)
	}
	if !strings.Contains(output, "<li>b</li>") {
		t.Fatalf("expected nested item: %s", output)
	}
	// The nested list must open inside the first <li>, before its </li>.
	firstClose := strings.Index(output, "</li>")
	secondUL := strings.LastIndex(output, "<ul>")
	if secondUL > firstClose {
		t.Fatalf("expected nested <ul> before the first </li>: %s", output)
	}
}

func TestStdlibMarkdownRendererOrderedNestedUnderUnordered(t *testing.T) {
	renderer := StdlibMarkdownRenderer{}
	output, _ := renderer.Render("- a\n  1. one\n  2. two\n- b")

	if !strings.Contains(output, "<ul>") || !strings.Contains(output, "<ol>") {
		t.Fatalf("expected an outer ul with a nested ol: %s", output)
	}
	if !strings.Contains(output, "<li>one</li>") || !strings.Contains(output, "<li>two</li>") {
		t.Fatalf("expected nested ordered items: %s", output)
	}
}

func TestStdlibMarkdownRendererFlatListUnchanged(t *testing.T) {
	renderer := StdlibMarkdownRenderer{}
	output, _ := renderer.Render("- first\n- second")

	want := "<ul>\n  <li>first</li>\n  <li>second</li>\n</ul>\n"
	if output != want {
		t.Fatalf("flat list output changed.\n got: %q\nwant: %q", output, want)
	}
}

func TestStdlibMarkdownRendererRendersCallouts(t *testing.T) {
	kinds := map[string]string{
		"NOTE":      "callout-note",
		"TIP":       "callout-tip",
		"IMPORTANT": "callout-important",
		"WARNING":   "callout-warning",
		"CAUTION":   "callout-caution",
	}
	renderer := StdlibMarkdownRenderer{}
	for marker, class := range kinds {
		output, _ := renderer.Render("> [!" + marker + "]\n> body text")
		if !strings.Contains(output, `<div class="callout `+class+`">`) {
			t.Fatalf("expected %s callout div: %s", marker, output)
		}
		if !strings.Contains(output, `<p class="callout-title">`) {
			t.Fatalf("expected callout title for %s: %s", marker, output)
		}
		if !strings.Contains(output, "body text") {
			t.Fatalf("expected callout body for %s: %s", marker, output)
		}
	}
}

func TestStdlibMarkdownRendererCalloutInlineTitleAndEmphasis(t *testing.T) {
	renderer := StdlibMarkdownRenderer{}
	output, _ := renderer.Render("> [!TIP] Use **bold** here")

	if !strings.Contains(output, `<div class="callout callout-tip">`) {
		t.Fatalf("expected tip callout: %s", output)
	}
	if !strings.Contains(output, "<strong>bold</strong>") {
		t.Fatalf("expected emphasis in callout body: %s", output)
	}
}

func TestStdlibMarkdownRendererUnknownCalloutFallsBackToBlockquote(t *testing.T) {
	renderer := StdlibMarkdownRenderer{}
	output, _ := renderer.Render("> [!FOO]\n> body")

	if strings.Contains(output, "callout") {
		t.Fatalf("did not expect a callout for an unknown kind: %s", output)
	}
	if !strings.Contains(output, "<blockquote>") {
		t.Fatalf("expected blockquote fallback: %s", output)
	}
}
