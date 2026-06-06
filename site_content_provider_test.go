package gomark

import (
	"errors"
	"fmt"
	"testing"
)

func TestMapContentProviderError(t *testing.T) {
	if mapContentProviderError(nil) != nil {
		t.Fatal("nil should map to nil")
	}

	notFound := mapContentProviderError(ErrMarkdownNotFound)
	var httpErr *HTTPError
	if !errors.As(notFound, &httpErr) || httpErr.Status != 404 {
		t.Fatalf("expected 404 HTTPError, got %v", notFound)
	}

	invalid := mapContentProviderError(ErrInvalidMarkdownPath)
	var badReq *BadRequestError
	if !errors.As(invalid, &badReq) {
		t.Fatalf("expected BadRequestError, got %v", invalid)
	}

	generic := fmt.Errorf("disk failure")
	if got := mapContentProviderError(generic); got != generic {
		t.Fatalf("expected passthrough, got %v", got)
	}
}

func TestLiveMarkdownProviderGet(t *testing.T) {
	dir := t.TempDir()
	mustWrite(t, dir+"/index.md", "# Home\n\nWelcome")
	p := newLiveMarkdownProvider(dir, StdlibMarkdownRenderer{})
	page, err := p.Get("index")
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if page.Title != "Home" {
		t.Fatalf("title = %q", page.Title)
	}

	if _, err := p.Get("does-not-exist"); !errors.Is(err, ErrMarkdownNotFound) {
		t.Fatalf("expected not found, got %v", err)
	}
}

func TestPreRenderedMarkdownProviderGet(t *testing.T) {
	dir := t.TempDir()
	mustWrite(t, dir+"/index.md", "# Home")
	mustWrite(t, dir+"/about.md", "# About")

	p, err := newPreRenderedMarkdownProvider(dir, StdlibMarkdownRenderer{})
	if err != nil {
		t.Fatalf("build: %v", err)
	}

	if _, err := p.Get("about"); err != nil {
		t.Fatalf("get about: %v", err)
	}
	if _, err := p.Get("missing"); !errors.Is(err, ErrMarkdownNotFound) {
		t.Fatalf("expected not found, got %v", err)
	}
}

func TestPreRenderedMarkdownProviderWalkError(t *testing.T) {
	// A non-existent content dir surfaces a walk error.
	if _, err := newPreRenderedMarkdownProvider(t.TempDir()+"/nope", StdlibMarkdownRenderer{}); err == nil {
		t.Fatal("expected error for missing dir")
	}
}

func TestNewContentPageProviderModes(t *testing.T) {
	dir := t.TempDir()
	mustWrite(t, dir+"/index.md", "# Home")

	live := &App{Mode: LiveRender}
	if _, err := live.newContentPageProvider(dir, StdlibMarkdownRenderer{}); err != nil {
		t.Fatalf("live provider: %v", err)
	}

	pre := &App{Mode: PreRender}
	p, err := pre.newContentPageProvider(dir, StdlibMarkdownRenderer{})
	if err != nil {
		t.Fatalf("pre provider: %v", err)
	}
	if _, ok := p.(preRenderedMarkdownProvider); !ok {
		t.Fatalf("expected preRenderedMarkdownProvider, got %T", p)
	}
}
