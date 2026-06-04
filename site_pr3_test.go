package gomark

import (
	"strings"
	"testing"
)

func TestRouteFromFrontmatter(t *testing.T) {
	cases := []struct {
		name string
		meta map[string]string
		want string
	}{
		{"none", map[string]string{"title": "X"}, ""},
		{"slug relative", map[string]string{"slug": "custom-page"}, "/custom-page"},
		{"permalink absolute", map[string]string{"permalink": "/api/v2/users"}, "/api/v2/users"},
		{"route key", map[string]string{"route": "/changelog"}, "/changelog"},
		{"nil", nil, ""},
	}
	for _, tc := range cases {
		if got := routeFromFrontmatter(tc.meta); got != tc.want {
			t.Fatalf("%s: routeFromFrontmatter = %q, want %q", tc.name, got, tc.want)
		}
	}
}

func TestSearchIndexHonorsFrontmatterPermalink(t *testing.T) {
	contentDir := t.TempDir()
	writeSearchMarkdown(t, contentDir, "index.md", "---\ntitle: Home\n---\n")
	writeSearchMarkdown(t, contentDir, "guides/deep/nested.md", "---\ntitle: Nested\npermalink: /changelog\n---\nRelease notes and version history.")

	idx, err := BuildSearchIndex(contentDir)
	if err != nil {
		t.Fatalf("build search index: %v", err)
	}

	results := idx.Query("release notes", 10)
	if len(results) == 0 {
		t.Fatalf("expected a match")
	}
	if results[0].Path != "/changelog" {
		t.Fatalf("expected overridden path /changelog, got %q", results[0].Path)
	}
}

func TestContentIndexHonorsFrontmatterPermalink(t *testing.T) {
	contentDir := t.TempDir()
	writeSearchMarkdown(t, contentDir, "index.md", "---\ntitle: Home\n---\n")
	writeSearchMarkdown(t, contentDir, "guides/deep/nested.md", "---\ntitle: Nested\nslug: /changelog\n---\nbody")

	idx, err := BuildContentIndex(contentDir)
	if err != nil {
		t.Fatalf("build content index: %v", err)
	}

	routes := buildSitemapRoutes(idx)
	found := false
	for _, r := range routes {
		if r == "/changelog" {
			found = true
		}
		if r == "/guides/deep/nested" {
			t.Fatalf("did not expect the path-derived route when a permalink is set: %v", routes)
		}
	}
	if !found {
		t.Fatalf("expected /changelog in sitemap routes, got %v", routes)
	}
}

func TestTOCFrontmatterControls(t *testing.T) {
	if !tocHidden(map[string]string{"toc": "false"}) {
		t.Fatalf("expected toc:false to hide the TOC")
	}
	if tocHidden(map[string]string{"toc": "true"}) {
		t.Fatalf("expected toc:true to keep the TOC")
	}
	if tocHidden(nil) {
		t.Fatalf("expected default (no key) to keep the TOC")
	}
	if got := tocDepth(map[string]string{"toc_depth": "2"}); got != 2 {
		t.Fatalf("expected toc_depth 2, got %d", got)
	}
	if got := tocDepth(nil); got != 3 {
		t.Fatalf("expected default toc_depth 3, got %d", got)
	}

	headings := []Heading{{Level: 2, Text: "A", ID: "a"}, {Level: 3, Text: "B", ID: "b"}}
	limited := limitTOCDepth(headings, 2)
	if len(limited) != 1 || limited[0].Level != 2 {
		t.Fatalf("expected only the H2 after limiting to depth 2, got %+v", limited)
	}
}

func TestSearchRankingTitleBeatsBodyOnly(t *testing.T) {
	contentDir := t.TempDir()
	writeSearchMarkdown(t, contentDir, "index.md", "---\ntitle: Home\n---\n")
	writeSearchMarkdown(t, contentDir, "routing.md", "---\ntitle: Routing\n---\nGeneral overview that mentions caching once.")
	writeSearchMarkdown(t, contentDir, "caching.md", "---\ntitle: Caching\n---\nAll about caching layers and strategies.")

	idx, err := BuildSearchIndex(contentDir)
	if err != nil {
		t.Fatalf("build search index: %v", err)
	}

	results := idx.Query("caching", 10)
	if len(results) < 2 {
		t.Fatalf("expected at least two matches, got %d", len(results))
	}
	if results[0].Path != "/caching" {
		t.Fatalf("expected the title match /caching to rank first, got %q", results[0].Path)
	}
}

func TestStdlibMarkdownRendererAddsHeadingAnchors(t *testing.T) {
	renderer := StdlibMarkdownRenderer{}
	output, _ := renderer.Render("## Configuration\n\ntext")

	if !strings.Contains(output, `<h2 id="configuration">`) {
		t.Fatalf("expected heading id: %s", output)
	}
	if !strings.Contains(output, `<a class="heading-anchor" href="#configuration"`) {
		t.Fatalf("expected heading anchor link: %s", output)
	}
}

func TestStdlibMarkdownRendererTabbedCodeGroup(t *testing.T) {
	renderer := StdlibMarkdownRenderer{}
	output, _ := renderer.Render("```go:title=\"main.go\":group=\"demo\"\npackage main\n```")

	if !strings.Contains(output, `data-tab-group="demo"`) {
		t.Fatalf("expected tab group attribute: %s", output)
	}
	if !strings.Contains(output, `data-tab-title="main.go"`) {
		t.Fatalf("expected tab title attribute: %s", output)
	}
}
