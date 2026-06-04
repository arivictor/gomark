package gomark

import (
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

type Option func(*config)

type Runner struct {
	options []Option
}

type config struct {
	addr      string
	port      string
	authMode  AuthMode
	authToken string
	timeout   time.Duration
}

func NewRunner(options ...Option) *Runner {
	copied := make([]Option, 0, len(options))
	for _, option := range options {
		if option == nil {
			continue
		}
		copied = append(copied, option)
	}
	return &Runner{options: copied}
}

func (r *Runner) Start() error {
	if r == nil {
		return NewRunner().Start()
	}

	cfg := resolveConfig(r.options...)

	h, err := NewHandler(AuthConfig{
		Mode:        cfg.authMode,
		BearerToken: cfg.authToken,
	})
	if err != nil {
		return err
	}
	h.runner = GoExecutor{Timeout: cfg.timeout}

	mux := http.NewServeMux()
	h.Register(mux)

	log.Printf("runner listening on %s", cfg.addr)
	return http.ListenAndServe(cfg.addr, mux)
}

func WithPort(port string) Option {
	return func(c *config) {
		clean := strings.TrimSpace(port)
		if clean == "" {
			return
		}
		c.port = clean
		c.addr = ":" + clean
	}
}

func WithAddress(addr string) Option {
	return func(c *config) {
		clean := strings.TrimSpace(addr)
		if clean == "" {
			return
		}
		c.addr = clean
	}
}

func WithAuth(mode AuthMode, token string) Option {
	return func(c *config) {
		c.authMode = AuthMode(strings.TrimSpace(string(mode)))
		c.authToken = strings.TrimSpace(token)
	}
}

// WithTimeout sets the execution timeout in seconds for each /run request.
// Values <= 0 are ignored.
func WithTimeout(seconds int) Option {
	return func(c *config) {
		if seconds <= 0 {
			return
		}
		c.timeout = time.Duration(seconds) * time.Second
	}
}

func resolveConfig(options ...Option) config {
	port := strings.TrimSpace(os.Getenv("PORT"))
	if port == "" {
		port = "8080"
	}

	addr := strings.TrimSpace(os.Getenv("RUNNER_ADDR"))
	if addr == "" {
		addr = ":" + port
	}

	cfg := config{
		addr:      addr,
		port:      port,
		authMode:  AuthMode(strings.TrimSpace(os.Getenv("RUNNER_AUTH_MODE"))),
		authToken: strings.TrimSpace(os.Getenv("RUNNER_AUTH_TOKEN")),
		timeout:   RunTimeout,
	}

	for _, option := range options {
		if option == nil {
			continue
		}
		option(&cfg)
	}

	if strings.TrimSpace(cfg.addr) == "" {
		cfg.addr = ":" + cfg.port
	}

	return cfg
}
