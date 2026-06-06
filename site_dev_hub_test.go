package gomark

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestLiveReloadHubSubscribeBroadcastUnsubscribe(t *testing.T) {
	hub := newLiveReloadHub()
	ch := hub.subscribe()

	hub.broadcast()
	select {
	case <-ch:
	case <-time.After(time.Second):
		t.Fatal("expected broadcast to reach subscriber")
	}

	// Coalescing: a second broadcast while one is pending must not block.
	hub.broadcast()
	hub.broadcast() // buffer full -> default branch, no block.

	hub.unsubscribe(ch)
	// After unsubscribe, broadcast must not panic and channel gets nothing new.
	hub.broadcast()
}

func TestLiveReloadHubHandlerStreamsAndStops(t *testing.T) {
	hub := newLiveReloadHub()

	ctx, cancel := context.WithCancel(context.Background())
	req := httptest.NewRequest("GET", liveReloadPath, nil).WithContext(ctx)
	// A synchronizing writer lets the test observe the first event without racing
	// on the ResponseRecorder's buffer (which the handler writes from a goroutine).
	sw := &syncFlushWriter{header: make(http.Header), wrote: make(chan struct{}, 1)}

	done := make(chan error, 1)
	go func() {
		done <- hub.handler(sw, req)
	}()

	// Give the handler a moment to subscribe, then broadcast reloads until it
	// writes the first event (or we time out).
	deadline := time.Now().Add(2 * time.Second)
	got := false
	for time.Now().Before(deadline) {
		hub.broadcast()
		select {
		case <-sw.wrote:
			got = true
		case <-time.After(20 * time.Millisecond):
		}
		if got {
			break
		}
	}

	cancel()
	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("handler returned error: %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("handler did not return after context cancel")
	}

	if !got {
		t.Fatal("expected at least one reload event written")
	}
	if ct := sw.header.Get("Content-Type"); ct != "text/event-stream" {
		t.Fatalf("content-type = %q", ct)
	}
}

// syncFlushWriter is a minimal http.ResponseWriter+Flusher whose Write signals on
// a channel, so a test can observe progress without touching shared buffers.
type syncFlushWriter struct {
	header http.Header
	wrote  chan struct{}
}

func (s *syncFlushWriter) Header() http.Header { return s.header }
func (s *syncFlushWriter) WriteHeader(int)     {}
func (s *syncFlushWriter) Flush()              {}
func (s *syncFlushWriter) Write(p []byte) (int, error) {
	select {
	case s.wrote <- struct{}{}:
	default:
	}
	return len(p), nil
}

// nonFlushWriter is an http.ResponseWriter that does not implement http.Flusher.
type nonFlushWriter struct {
	header http.Header
	code   int
}

func (n *nonFlushWriter) Header() http.Header {
	if n.header == nil {
		n.header = make(http.Header)
	}
	return n.header
}
func (n *nonFlushWriter) Write(p []byte) (int, error) { return len(p), nil }
func (n *nonFlushWriter) WriteHeader(code int)         { n.code = code }

func TestLiveReloadHubHandlerRequiresFlusher(t *testing.T) {
	hub := newLiveReloadHub()
	err := hub.handler(&nonFlushWriter{}, httptest.NewRequest("GET", liveReloadPath, nil))
	httpErr, ok := err.(*HTTPError)
	if !ok || httpErr.Status != http.StatusInternalServerError {
		t.Fatalf("expected streaming-unsupported HTTPError, got %v", err)
	}
}
