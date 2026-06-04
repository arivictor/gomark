package gomark

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

type roundTripperFunc func(*http.Request) (*http.Response, error)

func (f roundTripperFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

func TestNewRunnerClientRequiresBearerToken(t *testing.T) {
	_, err := NewRunnerClient("http://example.com", AuthConfig{Mode: AuthBearerStatic})
	if err == nil {
		t.Fatalf("expected error when bearer token is missing")
	}
}

func TestRunnerClientRunBearerStaticSetsAuthorization(t *testing.T) {
	var authHeader string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader = r.Header.Get("Authorization")
		_ = json.NewEncoder(w).Encode(RunResponse{OK: true, Output: "ok", ExitCode: 0})
	}))
	defer server.Close()

	client, err := NewRunnerClient(server.URL, AuthConfig{Mode: AuthBearerStatic, BearerToken: "secret-token"})
	if err != nil {
		t.Fatalf("new runner client: %v", err)
	}

	_, err = client.Run(context.Background(), RunRequest{Code: "package gomark"})
	if err != nil {
		t.Fatalf("run runner client: %v", err)
	}

	if authHeader != "Bearer secret-token" {
		t.Fatalf("expected bearer authorization header, got %q", authHeader)
	}
}

func TestRunnerClientRunNoneAuthSkipsAuthorization(t *testing.T) {
	var authHeader string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader = r.Header.Get("Authorization")
		_ = json.NewEncoder(w).Encode(RunResponse{OK: true, Output: "ok", ExitCode: 0})
	}))
	defer server.Close()

	client, err := NewRunnerClient(server.URL, AuthConfig{Mode: AuthNone})
	if err != nil {
		t.Fatalf("new runner client: %v", err)
	}

	_, err = client.Run(context.Background(), RunRequest{Code: "package gomark"})
	if err != nil {
		t.Fatalf("run runner client: %v", err)
	}

	if authHeader != "" {
		t.Fatalf("expected no authorization header, got %q", authHeader)
	}
}

func TestRunnerClientRunRetriesOnHTTPStatusAndSucceeds(t *testing.T) {
	attempts := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		if attempts == 1 {
			w.WriteHeader(http.StatusServiceUnavailable)
			_, _ = w.Write([]byte(`{"ok":false,"error":"warming up"}`))
			return
		}
		_ = json.NewEncoder(w).Encode(RunResponse{OK: true, Output: "ok", ExitCode: 0})
	}))
	defer server.Close()

	client, err := NewRunnerClient(server.URL, AuthConfig{Mode: AuthNone})
	if err != nil {
		t.Fatalf("new runner client: %v", err)
	}

	resp, err := client.Run(context.Background(), RunRequest{Code: "package gomark"})
	if err != nil {
		t.Fatalf("run runner client: %v", err)
	}

	if !resp.OK || resp.Output != "ok" {
		t.Fatalf("expected successful response after retry, got %+v", resp)
	}
	if attempts != 2 {
		t.Fatalf("expected 2 attempts, got %d", attempts)
	}
}

func TestRunnerClientRunRetriesOnTransportErrorAndSucceeds(t *testing.T) {
	attempts := 0
	client, err := NewRunnerClient("http://example.com", AuthConfig{Mode: AuthNone})
	if err != nil {
		t.Fatalf("new runner client: %v", err)
	}

	client.http = &http.Client{Transport: roundTripperFunc(func(req *http.Request) (*http.Response, error) {
		attempts++
		if attempts == 1 {
			return nil, errors.New("dial tcp: connection refused")
		}
		return &http.Response{
			StatusCode: http.StatusOK,
			Header:     make(http.Header),
			Body:       io.NopCloser(strings.NewReader(`{"ok":true,"output":"ok","exitCode":0}`)),
		}, nil
	})}

	resp, runErr := client.Run(context.Background(), RunRequest{Code: "package gomark"})
	if runErr != nil {
		t.Fatalf("run runner client: %v", runErr)
	}
	if !resp.OK || resp.Output != "ok" {
		t.Fatalf("expected successful response after transport retry, got %+v", resp)
	}
	if attempts != 2 {
		t.Fatalf("expected 2 attempts, got %d", attempts)
	}
}

func TestRunnerClientRunNon200IncludesUpstreamMessage(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(RunResponse{OK: false, Error: "execution failed"})
	}))
	defer server.Close()

	client, err := NewRunnerClient(server.URL, AuthConfig{Mode: AuthNone})
	if err != nil {
		t.Fatalf("new runner client: %v", err)
	}

	resp, runErr := client.Run(context.Background(), RunRequest{Code: "package gomark"})
	if runErr != nil {
		t.Fatalf("run runner client: %v", runErr)
	}
	if resp.OK {
		t.Fatalf("expected non-200 upstream to fail")
	}
	if resp.Error != "runner error (400 Bad Request): execution failed" {
		t.Fatalf("expected descriptive upstream error, got %q", resp.Error)
	}
}
