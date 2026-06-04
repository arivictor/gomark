package gomark

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

// RunnerClient is the Site's HTTP client to the Runner service.
type RunnerClient struct {
	runnerURL string
	authMode  AuthMode
	authToken string
	http      *http.Client
}

func NewRunnerClient(runnerURL string, authConfig AuthConfig) (*RunnerClient, error) {
	cleanURL := strings.TrimRight(strings.TrimSpace(runnerURL), "/")
	if cleanURL == "" {
		return nil, fmt.Errorf("runner URL is required")
	}
	if _, err := url.ParseRequestURI(cleanURL); err != nil {
		return nil, fmt.Errorf("invalid runner URL: %w", err)
	}

	mode := AuthMode(strings.ToLower(strings.TrimSpace(string(authConfig.Mode))))
	if mode == "" {
		mode = AuthBearerStatic
	}

	token := strings.TrimSpace(authConfig.BearerToken)
	if mode == AuthBearerStatic && token == "" {
		return nil, fmt.Errorf("runner bearer auth token is required")
	}
	if mode != AuthBearerStatic && mode != AuthNone {
		return nil, fmt.Errorf("unsupported runner auth mode %q", mode)
	}

	return &RunnerClient{
		runnerURL: cleanURL,
		authMode:  mode,
		authToken: token,
		// No fixed client timeout here; request lifetime is controlled by context.
		http: &http.Client{},
	}, nil
}

func (c *RunnerClient) Run(ctx context.Context, req RunRequest) (RunResponse, error) {
	if c == nil {
		return RunResponse{}, fmt.Errorf("runner client is not configured")
	}

	payload, err := json.Marshal(req)
	if err != nil {
		return RunResponse{}, err
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, c.runnerURL+"/run", bytes.NewReader(payload))
	if err != nil {
		return RunResponse{}, err
	}
	httpReq.Header.Set("Content-Type", "application/json")
	if c.authMode == AuthBearerStatic {
		httpReq.Header.Set("Authorization", "Bearer "+c.authToken)
	}

	resp, err := c.http.Do(httpReq)
	if err != nil {
		return RunResponse{}, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return RunResponse{}, err
	}

	if resp.StatusCode != http.StatusOK {
		return RunResponse{OK: false, Error: "cannot run"}, nil
	}

	var decoded RunResponse
	if err := json.Unmarshal(body, &decoded); err != nil {
		return RunResponse{}, err
	}

	if !decoded.OK && strings.TrimSpace(decoded.Error) == "" {
		decoded.Error = "cannot run"
	}

	return decoded, nil
}
