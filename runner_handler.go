package gomark

import (
	"crypto/subtle"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

type Handler struct {
	runner    Executor
	authMode  AuthMode
	authToken string
}

func NewHandler(authConfig AuthConfig) (Handler, error) {
	mode := AuthMode(strings.ToLower(strings.TrimSpace(string(authConfig.Mode))))
	if mode == "" {
		mode = AuthBearerStatic
	}
	token := strings.TrimSpace(authConfig.BearerToken)

	if mode == AuthBearerStatic && token == "" {
		return Handler{}, fmt.Errorf("runner bearer auth token is required")
	}
	if mode != AuthBearerStatic && mode != AuthNone {
		return Handler{}, fmt.Errorf("unsupported runner auth mode %q", mode)
	}

	return Handler{runner: GoExecutor{}, authMode: mode, authToken: token}, nil
}

func (h Handler) Register(mux *http.ServeMux) {
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})
	mux.HandleFunc("/run", h.handleRun)
}

func (h Handler) handleRun(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("Received /run request from %s\n", r.RemoteAddr)
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if !h.allowRequest(r) {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	var req RunRequest
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, MaxSourceBytes+1024)).Decode(&req); err != nil {
		fmt.Printf("Error decoding /run request: %v\n", err)
		writeJSON(w, http.StatusBadRequest, RunResponse{OK: false, Error: err.Error(), ExitCode: 1})
		return
	}

	result := h.runner.Run(r.Context(), req.Code)
	status := http.StatusOK
	if !result.OK {
		status = http.StatusBadRequest
	}
	writeJSON(w, status, result)
}

func (h Handler) allowRequest(r *http.Request) bool {
	if h.authMode == AuthNone {
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

func writeJSON(w http.ResponseWriter, status int, payload RunResponse) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}
