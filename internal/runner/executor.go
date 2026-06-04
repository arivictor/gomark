package runner

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"go/parser"
	"go/token"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/arivictor/gomark/internal/protocol"
)

const (
	MaxSourceBytes = 64 << 10
	MaxOutputBytes = 64 << 10
	RunTimeout     = 2 * time.Second
)

// Executor runs untrusted Go source and reports the result. The Handler depends
// on this interface so it can be tested against a fake executor.
type Executor interface {
	Run(ctx context.Context, code string) protocol.RunResponse
}

// GoExecutor compiles and runs Go source in an isolated temp directory using
// `go run`, enforcing an import allow-list, size limits, and a timeout.
type GoExecutor struct{}

func (GoExecutor) Run(ctx context.Context, code string) protocol.RunResponse {
	start := time.Now()

	if len(strings.TrimSpace(code)) == 0 || len(code) > MaxSourceBytes {
		return failure(start)
	}
	if !usesOnlyAllowedImports(code) {
		return failure(start)
	}

	dir, err := os.MkdirTemp("", "go-runner-*")
	if err != nil {
		return failure(start)
	}
	defer os.RemoveAll(dir)

	mainPath := filepath.Join(dir, "main.go")
	if writeErr := os.WriteFile(mainPath, []byte(code), 0o600); writeErr != nil {
		return failure(start)
	}

	runCtx, cancel := context.WithTimeout(ctx, RunTimeout)
	defer cancel()

	var stdout, stderr cappedBuffer
	stdout.max = MaxOutputBytes
	stderr.max = MaxOutputBytes

	cmd := exec.CommandContext(runCtx, "go", "run", "main.go")
	cmd.Dir = dir
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	cmd.Env = append(os.Environ(),
		"GOMOD=off",
		"GOPROXY=off",
	)

	err = cmd.Run()
	duration := time.Since(start)

	if errors.Is(runCtx.Err(), context.DeadlineExceeded) {
		return failureWithOutput(start, combineOutput(stdout.String(), stderr.String()), 124)
	}

	if err != nil {
		exitCode := 1
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			exitCode = exitErr.ExitCode()
		}
		output := combineOutput(stdout.String(), stderr.String())
		return protocol.RunResponse{
			OK:         false,
			Output:     output,
			Error:      "execution failed",
			ExitCode:   exitCode,
			DurationMS: duration.Milliseconds(),
		}
	}

	return protocol.RunResponse{
		OK:         true,
		Output:     combineOutput(stdout.String(), stderr.String()),
		Error:      "",
		ExitCode:   0,
		DurationMS: duration.Milliseconds(),
	}
}

func failure(start time.Time) protocol.RunResponse {
	return protocol.RunResponse{OK: false, Error: "cannot run", ExitCode: 1, DurationMS: time.Since(start).Milliseconds()}
}

func failureWithOutput(start time.Time, output string, code int) protocol.RunResponse {
	return protocol.RunResponse{OK: false, Output: output, Error: "cannot run", ExitCode: code, DurationMS: time.Since(start).Milliseconds()}
}

func usesOnlyAllowedImports(src string) bool {
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "main.go", src, parser.ImportsOnly)
	if err != nil {
		return false
	}

	for _, imp := range file.Imports {
		path, err := strconv.Unquote(imp.Path.Value)
		if err != nil {
			return false
		}
		if !isAllowedImport(path) {
			return false
		}
	}
	return true
}

func isAllowedImport(path string) bool {
	path = strings.TrimSpace(path)
	if path == "" {
		return false
	}
	if strings.Contains(path, ".") {
		return false
	}
	if strings.HasPrefix(path, "net/") || path == "net" {
		return false
	}

	blocked := map[string]struct{}{
		"os":            {},
		"os/exec":       {},
		"io/ioutil":     {},
		"path/filepath": {},
		"plugin":        {},
		"syscall":       {},
		"unsafe":        {},
	}
	_, banned := blocked[path]
	return !banned
}

func combineOutput(stdout, stderr string) string {
	out := strings.TrimSpace(stdout)
	err := strings.TrimSpace(stderr)
	switch {
	case out == "" && err == "":
		return ""
	case out == "":
		return err
	case err == "":
		return out
	default:
		return out + "\n" + err
	}
}

type cappedBuffer struct {
	buf bytes.Buffer
	max int
}

func (c *cappedBuffer) Write(p []byte) (int, error) {
	if c.max <= 0 {
		return len(p), nil
	}
	remaining := c.max - c.buf.Len()
	if remaining <= 0 {
		return len(p), nil
	}
	if len(p) > remaining {
		_, _ = c.buf.Write(p[:remaining])
		return len(p), nil
	}
	_, err := c.buf.Write(p)
	if err != nil {
		return 0, fmt.Errorf("write output buffer: %w", err)
	}
	return len(p), nil
}

func (c *cappedBuffer) String() string {
	return c.buf.String()
}
