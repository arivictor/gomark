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
	"time"
)

const runnerMaxAttempts = 3

var runnerRetryDelays = []time.Duration{300 * time.Millisecond, 900 * time.Millisecond}

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

	var lastErr error
	for attempt := 1; attempt <= runnerMaxAttempts; attempt++ {
		httpReq, reqErr := http.NewRequestWithContext(ctx, http.MethodPost, c.runnerURL+"/run", bytes.NewReader(payload))
		if reqErr != nil {
			return RunResponse{}, reqErr
		}
		httpReq.Header.Set("Content-Type", "application/json")
		if c.authMode == AuthBearerStatic {
			httpReq.Header.Set("Authorization", "Bearer "+c.authToken)
		}

		resp, doErr := c.http.Do(httpReq)
		if doErr != nil {
			lastErr = doErr
			if attempt == runnerMaxAttempts || !runnerRetryableError(doErr) {
				return RunResponse{}, doErr
			}
			if sleepErr := runnerSleepWithContext(ctx, runnerRetryDelay(attempt)); sleepErr != nil {
				return RunResponse{}, doErr
			}
			continue
		}

		body, readErr := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
		_ = resp.Body.Close()
		if readErr != nil {
			return RunResponse{}, readErr
		}

		if resp.StatusCode == http.StatusOK {
			var decoded RunResponse
			if err := json.Unmarshal(body, &decoded); err != nil {
				return RunResponse{}, err
			}

			if !decoded.OK && strings.TrimSpace(decoded.Error) == "" {
				decoded.Error = "runner returned an unsuccessful response"
			}

			return decoded, nil
		}

		if attempt < runnerMaxAttempts && runnerRetryableStatus(resp.StatusCode) {
			if sleepErr := runnerSleepWithContext(ctx, runnerRetryDelay(attempt)); sleepErr != nil {
				return RunResponse{OK: false, Error: fmt.Sprintf("runner unavailable: %s", resp.Status)}, nil
			}
			continue
		}

		return decodeRunnerError(resp.Status, body), nil
	}

	if lastErr != nil {
		return RunResponse{}, lastErr
	}

	return RunResponse{OK: false, Error: "runner unavailable"}, nil
}

func decodeRunnerError(status string, body []byte) RunResponse {
	var payload RunResponse
	if err := json.Unmarshal(body, &payload); err == nil {
		msg := strings.TrimSpace(payload.Error)
		if msg != "" {
			return RunResponse{OK: false, Error: fmt.Sprintf("runner error (%s): %s", status, msg)}
		}
	}

	return RunResponse{OK: false, Error: fmt.Sprintf("runner returned %s", status)}
}

func runnerRetryableError(err error) bool {
	return err != nil
}

func runnerRetryableStatus(status int) bool {
	switch status {
	case http.StatusTooManyRequests, http.StatusBadGateway, http.StatusServiceUnavailable, http.StatusGatewayTimeout:
		return true
	default:
		return false
	}
}

func runnerRetryDelay(attempt int) time.Duration {
	if attempt <= 0 {
		return 0
	}
	idx := attempt - 1
	if idx >= len(runnerRetryDelays) {
		idx = len(runnerRetryDelays) - 1
	}
	if idx < 0 {
		return 0
	}
	return runnerRetryDelays[idx]
}

func runnerSleepWithContext(ctx context.Context, d time.Duration) error {
	if d <= 0 {
		return nil
	}
	t := time.NewTimer(d)
	defer t.Stop()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-t.C:
		return nil
	}
}
