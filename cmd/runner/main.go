package main

import (
	"log"

	"github.com/arivictor/gomark"
)

func main() {
	options := []gomark.Option{
		gomark.WithPort("8080"),
		gomark.WithAuth(gomark.AuthModeNone, ""),
	}
	if err := gomark.Start(options...); err != nil {
		log.Fatal(err)
	}
}
