package gomark

import (
	"testing"
	"time"
)

func TestResolveConfigUsesDefaults(t *testing.T) {
	t.Setenv("PORT", "")
	t.Setenv("RUNNER_ADDR", "")
	t.Setenv("RUNNER_AUTH_MODE", "")
	t.Setenv("RUNNER_AUTH_TOKEN", "")

	cfg := resolveConfig()
	if cfg.addr != ":8080" {
		t.Fatalf("expected default addr :8080, got %q", cfg.addr)
	}
	if cfg.authMode != "" {
		t.Fatalf("expected empty auth mode by default, got %q", cfg.authMode)
	}
}

func TestResolveConfigUsesEnvironment(t *testing.T) {
	t.Setenv("PORT", "9090")
	t.Setenv("RUNNER_ADDR", "127.0.0.1:9999")
	t.Setenv("RUNNER_AUTH_MODE", "none")
	t.Setenv("RUNNER_AUTH_TOKEN", "ignored")

	cfg := resolveConfig()
	if cfg.addr != "127.0.0.1:9999" {
		t.Fatalf("expected env addr, got %q", cfg.addr)
	}
	if cfg.authMode != AuthNone {
		t.Fatalf("expected auth mode none, got %q", cfg.authMode)
	}
}

func TestResolveConfigOptionsOverrideEnvironment(t *testing.T) {
	t.Setenv("PORT", "9090")
	t.Setenv("RUNNER_ADDR", "127.0.0.1:9999")
	t.Setenv("RUNNER_AUTH_MODE", "none")
	t.Setenv("RUNNER_AUTH_TOKEN", "env-token")

	cfg := resolveConfig(
		WithPort("7000"),
		WithAuth(AuthBearerStatic, "abc123"),
	)

	if cfg.addr != ":7000" {
		t.Fatalf("expected addr from WithPort, got %q", cfg.addr)
	}
	if cfg.authMode != AuthBearerStatic {
		t.Fatalf("expected bearer static auth mode, got %q", cfg.authMode)
	}
	if cfg.authToken != "abc123" {
		t.Fatalf("expected auth token override, got %q", cfg.authToken)
	}
}

func TestResolveConfigUsesDefaultTimeout(t *testing.T) {
	cfg := resolveConfig()
	if cfg.timeout != RunTimeout {
		t.Fatalf("expected default timeout %s, got %s", RunTimeout, cfg.timeout)
	}
}

func TestResolveConfigWithTimeoutOption(t *testing.T) {
	cfg := resolveConfig(WithTimeout(15))
	if cfg.timeout != 15*time.Second {
		t.Fatalf("expected timeout 15s, got %s", cfg.timeout)
	}
}
