package gomark

import (
	"bytes"
	"context"
	"crypto/subtle"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const defaultPlaygroundRunTimeout = 3 * time.Second

type PlaygroundAuthMode string

const (
	PlaygroundAuthBearerStatic PlaygroundAuthMode = "bearer_static"
	PlaygroundAuthNone         PlaygroundAuthMode = "none"
)

type PlaygroundAuthConfig struct {
	Mode        PlaygroundAuthMode
	BearerToken string
}

type PlaygroundRunRequest struct {
	Code string `json:"code"`
}

type PlaygroundRunResponse struct {
	OK         bool   `json:"ok"`
	Output     string `json:"output,omitempty"`
	Error      string `json:"error,omitempty"`
	ExitCode   int    `json:"exitCode"`
	DurationMS int64  `json:"durationMs"`
}

type runnerRunRequest struct {
	Code string `json:"code"`
}

type runnerRunResponse struct {
	OK         bool   `json:"ok"`
	Output     string `json:"output,omitempty"`
	Error      string `json:"error,omitempty"`
	ExitCode   int    `json:"exitCode"`
	DurationMS int64  `json:"durationMs"`
}

type PlaygroundClient struct {
	runnerURL string
	authMode  PlaygroundAuthMode
	authToken string
	http      *http.Client
}

func NewPlaygroundClient(runnerURL string, authConfig PlaygroundAuthConfig) (*PlaygroundClient, error) {
	cleanURL := strings.TrimRight(strings.TrimSpace(runnerURL), "/")
	if cleanURL == "" {
		return nil, fmt.Errorf("playground runner URL is required")
	}
	if _, err := url.ParseRequestURI(cleanURL); err != nil {
		return nil, fmt.Errorf("invalid playground runner URL: %w", err)
	}

	mode := PlaygroundAuthMode(strings.ToLower(strings.TrimSpace(string(authConfig.Mode))))
	if mode == "" {
		mode = PlaygroundAuthBearerStatic
	}

	token := strings.TrimSpace(authConfig.BearerToken)
	if mode == PlaygroundAuthBearerStatic && token == "" {
		return nil, fmt.Errorf("playground bearer auth token is required")
	}
	if mode != PlaygroundAuthBearerStatic && mode != PlaygroundAuthNone {
		return nil, fmt.Errorf("unsupported playground auth mode %q", mode)
	}

	return &PlaygroundClient{
		runnerURL: cleanURL,
		authMode:  mode,
		authToken: token,
		http:      &http.Client{Timeout: defaultPlaygroundRunTimeout},
	}, nil
}

func (c *PlaygroundClient) Run(ctx context.Context, req PlaygroundRunRequest) (PlaygroundRunResponse, error) {
	if c == nil {
		return PlaygroundRunResponse{}, fmt.Errorf("playground client is not configured")
	}

	payload, err := json.Marshal(runnerRunRequest{Code: req.Code})
	if err != nil {
		return PlaygroundRunResponse{}, err
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, c.runnerURL+"/run", bytes.NewReader(payload))
	if err != nil {
		return PlaygroundRunResponse{}, err
	}
	httpReq.Header.Set("Content-Type", "application/json")
	if c.authMode == PlaygroundAuthBearerStatic {
		httpReq.Header.Set("Authorization", "Bearer "+c.authToken)
	}

	resp, err := c.http.Do(httpReq)
	if err != nil {
		return PlaygroundRunResponse{}, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return PlaygroundRunResponse{}, err
	}

	if resp.StatusCode != http.StatusOK {
		return PlaygroundRunResponse{OK: false, Error: "cannot run"}, nil
	}

	var decoded runnerRunResponse
	if err := json.Unmarshal(body, &decoded); err != nil {
		return PlaygroundRunResponse{}, err
	}

	if !decoded.OK && strings.TrimSpace(decoded.Error) == "" {
		decoded.Error = "cannot run"
	}

	return PlaygroundRunResponse{
		OK:         decoded.OK,
		Output:     decoded.Output,
		Error:      decoded.Error,
		ExitCode:   decoded.ExitCode,
		DurationMS: decoded.DurationMS,
	}, nil
}

func secureTokenMatch(left, right string) bool {
	return subtle.ConstantTimeCompare([]byte(left), []byte(right)) == 1
}
