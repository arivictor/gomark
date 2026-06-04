package main

import (
	"log"
	"os"

	"github.com/arivictor/gomark"
)

func main() {
	runnerURL := os.Getenv("RUNNER_URL")
	if runnerURL == "" {
		runnerURL = "http://localhost:8081"
	}

	s := gomark.NewSite(
		gomark.WithSiteAddress(":8080"),
		gomark.WithSiteContentDir("cmd/site/content"),
		gomark.WithSiteMode(gomark.PreRender),
		gomark.WithSiteRunner(runnerURL, gomark.AuthNone, ""),
	)

	if err := s.Start(); err != nil {
		log.Fatal(err)
	}
}
