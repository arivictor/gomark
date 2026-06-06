package gomark

import (
	"strings"
	"testing"
)

func TestParseWikiLink(t *testing.T) {
	cases := []struct {
		name      string
		raw       string
		wantHref  string
		wantLabel string
	}{
		{"empty", "", "", ""},
		{"simple", "guide", "/guide", "guide"},
		{"with label", "guide|Read the guide", "/guide", "Read the guide"},
		{"md suffix stripped", "docs/install.md", "/docs/install", "install"},
		{"anchor", "page#section", "/page#section", "page"},
		{"absolute http", "https://example.com/x", "https://example.com/x", "x"},
		{"empty target", "  |label", "", ""},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			href, label := parseWikiLink(tc.raw)
			if href != tc.wantHref {
				t.Errorf("href = %q, want %q", href, tc.wantHref)
			}
			if label != tc.wantLabel {
				t.Errorf("label = %q, want %q", label, tc.wantLabel)
			}
		})
	}
}

func TestDefaultWikiLabel(t *testing.T) {
	cases := map[string]string{
		"guide":              "guide",
		"docs/install.md":    "install",
		"page#section":       "page",
		"my-cool_page":       "my cool page",
		"/a/b/c":             "c",
		"folder/sub-page.md": "sub page",
	}
	for in, want := range cases {
		if got := defaultWikiLabel(in); got != want {
			t.Errorf("defaultWikiLabel(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestIsASCIIPunct(t *testing.T) {
	punct := []byte{'!', '/', ':', '@', '[', '`', '{', '~'}
	for _, b := range punct {
		if !isASCIIPunct(b) {
			t.Errorf("expected %q to be punct", string(b))
		}
	}
	notPunct := []byte{'a', 'Z', '0', '9', ' '}
	for _, b := range notPunct {
		if isASCIIPunct(b) {
			t.Errorf("expected %q not to be punct", string(b))
		}
	}
}

func TestCalloutLabel(t *testing.T) {
	if calloutLabel("") != "" {
		t.Error("empty kind -> empty label")
	}
	if got := calloutLabel("note"); got != "Note" {
		t.Errorf("got %q", got)
	}
	if got := calloutLabel("WARNING"); got != "Warning" {
		t.Errorf("got %q", got)
	}
	if got := calloutLabel("t"); got != "T" {
		t.Errorf("single char got %q", got)
	}
}

func TestStdlibRendererWikiLinks(t *testing.T) {
	renderer := StdlibMarkdownRenderer{}

	out, _ := renderer.Render("See [[getting-started|the guide]] for details.")
	if !strings.Contains(out, `href="/getting-started"`) {
		t.Fatalf("expected wiki link href: %s", out)
	}
	if !strings.Contains(out, "the guide") {
		t.Fatalf("expected wiki link label: %s", out)
	}

	// Unterminated wiki link is escaped, not turned into a link.
	unterminated, _ := renderer.Render("text [[broken")
	if strings.Contains(unterminated, "<a href") {
		t.Fatalf("did not expect a link for unterminated wiki link: %s", unterminated)
	}
}

func TestStdlibRendererCallout(t *testing.T) {
	renderer := StdlibMarkdownRenderer{}
	out, _ := renderer.Render("> [!note] Pay attention here.")
	if !strings.Contains(out, `class="callout callout-note"`) {
		t.Fatalf("expected callout div: %s", out)
	}
	if !strings.Contains(out, "Note") {
		t.Fatalf("expected callout label: %s", out)
	}
	if !strings.Contains(out, "Pay attention here.") {
		t.Fatalf("expected callout body: %s", out)
	}
}

func TestStdlibRendererBackslashEscape(t *testing.T) {
	renderer := StdlibMarkdownRenderer{}
	out, _ := renderer.Render(`Not \*bold\* text.`)
	if strings.Contains(out, "<em>") || strings.Contains(out, "<strong>") {
		t.Fatalf("expected escaped asterisks to suppress emphasis: %s", out)
	}
	if !strings.Contains(out, "*bold*") {
		t.Fatalf("expected literal asterisks: %s", out)
	}
}
