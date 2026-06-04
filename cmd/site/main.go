package main

import (
	"log"
	"os"

	gm "github.com/arivictor/gomark"
)

func main() {
	runnerURL := os.Getenv("RUNNER_URL")
	if runnerURL == "" {
		runnerURL = "http://localhost:8081"
	}

	secret := os.Getenv("RUNNER_AUTH_TOKEN")

	s := gm.NewSite(
		gm.WithSiteAddress(":8080"),
		gm.WithSiteContentDir("cmd/site/content"),
		gm.WithSiteMode(gm.PreRender),
		gm.WithSiteRunner(runnerURL, gm.AuthBearerStatic, secret),
	)

	if err := s.Start(); err != nil {
		log.Fatal(err)
	}
}
