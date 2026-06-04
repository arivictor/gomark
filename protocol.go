// Package protocol defines the shared wire contract between the Site (client)
// and the Runner (server). Both sides depend on these types so the request,
// response, and auth shapes have a single source of truth.
package gomark

// RunRequest is the body the Site sends to the Runner's /run endpoint.
type RunRequest struct {
	Code string `json:"code"`
}

// RunResponse is the Runner's reply describing an execution result.
type RunResponse struct {
	OK         bool   `json:"ok"`
	Output     string `json:"output,omitempty"`
	Error      string `json:"error,omitempty"`
	ExitCode   int    `json:"exitCode"`
	DurationMS int64  `json:"durationMs"`
}

// AuthMode selects how the Runner authenticates incoming requests.
type AuthMode string

const (
	AuthBearerStatic AuthMode = "bearer_static"
	AuthNone         AuthMode = "none"
)

// AuthConfig configures authentication for both the Runner server and the
// Site's client to it.
type AuthConfig struct {
	Mode        AuthMode
	BearerToken string
}
