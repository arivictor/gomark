package gomark

import (
	"crypto/subtle"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

)

type AuthMode string

const (
	AuthModeBearerStatic AuthMode = "bearer_static"
	AuthModeNone         AuthMode = "none"
)

type AuthConfig struct {
	Mode        AuthMode
	BearerToken string
}

type Handler struct {
	runner    GoRunner
	authMode  AuthMode
	authToken string
}

type runRequest struct {
	Code string `json:"code"`
}

type runResponse struct {
	OK         bool   `json:"ok"`
	Output     string `json:"output,omitempty"`
	Error      string `json:"error,omitempty"`
	ExitCode   int    `json:"exitCode"`
	DurationMS int64  `json:"durationMs"`
}

func NewHandler(authConfig AuthConfig) (Handler, error) {
	mode := AuthMode(strings.ToLower(strings.TrimSpace(string(authConfig.Mode))))
	if mode == "" {
		mode = AuthModeBearerStatic
	}
	token := strings.TrimSpace(authConfig.BearerToken)

	if mode == AuthModeBearerStatic && token == "" {
		return Handler{}, fmt.Errorf("runner bearer auth token is required")
	}
	if mode != AuthModeBearerStatic && mode != AuthModeNone {
		return Handler{}, fmt.Errorf("unsupported runner auth mode %q", mode)
	}

	return Handler{runner: GoRunner{}, authMode: mode, authToken: token}, nil
}

func (h Handler) Register(mux *http.ServeMux) {
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})
	mux.HandleFunc("/run", h.handleRun)
}

func (h Handler) handleRun(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if !h.allowRequest(r) {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	var req runRequest
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, MaxSourceBytes+1024)).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, runResponse{OK: false, Error: "cannot run", ExitCode: 1})
		return
	}

	result := h.runner.Run(r.Context(), req.Code)
	status := http.StatusOK
	if !result.OK {
		status = http.StatusBadRequest
	}
	writeJSON(w, status, runResponse{
		OK:         result.OK,
		Output:     result.Output,
		Error:      result.Error,
		ExitCode:   result.ExitCode,
		DurationMS: result.DurationMS,
	})
}

func (h Handler) allowRequest(r *http.Request) bool {
	if h.authMode == AuthModeNone {
		return true
	}

	authHeader := strings.TrimSpace(r.Header.Get("Authorization"))
	if !strings.HasPrefix(strings.ToLower(authHeader), "bearer ") {
		return false
	}
	token := strings.TrimSpace(authHeader[len("Bearer "):])
	if token == "" {
		return false
	}

	return subtle.ConstantTimeCompare([]byte(token), []byte(h.authToken)) == 1
}

func writeJSON(w http.ResponseWriter, status int, payload runResponse) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}
