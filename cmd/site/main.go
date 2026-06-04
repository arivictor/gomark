package main

import (
	"log"
	"os"

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

	// Static export: `go run ./cmd/site dist` (or EXPORT_DIR=dist) writes a
	// self-contained static site instead of serving it.
	if len(os.Args) > 1 {
		if err := s.Export(os.Args[1]); err != nil {
			log.Fatal(err)
		}
		log.Printf("exported static site to %s", os.Args[1])
		return
	}

	if err := s.Start(); err != nil {
		log.Fatal(err)
	}
}
