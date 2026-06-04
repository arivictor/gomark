//go:build js && wasm

package main

import (
	"bytes"
	"fmt"
)

// maxOutputBytes caps how much a snippet may print. The interpreter runs in the
// page's main thread, so an unbounded print loop would otherwise grow the tab's
// memory without limit.
const maxOutputBytes = 256 << 10

// cappedBuffer is an io.Writer that stops collecting after max bytes and records
// that truncation happened. It is not safe for concurrent use, which is fine:
// each runGo call owns its own buffer and the interpreter is single-threaded.
type cappedBuffer struct {
	buf       bytes.Buffer
	max       int
	truncated bool
}

func (c *cappedBuffer) Write(p []byte) (int, error) {
	if c.max <= 0 {
		return len(p), nil
	}
	remaining := c.max - c.buf.Len()
	if remaining <= 0 {
		c.truncated = true
		return len(p), nil
	}
	if len(p) > remaining {
		c.buf.Write(p[:remaining])
		c.truncated = true
		return len(p), nil
	}
	return c.buf.Write(p)
}

func (c *cappedBuffer) String() string {
	s := c.buf.String()
	if c.truncated {
		s += "\n... output truncated"
	}
	return s
}

func formatPanic(r any) string {
	return fmt.Sprintf("panic: %v", r)
}
