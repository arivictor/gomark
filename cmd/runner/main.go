package main

import (
	"log"

	"github.com/arivictor/gomark/internal/protocol"
	"github.com/arivictor/gomark/internal/runner"
)

func main() {
	r := runner.NewRunner(
		runner.WithAuth(protocol.AuthNone, ""),
		runner.WithTimeout(30),
	)
	if err := r.Start(); err != nil {
		log.Fatal(err)
	}
}
