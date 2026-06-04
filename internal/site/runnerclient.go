package site

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/arivictor/gomark/internal/protocol"
)

const defaultRunnerRunTimeout = 3 * time.Second

// RunnerClient is the Site's HTTP client to the Runner service.
type RunnerClient struct {
	runnerURL string
	authMode  protocol.AuthMode
	authToken string
	http      *http.Client
}

func NewRunnerClient(runnerURL string, authConfig protocol.AuthConfig) (*RunnerClient, error) {
	cleanURL := strings.TrimRight(strings.TrimSpace(runnerURL), "/")
	if cleanURL == "" {
		return nil, fmt.Errorf("runner URL is required")
	}
	if _, err := url.ParseRequestURI(cleanURL); err != nil {
		return nil, fmt.Errorf("invalid runner URL: %w", err)
	}

	mode := protocol.AuthMode(strings.ToLower(strings.TrimSpace(string(authConfig.Mode))))
	if mode == "" {
		mode = protocol.AuthBearerStatic
	}

	token := strings.TrimSpace(authConfig.BearerToken)
	if mode == protocol.AuthBearerStatic && token == "" {
		return nil, fmt.Errorf("runner bearer auth token is required")
	}
	if mode != protocol.AuthBearerStatic && mode != protocol.AuthNone {
		return nil, fmt.Errorf("unsupported runner auth mode %q", mode)
	}

	return &RunnerClient{
		runnerURL: cleanURL,
		authMode:  mode,
		authToken: token,
		http:      &http.Client{Timeout: defaultRunnerRunTimeout},
	}, nil
}

func (c *RunnerClient) Run(ctx context.Context, req protocol.RunRequest) (protocol.RunResponse, error) {
	if c == nil {
		return protocol.RunResponse{}, fmt.Errorf("runner client is not configured")
	}

	payload, err := json.Marshal(req)
	if err != nil {
		return protocol.RunResponse{}, err
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, c.runnerURL+"/run", bytes.NewReader(payload))
	if err != nil {
		return protocol.RunResponse{}, err
	}
	httpReq.Header.Set("Content-Type", "application/json")
	if c.authMode == protocol.AuthBearerStatic {
		httpReq.Header.Set("Authorization", "Bearer "+c.authToken)
	}

	resp, err := c.http.Do(httpReq)
	if err != nil {
		return protocol.RunResponse{}, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return protocol.RunResponse{}, err
	}

	if resp.StatusCode != http.StatusOK {
		return protocol.RunResponse{OK: false, Error: "cannot run"}, nil
	}

	var decoded protocol.RunResponse
	if err := json.Unmarshal(body, &decoded); err != nil {
		return protocol.RunResponse{}, err
	}

	if !decoded.OK && strings.TrimSpace(decoded.Error) == "" {
		decoded.Error = "cannot run"
	}

	return decoded, nil
}
