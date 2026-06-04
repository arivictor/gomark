package main

import (
	"log"
	"os"

	gm "github.com/arivictor/gomark"
)

func main() {
	secret := os.Getenv("RUNNER_AUTH_TOKEN")

	r := gm.NewRunner(
		gm.WithAuth(gm.AuthBearerStatic, secret),
		gm.WithTimeout(30),
	)
	if err := r.Start(); err != nil {
		log.Fatal(err)
	}
}
