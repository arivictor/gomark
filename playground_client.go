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

const defaultRunnerRunTimeout = 3 * time.Second

type RunnerAuthMode string

const (
	RunnerAuthBearerStatic RunnerAuthMode = "bearer_static"
	RunnerAuthNone         RunnerAuthMode = "none"
)

type RunnerAuthConfig struct {
	Mode        RunnerAuthMode
	BearerToken string
}

type RunnerRunRequest struct {
	Code string `json:"code"`
}

type RunnerRunResponse struct {
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

type RunnerClient struct {
	runnerURL string
	authMode  RunnerAuthMode
	authToken string
	http      *http.Client
}

func NewRunnerClient(runnerURL string, authConfig RunnerAuthConfig) (*RunnerClient, error) {
	cleanURL := strings.TrimRight(strings.TrimSpace(runnerURL), "/")
	if cleanURL == "" {
		return nil, fmt.Errorf("runner URL is required")
	}
	if _, err := url.ParseRequestURI(cleanURL); err != nil {
		return nil, fmt.Errorf("invalid runner URL: %w", err)
	}

	mode := RunnerAuthMode(strings.ToLower(strings.TrimSpace(string(authConfig.Mode))))
	if mode == "" {
		mode = RunnerAuthBearerStatic
	}

	token := strings.TrimSpace(authConfig.BearerToken)
	if mode == RunnerAuthBearerStatic && token == "" {
		return nil, fmt.Errorf("runner bearer auth token is required")
	}
	if mode != RunnerAuthBearerStatic && mode != RunnerAuthNone {
		return nil, fmt.Errorf("unsupported runner auth mode %q", mode)
	}

	return &RunnerClient{
		runnerURL: cleanURL,
		authMode:  mode,
		authToken: token,
		http:      &http.Client{Timeout: defaultRunnerRunTimeout},
	}, nil
}

func (c *RunnerClient) Run(ctx context.Context, req RunnerRunRequest) (RunnerRunResponse, error) {
	if c == nil {
		return RunnerRunResponse{}, fmt.Errorf("runner client is not configured")
	}

	payload, err := json.Marshal(runnerRunRequest{Code: req.Code})
	if err != nil {
		return RunnerRunResponse{}, err
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, c.runnerURL+"/run", bytes.NewReader(payload))
	if err != nil {
		return RunnerRunResponse{}, err
	}
	httpReq.Header.Set("Content-Type", "application/json")
	if c.authMode == RunnerAuthBearerStatic {
		httpReq.Header.Set("Authorization", "Bearer "+c.authToken)
	}

	resp, err := c.http.Do(httpReq)
	if err != nil {
		return RunnerRunResponse{}, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return RunnerRunResponse{}, err
	}

	if resp.StatusCode != http.StatusOK {
		return RunnerRunResponse{OK: false, Error: "cannot run"}, nil
	}

	var decoded runnerRunResponse
	if err := json.Unmarshal(body, &decoded); err != nil {
		return RunnerRunResponse{}, err
	}

	if !decoded.OK && strings.TrimSpace(decoded.Error) == "" {
		decoded.Error = "cannot run"
	}

	return RunnerRunResponse{
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
