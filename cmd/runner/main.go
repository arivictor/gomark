package main

import (
	"log"

	"github.com/arivictor/gomark"
)

func main() {
	runner := gomark.NewRunner(
		gomark.WithPort("8081"),
		gomark.WithAuth(gomark.AuthModeNone, ""),
	)
	if err := runner.Start(); err != nil {
		log.Fatal(err)
	}
}
