package main

import (
	"log"

	"github.com/arivictor/gomark/internal/protocol"
	"github.com/arivictor/gomark/internal/runner"
)

func main() {
	r := runner.NewRunner(
		runner.WithPort("8081"),
		runner.WithAuth(protocol.AuthNone, ""),
		runner.WithTimeout(10),
	)
	if err := r.Start(); err != nil {
		log.Fatal(err)
	}
}
