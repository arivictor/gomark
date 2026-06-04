package main

import (
	"log"

	"github.com/arivictor/gomark"
)

func main() {
	r := gomark.NewRunner(
		gomark.WithAuth(gomark.AuthNone, ""),
		gomark.WithTimeout(30),
		gomark.WithPort("8081"),
	)
	if err := r.Start(); err != nil {
		log.Fatal(err)
	}
}
