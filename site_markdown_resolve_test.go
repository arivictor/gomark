package gomark

import (
	"errors"
	"path/filepath"
	"strings"
	"testing"
)

func TestResolveContentPath(t *testing.T) {
	dir := "content"

	// Empty slug -> home.md
	got, err := resolveContentPath(dir, "")
	if err != nil {
		t.Fatalf("empty: %v", err)
	}
	if filepath.Base(got) != "home.md" {
		t.Fatalf("expected home.md, got %q", got)
	}

	// Bare slug gets .md appended.
	got, err = resolveContentPath(dir, "guide")
	if err != nil {
		t.Fatalf("guide: %v", err)
	}
	if !strings.HasSuffix(got, filepath.Join("content", "guide.md")) {
		t.Fatalf("got %q", got)
	}

	// Explicit .md is fine.
	if _, err := resolveContentPath(dir, "guide.md"); err != nil {
		t.Fatalf("guide.md: %v", err)
	}

	// Non-markdown extension is rejected.
	if _, err := resolveContentPath(dir, "guide.txt"); !errors.Is(err, ErrInvalidMarkdownPath) {
		t.Fatalf("expected invalid path for .txt, got %v", err)
	}

	// Path traversal is rejected.
	if _, err := resolveContentPath(dir, "../../etc/passwd"); !errors.Is(err, ErrInvalidMarkdownPath) {
		t.Fatalf("expected invalid path for traversal, got %v", err)
	}

	// Absolute path is rejected.
	if _, err := resolveContentPath(dir, "/etc/passwd"); !errors.Is(err, ErrInvalidMarkdownPath) {
		t.Fatalf("expected invalid path for absolute, got %v", err)
	}
}

func TestNewMarkdownServiceDefaults(t *testing.T) {
	s := NewMarkdownService(nil, "")
	if s.contentDir != "content" {
		t.Fatalf("expected default content dir, got %q", s.contentDir)
	}
	if s.renderer == nil {
		t.Fatal("expected default renderer")
	}

	custom := NewMarkdownService(StdlibMarkdownRenderer{RunnerEnabled: true}, "docs/./pages/")
	if custom.contentDir != filepath.Join("docs", "pages") {
		t.Fatalf("expected cleaned content dir, got %q", custom.contentDir)
	}
}

func TestParseBoolMeta(t *testing.T) {
	for _, v := range []string{"1", "true", "yes", "on", "  TRUE  ", `"yes"`} {
		if !parseBoolMeta(v) {
			t.Errorf("expected %q true", v)
		}
	}
	for _, v := range []string{"0", "false", "no", "off", "", "maybe"} {
		if parseBoolMeta(v) {
			t.Errorf("expected %q false", v)
		}
	}
}

func TestSlugify(t *testing.T) {
	cases := map[string]string{
		"Hello World":      "hello-world",
		"  Trim Me  ":      "trim-me",
		"Multiple   Space": "multiple-space",
		"Special!@#Chars":  "special-chars",
		"":                 "section",
		"---":              "section",
	}
	for in, want := range cases {
		if got := slugify(in); got != want {
			t.Errorf("slugify(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestParseHeading(t *testing.T) {
	level, text, ok := parseHeading("### Hello")
	if !ok || level != 3 || text != "Hello" {
		t.Fatalf("got %d %q %v", level, text, ok)
	}
	if _, _, ok := parseHeading("####### too deep"); ok {
		t.Fatal("expected >6 hashes to fail")
	}
	if _, _, ok := parseHeading("#no-space"); ok {
		t.Fatal("expected missing space to fail")
	}
	if _, _, ok := parseHeading("not a heading"); ok {
		t.Fatal("expected non-heading to fail")
	}
}

func TestLoadAndRenderInvalidPath(t *testing.T) {
	s := NewMarkdownService(StdlibMarkdownRenderer{}, "content")
	if _, err := s.LoadAndRender("../escape"); !errors.Is(err, ErrInvalidMarkdownPath) {
		t.Fatalf("expected invalid path error, got %v", err)
	}
}

func TestLoadAndRenderFrontmatterAndH1(t *testing.T) {
	dir := t.TempDir()
	mustWrite(t, filepath.Join(dir, "page.md"), "---\ndescription: A page\n---\n# Title Heading\n\nBody text.")
	s := NewMarkdownService(StdlibMarkdownRenderer{}, dir)
	page, err := s.LoadAndRender("page")
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	// Title taken from the leading H1 when frontmatter has no title.
	if page.Title != "Title Heading" {
		t.Fatalf("title = %q", page.Title)
	}
	if page.Description != "A page" {
		t.Fatalf("description = %q", page.Description)
	}
	if strings.Contains(page.HTML, "Title Heading") {
		t.Fatalf("expected leading H1 to be stripped from body: %s", page.HTML)
	}
}
