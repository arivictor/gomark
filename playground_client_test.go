package gomark

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestNewPlaygroundClientRequiresBearerToken(t *testing.T) {
	_, err := NewPlaygroundClient("http://example.com", PlaygroundAuthConfig{Mode: PlaygroundAuthBearerStatic})
	if err == nil {
		t.Fatalf("expected error when bearer token is missing")
	}
}

func TestPlaygroundClientRunBearerStaticSetsAuthorization(t *testing.T) {
	var authHeader string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader = r.Header.Get("Authorization")
		_ = json.NewEncoder(w).Encode(runnerRunResponse{OK: true, Output: "ok", ExitCode: 0})
	}))
	defer server.Close()

	client, err := NewPlaygroundClient(server.URL, PlaygroundAuthConfig{Mode: PlaygroundAuthBearerStatic, BearerToken: "secret-token"})
	if err != nil {
		t.Fatalf("new playground client: %v", err)
	}

	_, err = client.Run(context.Background(), PlaygroundRunRequest{Code: "package main"})
	if err != nil {
		t.Fatalf("run playground client: %v", err)
	}

	if authHeader != "Bearer secret-token" {
		t.Fatalf("expected bearer authorization header, got %q", authHeader)
	}
}

func TestPlaygroundClientRunNoneAuthSkipsAuthorization(t *testing.T) {
	var authHeader string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader = r.Header.Get("Authorization")
		_ = json.NewEncoder(w).Encode(runnerRunResponse{OK: true, Output: "ok", ExitCode: 0})
	}))
	defer server.Close()

	client, err := NewPlaygroundClient(server.URL, PlaygroundAuthConfig{Mode: PlaygroundAuthNone})
	if err != nil {
		t.Fatalf("new playground client: %v", err)
	}

	_, err = client.Run(context.Background(), PlaygroundRunRequest{Code: "package main"})
	if err != nil {
		t.Fatalf("run playground client: %v", err)
	}

	if authHeader != "" {
		t.Fatalf("expected no authorization header, got %q", authHeader)
	}
}
