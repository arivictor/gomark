package gomark

import (
	"log"
	"os"

	"github.com/arivictor/gomark/protocol"
	"github.com/arivictor/gomark/site"
)

func main() {
	runnerURL := os.Getenv("RUNNER_URL")
	if runnerURL == "" {
		runnerURL = "http://localhost:8081"
	}
	s := site.NewSite(
		site.WithSiteAddress(":8080"),
		site.WithSiteContentDir("cmd/site/content"),
		site.WithSiteMode(site.PreRender),
		site.WithSiteRunner(runnerURL, protocol.AuthNone, ""),
	)
	if err := s.Start(); err != nil {
		log.Fatal(err)
	}
}
