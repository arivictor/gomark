package gomark

import (
	"net/http/httptest"
	"testing"
)

func TestNewHandlerRequiresTokenForBearerStatic(t *testing.T) {
	_, err := NewHandler(AuthConfig{Mode: AuthModeBearerStatic})
	if err == nil {
		t.Fatalf("expected missing token error")
	}
}

func TestAllowRequestBearerStatic(t *testing.T) {
	h, err := NewHandler(AuthConfig{Mode: AuthModeBearerStatic, BearerToken: "secret"})
	if err != nil {
		t.Fatalf("new handler: %v", err)
	}

	req := httptest.NewRequest("POST", "/run", nil)
	req.Header.Set("Authorization", "Bearer secret")
	if !h.allowRequest(req) {
		t.Fatalf("expected request to be allowed")
	}

	bogus := httptest.NewRequest("POST", "/run", nil)
	bogus.Header.Set("Authorization", "Bearer wrong")
	if h.allowRequest(bogus) {
		t.Fatalf("expected request with wrong token to be denied")
	}
}

func TestAllowRequestNoneMode(t *testing.T) {
	h, err := NewHandler(AuthConfig{Mode: AuthModeNone})
	if err != nil {
		t.Fatalf("new handler: %v", err)
	}

	req := httptest.NewRequest("POST", "/run", nil)
	if !h.allowRequest(req) {
		t.Fatalf("expected none mode to allow request")
	}
}
