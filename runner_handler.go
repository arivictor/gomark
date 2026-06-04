package gomark

import (
	"crypto/subtle"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/arivictor/gomark/protocol"
)

type Handler struct {
	runner    Executor
	authMode  protocol.AuthMode
	authToken string
}

func NewHandler(authConfig protocol.AuthConfig) (Handler, error) {
	mode := protocol.AuthMode(strings.ToLower(strings.TrimSpace(string(authConfig.Mode))))
	if mode == "" {
		mode = protocol.AuthBearerStatic
	}
	token := strings.TrimSpace(authConfig.BearerToken)

	if mode == protocol.AuthBearerStatic && token == "" {
		return Handler{}, fmt.Errorf("runner bearer auth token is required")
	}
	if mode != protocol.AuthBearerStatic && mode != protocol.AuthNone {
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
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if !h.allowRequest(r) {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	var req protocol.RunRequest
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, MaxSourceBytes+1024)).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, protocol.RunResponse{OK: false, Error: "cannot run", ExitCode: 1})
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
	if h.authMode == protocol.AuthNone {
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

func writeJSON(w http.ResponseWriter, status int, payload protocol.RunResponse) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}
