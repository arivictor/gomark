//go:build js && wasm

// Command wasm is the client-side Go runner. It compiles to a WebAssembly
// module (GOOS=js GOARCH=wasm) and is loaded by the docs site to execute Go
// snippets entirely in the reader's browser — there is no server-side code
// execution. Source runs through the yaegi interpreter, so it covers a large
// subset of Go (stdlib, generics, goroutines) but is not the full gc toolchain.
//
// It exposes a single global function, runGo(source) -> { output, error },
// which the front-end calls when a reader clicks "Run".
package main

import (
	"syscall/js"

	"github.com/traefik/yaegi/interp"
	"github.com/traefik/yaegi/stdlib"
)

// runGo interprets a Go program and returns its combined stdout/stderr plus an
// optional error string. Each call uses a fresh interpreter so snippets never
// share state. Output is capped to keep a runaway print loop from exhausting
// browser memory.
func runGo(this js.Value, args []js.Value) any {
	result := map[string]any{"output": "", "error": ""}
	if len(args) == 0 {
		result["error"] = "no source provided"
		return js.ValueOf(result)
	}

	src := args[0].String()
	out := &cappedBuffer{max: maxOutputBytes}

	i := interp.New(interp.Options{Stdout: out, Stderr: out})
	if err := i.Use(stdlib.Symbols); err != nil {
		result["error"] = err.Error()
		return js.ValueOf(result)
	}

	// yaegi runs func main automatically when evaluating a `package main`
	// program; a recover guards against interpreter panics so a bad snippet
	// reports an error instead of tearing down the wasm instance.
	func() {
		defer func() {
			if r := recover(); r != nil {
				result["error"] = formatPanic(r)
			}
		}()
		if _, err := i.Eval(src); err != nil {
			result["error"] = err.Error()
		}
	}()

	result["output"] = out.String()
	if out.truncated {
		result["truncated"] = true
	}
	return js.ValueOf(result)
}

func main() {
	js.Global().Set("runGo", js.FuncOf(runGo))
	// Keep the wasm instance alive so runGo stays callable for the page's life.
	select {}
}
