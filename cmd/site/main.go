package main

import (
	"log"

	"github.com/arivictor/gomark"
)


func main() {
	site := gomark.NewSite(
		gomark.WithSiteAddress(":8080"),
		gomark.WithSiteContentDir("cmd/site/content"),
		gomark.WithSiteMode(gomark.PreRender),
		gomark.WithSiteRunner("http://localhost:8081", gomark.RunnerAuthNone, ""),
	)
	if err := site.Start(); err != nil {
		log.Fatal(err)
	}
}
