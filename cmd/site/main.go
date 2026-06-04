package main

import (
	"log"

	gm "github.com/arivictor/gomark"
)

func main() {
	// The Go runner executes entirely in the reader's browser (a WebAssembly
	// build of the yaegi interpreter), so it is enabled by default with no
	// service to configure. Disable it with gm.WithSiteRunnerEnabled(false).
	s := gm.NewSite(
		gm.WithSiteAddress(":8080"),
		gm.WithSiteContentDir("cmd/site/content"),
		gm.WithSiteMode(gm.PreRender),
	)

	if err := s.Start(); err != nil {
		log.Fatal(err)
	}
}
