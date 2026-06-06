package gomark

import (
	"strings"
	"testing"
)

func TestMinInt(t *testing.T) {
	if minInt(2, 5) != 2 {
		t.Error("minInt(2,5)")
	}
	if minInt(9, 4) != 4 {
		t.Error("minInt(9,4)")
	}
	if minInt(3, 3) != 3 {
		t.Error("minInt(3,3)")
	}
}

func TestIsWholeWord(t *testing.T) {
	if !isWholeWord("learn go today", "go") {
		t.Error("expected whole word match")
	}
	if isWholeWord("golang rocks", "go") {
		t.Error("did not expect partial match in golang")
	}
	if isWholeWord("nothing here", "absent") {
		t.Error("did not expect a match")
	}
	if !isWholeWord("go", "go") {
		t.Error("expected exact match")
	}
	// term appears as substring first, then as whole word later.
	if !isWholeWord("category go", "go") {
		t.Error("expected to find whole word after a partial occurrence")
	}
}

func TestSplitTerms(t *testing.T) {
	if splitTerms("   ") != nil {
		t.Error("expected nil for blank")
	}
	terms := splitTerms("  Hello  World ")
	if len(terms) != 2 || terms[0] != "hello" || terms[1] != "world" {
		t.Errorf("terms = %v", terms)
	}
}

func TestMakeSnippet(t *testing.T) {
	if makeSnippet("", []string{"x"}) != "" {
		t.Error("empty body -> empty snippet")
	}

	// Short body with no term match returns the body as-is.
	if got := makeSnippet("short body", []string{"absent"}); got != "short body" {
		t.Errorf("got %q", got)
	}

	// Long body with no match is truncated with an ellipsis.
	long := strings.Repeat("word ", 60)
	got := makeSnippet(long, []string{"absent"})
	if !strings.HasSuffix(got, "...") {
		t.Errorf("expected ellipsis, got %q", got)
	}

	// Match in the middle is windowed with surrounding ellipses.
	body := strings.Repeat("a ", 60) + "needle " + strings.Repeat("b ", 60)
	snip := makeSnippet(body, []string{"needle"})
	if !strings.Contains(snip, "needle") {
		t.Errorf("expected needle in snippet: %q", snip)
	}
	if !strings.HasPrefix(snip, "...") || !strings.HasSuffix(snip, "...") {
		t.Errorf("expected leading and trailing ellipses: %q", snip)
	}
}

func TestEntriesNilIndex(t *testing.T) {
	var idx *SearchIndex
	if idx.Entries() != nil {
		t.Error("nil index entries should be nil")
	}
	if idx.Query("x", 5) != nil {
		t.Error("nil index query should be nil")
	}
}

func TestQueryLimitAndPhraseBoost(t *testing.T) {
	dir := t.TempDir()
	writeSearchMarkdown(t, dir, "index.md", "---\ntitle: Home\n---\n")
	writeSearchMarkdown(t, dir, "a.md", "---\ntitle: Quick Start Guide\n---\nThe quick start guide helps you begin quickly.")
	writeSearchMarkdown(t, dir, "b.md", "---\ntitle: Other\n---\nA quick mention and a start mention separately.")

	idx, err := BuildSearchIndex(dir)
	if err != nil {
		t.Fatalf("index: %v", err)
	}

	// Phrase appearing in the title should rank that doc first.
	results := idx.Query("quick start", 10)
	if len(results) < 2 {
		t.Fatalf("expected at least 2 results, got %d", len(results))
	}
	if results[0].Path != "/a" {
		t.Fatalf("expected /a first for phrase match, got %q", results[0].Path)
	}

	// Limit is respected.
	limited := idx.Query("quick", 1)
	if len(limited) != 1 {
		t.Fatalf("expected limit of 1, got %d", len(limited))
	}

	// Negative/zero limit defaults to 8.
	def := idx.Query("quick", 0)
	if len(def) == 0 {
		t.Fatal("expected results with default limit")
	}
}

func TestQueryRequiresAllTerms(t *testing.T) {
	dir := t.TempDir()
	writeSearchMarkdown(t, dir, "index.md", "---\ntitle: Home\n---\n")
	writeSearchMarkdown(t, dir, "doc.md", "---\ntitle: Routing\n---\nRouting is about paths.")
	idx, err := BuildSearchIndex(dir)
	if err != nil {
		t.Fatalf("index: %v", err)
	}
	// "routing" present, "nonexistentterm" absent -> no match.
	if res := idx.Query("routing nonexistentterm", 10); len(res) != 0 {
		t.Fatalf("expected no result when a term is missing, got %v", res)
	}
}
