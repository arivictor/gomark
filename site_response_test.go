package gomark

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestHTTPErrorAndBadRequestErrorMessages(t *testing.T) {
	he := &HTTPError{Status: 404, Message: "not found"}
	if he.Error() != "not found" {
		t.Fatalf("HTTPError.Error = %q", he.Error())
	}
	br := &BadRequestError{Message: "bad input"}
	if br.Error() != "bad input" {
		t.Fatalf("BadRequestError.Error = %q", br.Error())
	}
}

func TestHTMLErrorResponderNilRenderer(t *testing.T) {
	r := HTMLErrorResponder{}
	rec := httptest.NewRecorder()
	r.Handle(rec, httptest.NewRequest("GET", "/", nil), fmt.Errorf("x"))
	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("code = %d", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "internal server error") {
		t.Fatalf("body = %q", rec.Body.String())
	}
}

func newErrorResponder(t *testing.T) HTMLErrorResponder {
	t.Helper()
	renderer, err := NewFileTemplateRenderer("", "")
	if err != nil {
		t.Fatalf("renderer: %v", err)
	}
	return HTMLErrorResponder{
		Renderer: renderer,
		SiteName: "GoMark",
		Logger:   log.New(io.Discard, "", 0),
	}
}

func TestHTMLErrorResponderHTTPError(t *testing.T) {
	r := newErrorResponder(t)
	rec := httptest.NewRecorder()
	r.Handle(rec, httptest.NewRequest("GET", "/missing", nil), &HTTPError{Status: http.StatusNotFound, Message: "page not found"})
	if rec.Code != http.StatusNotFound {
		t.Fatalf("code = %d", rec.Code)
	}
	body := rec.Body.String()
	if !strings.Contains(body, "page not found") {
		t.Fatalf("expected message in body: %q", body)
	}
}

func TestHTMLErrorResponderBadRequest(t *testing.T) {
	r := newErrorResponder(t)
	rec := httptest.NewRecorder()
	r.Handle(rec, httptest.NewRequest("GET", "/x", nil), &BadRequestError{Message: "invalid content path"})
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("code = %d", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "invalid content path") {
		t.Fatalf("expected message in body: %q", rec.Body.String())
	}
}

func TestHTMLErrorResponderGenericServerError(t *testing.T) {
	r := newErrorResponder(t)
	rec := httptest.NewRecorder()
	r.Handle(rec, httptest.NewRequest("GET", "/x", nil), fmt.Errorf("kaboom"))
	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("code = %d", rec.Code)
	}
}

func TestHTMLErrorResponderEmptyHTTPErrorMessageGetsDefault(t *testing.T) {
	r := newErrorResponder(t)
	rec := httptest.NewRecorder()
	r.Handle(rec, httptest.NewRequest("GET", "/x", nil), &HTTPError{Status: http.StatusNotFound, Message: ""})
	if rec.Code != http.StatusNotFound {
		t.Fatalf("code = %d", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "could not be served") {
		t.Fatalf("expected default description: %q", rec.Body.String())
	}
}

// failingRenderer always errors, exercising the fallback path in Handle.
type failingRenderer struct{}

func (failingRenderer) Render(http.ResponseWriter, string, PageData) error { return nil }
func (failingRenderer) RenderStatus(http.ResponseWriter, int, string, PageData) error {
	return fmt.Errorf("render failed")
}

func TestHTMLErrorResponderRenderFailureFallback(t *testing.T) {
	r := HTMLErrorResponder{
		Renderer: failingRenderer{},
		Logger:   log.New(io.Discard, "", 0),
	}
	rec := httptest.NewRecorder()
	r.Handle(rec, httptest.NewRequest("GET", "/x", nil), &HTTPError{Status: http.StatusNotFound, Message: "nope"})
	if rec.Code != http.StatusNotFound {
		t.Fatalf("code = %d", rec.Code)
	}
}

func TestHTMLErrorResponderDefaultsLoggerWhenNil(t *testing.T) {
	renderer, err := NewFileTemplateRenderer("", "")
	if err != nil {
		t.Fatalf("renderer: %v", err)
	}
	// Logger nil -> Handle should set it to log.Default() and not panic.
	r := HTMLErrorResponder{Renderer: renderer, SiteName: "GoMark"}
	rec := httptest.NewRecorder()
	r.Handle(rec, httptest.NewRequest("GET", "/x", nil), &HTTPError{Status: http.StatusNotFound, Message: "nf"})
	if rec.Code != http.StatusNotFound {
		t.Fatalf("code = %d", rec.Code)
	}
}
