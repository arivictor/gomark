package main

import (
	"log"

	"github.com/arivictor/gomark/internal/protocol"
	"github.com/arivictor/gomark/internal/runner"
)

func main() {
	r := runner.NewRunner(
		runner.WithPort("8080"),
		runner.WithAuth(protocol.AuthNone, ""),
	)
	if err := r.Start(); err != nil {
		log.Fatal(err)
	}
}
