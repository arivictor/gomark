package gomark

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestLoggingMiddlewareCallsNext(t *testing.T) {
	called := false
	h := LoggingMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusTeapot)
	}))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest("GET", "/", nil))
	if !called {
		t.Fatal("expected next handler to be called")
	}
	if rec.Code != http.StatusTeapot {
		t.Fatalf("code = %d", rec.Code)
	}
}
