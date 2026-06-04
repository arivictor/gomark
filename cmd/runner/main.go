package gomark

import (
	"log"

	"github.com/arivictor/gomark/protocol"
	"github.com/arivictor/gomark/runner"
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
