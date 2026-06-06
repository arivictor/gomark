package gomark

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

// recordingResponder captures the error passed to the server's error responder.
type recordingResponder struct {
	err    error
	status int
}

func (r *recordingResponder) Handle(w http.ResponseWriter, req *http.Request, err error) {
	r.err = err
	status := http.StatusInternalServerError
	var httpErr *HTTPError
	if e, ok := err.(*HTTPError); ok {
		httpErr = e
		status = e.Status
	}
	_ = httpErr
	r.status = status
	http.Error(w, err.Error(), status)
}

func TestServerHandleSuccess(t *testing.T) {
	resp := &recordingResponder{}
	srv := NewServer(resp)
	srv.Handle("GET", "/hello", func(w http.ResponseWriter, r *http.Request) error {
		_, err := w.Write([]byte("hi"))
		return err
	})

	rec := httptest.NewRecorder()
	srv.mux.ServeHTTP(rec, httptest.NewRequest("GET", "/hello", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("code = %d", rec.Code)
	}
	if rec.Body.String() != "hi" {
		t.Fatalf("body = %q", rec.Body.String())
	}
	if resp.err != nil {
		t.Fatalf("unexpected error: %v", resp.err)
	}
}

func TestServerHandleMethodNotAllowed(t *testing.T) {
	resp := &recordingResponder{}
	srv := NewServer(resp)
	srv.Handle("GET", "/only-get", func(w http.ResponseWriter, r *http.Request) error {
		return nil
	})

	rec := httptest.NewRecorder()
	srv.mux.ServeHTTP(rec, httptest.NewRequest("POST", "/only-get", nil))
	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("code = %d", rec.Code)
	}
	var httpErr *HTTPError
	if e, ok := resp.err.(*HTTPError); ok {
		httpErr = e
	}
	if httpErr == nil || httpErr.Status != http.StatusMethodNotAllowed {
		t.Fatalf("expected method not allowed HTTPError, got %v", resp.err)
	}
}

func TestServerHandlePropagatesHandlerError(t *testing.T) {
	resp := &recordingResponder{}
	srv := NewServer(resp)
	wantErr := fmt.Errorf("boom")
	srv.Handle("GET", "/fail", func(w http.ResponseWriter, r *http.Request) error {
		return wantErr
	})

	rec := httptest.NewRecorder()
	srv.mux.ServeHTTP(rec, httptest.NewRequest("GET", "/fail", nil))
	if resp.err != wantErr {
		t.Fatalf("error = %v, want %v", resp.err, wantErr)
	}
}

func TestServerMiddlewareChainOrder(t *testing.T) {
	resp := &recordingResponder{}
	srv := NewServer(resp)
	var order []string
	srv.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			order = append(order, "first")
			next.ServeHTTP(w, r)
		})
	})
	srv.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			order = append(order, "second")
			next.ServeHTTP(w, r)
		})
	})
	srv.Handle("GET", "/m", func(w http.ResponseWriter, r *http.Request) error {
		order = append(order, "handler")
		return nil
	})

	rec := httptest.NewRecorder()
	srv.mux.ServeHTTP(rec, httptest.NewRequest("GET", "/m", nil))
	if len(order) != 3 || order[0] != "first" || order[1] != "second" || order[2] != "handler" {
		t.Fatalf("middleware order = %v", order)
	}
}

func TestServerRunReturnsErrorForInvalidAddr(t *testing.T) {
	srv := NewServer(&recordingResponder{})
	// Missing port -> ListenAndServe returns an error quickly without blocking.
	if err := srv.Run("invalid-address-no-port"); err == nil {
		t.Fatal("expected error for invalid address")
	}
}
