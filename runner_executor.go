package gomark

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	MaxSourceBytes = 64 << 10
	MaxOutputBytes = 64 << 10
	RunTimeout     = 2 * time.Second
)

// Executor runs untrusted Go source and reports the result. The Handler depends
// on this interface so it can be tested against a fake executor.
type Executor interface {
	Run(ctx context.Context, code string) RunResponse
}

// GoExecutor compiles and runs Go source in an isolated temp directory using
// `go run`, enforcing an import allow-list, size limits, and a timeout.
type GoExecutor struct {
	Timeout time.Duration
}

func (e GoExecutor) timeout() time.Duration {
	if e.Timeout <= 0 {
		return RunTimeout
	}
	return e.Timeout
}

func (e GoExecutor) Run(ctx context.Context, code string) RunResponse {
	start := time.Now()

	if len(strings.TrimSpace(code)) == 0 {
		return failure(start, "source code is empty")
	}
	if len(code) > MaxSourceBytes {
		return failure(start, fmt.Sprintf("source exceeds max size (%d bytes)", MaxSourceBytes))
	}
	if !usesOnlyAllowedImports(code) {
		return failure(start, "source imports include blocked or external packages")
	}
	if reason, ok := validateRunnableSource(code); !ok {
		return failure(start, reason)
	}

	dir, err := os.MkdirTemp("", "go-runner-*")
	if err != nil {
		log.Printf("runner: create temp dir failed: %v", err)
		return failure(start, "internal runner error creating workspace")
	}
	defer os.RemoveAll(dir)

	mainPath := filepath.Join(dir, "main.go")
	if writeErr := os.WriteFile(mainPath, []byte(code), 0o600); writeErr != nil {
		log.Printf("runner: write source failed: %v", writeErr)
		return failure(start, "internal runner error writing source")
	}

	runCtx, cancel := context.WithTimeout(ctx, e.timeout())
	defer cancel()

	var output cappedBuffer
	output.max = MaxOutputBytes

	cmd := exec.CommandContext(runCtx, "go", "run", "main.go")
	cmd.Dir = dir
	cmd.Stdout = &output
	cmd.Stderr = &output
	cmd.Env = append(os.Environ(),
		"GOMOD=off",
		"GOPROXY=off",
	)

	err = cmd.Run()
	duration := time.Since(start)

	if errors.Is(runCtx.Err(), context.DeadlineExceeded) {
		return failureWithOutput(start, output.String(), 124, "execution timed out")
	}

	if err != nil {
		log.Printf("runner: go run failed: %v", err)
		exitCode := 1
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			exitCode = exitErr.ExitCode()
		}
		runOutput := output.String()
		errorMessage := fmt.Sprintf("execution failed (exit code %d)", exitCode)
		if strings.TrimSpace(runOutput) == "" {
			errorMessage = fmt.Sprintf("execution failed: %v", err)
		}
		return RunResponse{
			OK:         false,
			Output:     runOutput,
			Error:      errorMessage,
			ExitCode:   exitCode,
			DurationMS: duration.Milliseconds(),
		}
	}

	return RunResponse{
		OK:         true,
		Output:     output.String(),
		Error:      "",
		ExitCode:   0,
		DurationMS: duration.Milliseconds(),
	}
}

func failure(start time.Time, reason string) RunResponse {
	msg := strings.TrimSpace(reason)
	if msg == "" {
		msg = "runner rejected request"
	}
	return RunResponse{OK: false, Error: msg, ExitCode: 1, DurationMS: time.Since(start).Milliseconds()}
}

func failureWithOutput(start time.Time, output string, code int, reason string) RunResponse {
	msg := strings.TrimSpace(reason)
	if msg == "" {
		msg = "runner rejected request"
	}
	return RunResponse{OK: false, Output: output, Error: msg, ExitCode: code, DurationMS: time.Since(start).Milliseconds()}
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

func validateRunnableSource(src string) (string, bool) {
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "main.go", src, parser.AllErrors)
	if err != nil {
		return "source has syntax errors", false
	}
	if file == nil || file.Name == nil {
		return "source has invalid package declaration", false
	}
	if strings.TrimSpace(file.Name.Name) != "main" {
		return "source must declare package main", false
	}

	for _, decl := range file.Decls {
		fn, ok := decl.(*ast.FuncDecl)
		if !ok || fn.Recv != nil || fn.Name == nil || fn.Name.Name != "main" {
			continue
		}
		if fn.Type == nil {
			continue
		}
		if fn.Type.Params != nil && len(fn.Type.Params.List) > 0 {
			continue
		}
		if fn.Type.Results != nil && len(fn.Type.Results.List) > 0 {
			continue
		}
		return "", true
	}

	return "source must define func main()", false
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
	mu  sync.Mutex
	buf bytes.Buffer
	max int
}

func (c *cappedBuffer) Write(p []byte) (int, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

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
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.buf.String()
}
