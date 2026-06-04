package gomark

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/arivictor/gomark/protocol"
)

func TestNewRunnerClientRequiresBearerToken(t *testing.T) {
	_, err := NewRunnerClient("http://example.com", protocol.AuthConfig{Mode: protocol.AuthBearerStatic})
	if err == nil {
		t.Fatalf("expected error when bearer token is missing")
	}
}

func TestRunnerClientRunBearerStaticSetsAuthorization(t *testing.T) {
	var authHeader string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader = r.Header.Get("Authorization")
		_ = json.NewEncoder(w).Encode(protocol.RunResponse{OK: true, Output: "ok", ExitCode: 0})
	}))
	defer server.Close()

	client, err := NewRunnerClient(server.URL, protocol.AuthConfig{Mode: protocol.AuthBearerStatic, BearerToken: "secret-token"})
	if err != nil {
		t.Fatalf("new runner client: %v", err)
	}

	_, err = client.Run(context.Background(), protocol.RunRequest{Code: "package gomark"})
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
		_ = json.NewEncoder(w).Encode(protocol.RunResponse{OK: true, Output: "ok", ExitCode: 0})
	}))
	defer server.Close()

	client, err := NewRunnerClient(server.URL, protocol.AuthConfig{Mode: protocol.AuthNone})
	if err != nil {
		t.Fatalf("new runner client: %v", err)
	}

	_, err = client.Run(context.Background(), protocol.RunRequest{Code: "package gomark"})
	if err != nil {
		t.Fatalf("run runner client: %v", err)
	}

	if authHeader != "" {
		t.Fatalf("expected no authorization header, got %q", authHeader)
	}
}
