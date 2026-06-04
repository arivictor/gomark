package main

import (
	"log"

	"github.com/arivictor/gomark/internal/protocol"
	"github.com/arivictor/gomark/internal/site"
)

func main() {
	s := site.NewSite(
		site.WithSiteAddress(":8080"),
		site.WithSiteContentDir("cmd/site/content"),
		site.WithSiteMode(site.PreRender),
		site.WithSiteRunner("http://localhost:8081", protocol.AuthNone, ""),
	)
	if err := s.Start(); err != nil {
		log.Fatal(err)
	}
}
